package places

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nguyentanvinh/recommendations-be/backend/internal/platform/model"
)

// ErrNotFound is returned when no gm_places row matches the requested id.
var ErrNotFound = errors.New("places: not found")

// placeColumns is the read projection. Its order is the contract that
// Place.scanDest depends on; change one and you must change the other.
const placeColumns = `id, place_key, name, address_line, address, district, province,
	latitude, longitude, phone_number, website, stars, ratings, price_level,
	types, open_hours, images, category, is_fnb, first_seen_at, last_seen_at, last_task_id`

// placeWriteColumns is the client-settable subset, in PlaceInput.args order.
// id and first_seen_at are the database's to assign; last_seen_at is stamped
// by the statement itself.
const placeWriteColumns = `place_key, name, address_line, address, district, province,
	latitude, longitude, phone_number, website, stars, ratings, price_level,
	types, open_hours, images, category, is_fnb, last_task_id`

// Repository is the only place that speaks SQL. It holds no business rules:
// callers hand it validated values and it hands back rows or an error. The
// pool is borrowed, not owned.
type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// ListParams filters and pages a list query. Paging is keyset: Cursor marks
// the last row the client already has, so a row inserted mid-scroll cannot
// shift a page boundary the way an offset would. Limit is always set by
// normalise, so an unbounded table scan cannot be requested.
type ListParams struct {
	Query    string
	Province string
	District string
	Category string

	// CursorParams carries Limit, Cursor and the decoded AfterID. Embedding
	// it is what makes this struct's paging identical to every other
	// entity's.
	model.CursorParams
}

// List returns up to Limit+1 places plus the total number of rows matching
// the filters. The extra row is a lookahead: the service uses its presence to
// decide whether a next page exists, which is cheaper than a second query.
func (r *Repository) List(ctx context.Context, params ListParams) ([]Place, int64, error) {
	var (
		filters = []string{"TRUE"}
		args    []any
	)
	add := func(clause string, value any) {
		args = append(args, value)
		filters = append(filters, fmt.Sprintf(clause, len(args)))
	}

	if params.Query != "" {
		// Escape the LIKE metacharacters so a user searching "100%" does not
		// get a wildcard.
		escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(params.Query)
		add("name ILIKE '%%' || $%d || '%%'", escaped)
	}
	if params.Province != "" {
		add("province = $%d", params.Province)
	}
	if params.District != "" {
		add("district = $%d", params.District)
	}
	if params.Category != "" {
		add("category = $%d", params.Category)
	}

	// The cursor predicate sits outside the CTE on purpose: total must count
	// every row matching the filters, not just the ones after the cursor,
	// otherwise the reported total would shrink with each page.
	cursorClause := ""
	if params.AfterID > 0 {
		args = append(args, params.AfterID)
		cursorClause = fmt.Sprintf("WHERE id > $%d", len(args))
	}

	args = append(args, params.Limit+1)
	query := fmt.Sprintf(`
		WITH filtered AS (
			SELECT %s FROM gm_places WHERE %s
		)
		SELECT %s, (SELECT COUNT(*) FROM filtered) AS total
		FROM filtered
		%s
		ORDER BY id
		LIMIT $%d`,
		placeColumns, strings.Join(filters, " AND "),
		placeColumns, cursorClause, len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("places: list: %w", err)
	}
	defer rows.Close()

	var (
		out   = []Place{}
		total int64
	)
	for rows.Next() {
		var place Place
		if err := rows.Scan(append(place.scanDest(), &total)...); err != nil {
			return nil, 0, fmt.Errorf("places: list scan: %w", err)
		}
		out = append(out, place)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("places: list rows: %w", err)
	}

	return out, total, nil
}

func (r *Repository) Get(ctx context.Context, id int64) (*Place, error) {
	query := fmt.Sprintf("SELECT %s FROM gm_places WHERE id = $1", placeColumns)

	var place Place
	if err := r.pool.QueryRow(ctx, query, id).Scan(place.scanDest()...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("places: get %d: %w", id, err)
	}
	return &place, nil
}

func (r *Repository) Create(ctx context.Context, in *PlaceInput) (*Place, error) {
	args := in.args()
	query := fmt.Sprintf(
		"INSERT INTO gm_places (%s) VALUES (%s) RETURNING %s",
		placeWriteColumns, placeholders(len(args)), placeColumns)

	var place Place
	if err := r.pool.QueryRow(ctx, query, args...).Scan(place.scanDest()...); err != nil {
		return nil, fmt.Errorf("places: create: %w", err)
	}
	return &place, nil
}

// Update replaces every client-settable column. last_seen_at is refreshed
// because a write is, by definition, the most recent sighting of the row.
func (r *Repository) Update(ctx context.Context, id int64, in *PlaceInput) (*Place, error) {
	args := in.args()
	query := fmt.Sprintf(
		"UPDATE gm_places SET %s, last_seen_at = NOW() WHERE id = $%d RETURNING %s",
		assignments(placeWriteColumns), len(args)+1, placeColumns)
	args = append(args, id)

	var place Place
	if err := r.pool.QueryRow(ctx, query, args...).Scan(place.scanDest()...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("places: update %d: %w", id, err)
	}
	return &place, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, "DELETE FROM gm_places WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("places: delete %d: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// placeholders renders "$1, $2, ... $n".
func placeholders(n int) string {
	parts := make([]string, n)
	for i := range parts {
		parts[i] = fmt.Sprintf("$%d", i+1)
	}
	return strings.Join(parts, ", ")
}

// assignments renders a column list as "col = $1, col = $2, ...", keeping the
// UPDATE parameter order identical to the INSERT's.
func assignments(columns string) string {
	parts := strings.Split(columns, ",")
	for i, column := range parts {
		parts[i] = fmt.Sprintf("%s = $%d", strings.TrimSpace(column), i+1)
	}
	return strings.Join(parts, ", ")
}
