// Package places serves the gm_places table over HTTP: the rows a scraper
// collected about food and drink venues.
package places

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/nguyentanvinh/recommendations-be/backend/internal/platform/handler"
)

// Place is one gm_places row as served to a client. Nullable columns are
// pointers (or json.RawMessage) so "absent" and "zero" stay distinguishable:
// a venue with no rating is not a venue rated 0.
type Place struct {
	ID          int64           `json:"id"`
	PlaceKey    string          `json:"place_key"`
	Name        string          `json:"name"`
	AddressLine *string         `json:"address_line"`
	Address     json.RawMessage `json:"address"`
	District    *string         `json:"district"`
	Province    *string         `json:"province"`
	Latitude    *float64        `json:"latitude"`
	Longitude   *float64        `json:"longitude"`
	PhoneNumber *string         `json:"phone_number"`
	Website     *string         `json:"website"`
	Stars       *float64        `json:"stars"`
	Ratings     *int32          `json:"ratings"`
	PriceLevel  *string         `json:"price_level"`
	Types       []string        `json:"types"`
	OpenHours   json.RawMessage `json:"open_hours"`
	Images      json.RawMessage `json:"images"`
	Category    *string         `json:"category"`
	IsFnB       bool            `json:"is_fnb"`
	FirstSeenAt time.Time       `json:"first_seen_at"`
	LastSeenAt  time.Time       `json:"last_seen_at"`
	LastTaskID  *int64          `json:"last_task_id"`
}

// scanDest returns the Scan targets for placeColumns, in that exact order.
// Keeping the two next to each other is what stops a column being silently
// read into the wrong field when the list changes.
func (p *Place) scanDest() []any {
	return []any{
		&p.ID, &p.PlaceKey, &p.Name, &p.AddressLine, &p.Address, &p.District,
		&p.Province, &p.Latitude, &p.Longitude, &p.PhoneNumber, &p.Website,
		&p.Stars, &p.Ratings, &p.PriceLevel, &p.Types, &p.OpenHours, &p.Images,
		&p.Category, &p.IsFnB, &p.FirstSeenAt, &p.LastSeenAt, &p.LastTaskID,
	}
}

// PlaceInput is the request body of both POST and PUT. PUT is a full
// replacement, so the two carry the same fields: anything omitted is written
// as NULL rather than left alone.
type PlaceInput struct {
	PlaceKey    string          `json:"place_key"`
	Name        string          `json:"name"`
	AddressLine *string         `json:"address_line"`
	Address     json.RawMessage `json:"address"`
	District    *string         `json:"district"`
	Province    *string         `json:"province"`
	Latitude    *float64        `json:"latitude"`
	Longitude   *float64        `json:"longitude"`
	PhoneNumber *string         `json:"phone_number"`
	Website     *string         `json:"website"`
	Stars       *float64        `json:"stars"`
	Ratings     *int32          `json:"ratings"`
	PriceLevel  *string         `json:"price_level"`
	Types       []string        `json:"types"`
	OpenHours   json.RawMessage `json:"open_hours"`
	Images      json.RawMessage `json:"images"`
	Category    *string         `json:"category"`
	// IsFnB defaults to true when omitted, matching the column default.
	IsFnB      *bool  `json:"is_fnb"`
	LastTaskID *int64 `json:"last_task_id"`
}

// Validate rejects values the database would either refuse or accept as
// nonsense. It returns a VALIDATION_ERROR whose message is safe to show.
func (in *PlaceInput) Validate() error {
	switch {
	case strings.TrimSpace(in.PlaceKey) == "":
		return invalid("place_key is required")
	case len(in.PlaceKey) > 512:
		return invalid("place_key must be 512 characters or fewer")
	case strings.TrimSpace(in.Name) == "":
		return invalid("name is required")
	case len(in.Name) > 512:
		return invalid("name must be 512 characters or fewer")
	case in.Latitude != nil && (*in.Latitude < -90 || *in.Latitude > 90):
		return invalid("latitude must be between -90 and 90")
	case in.Longitude != nil && (*in.Longitude < -180 || *in.Longitude > 180):
		return invalid("longitude must be between -180 and 180")
	case in.Stars != nil && (*in.Stars < 0 || *in.Stars > 5):
		return invalid("stars must be between 0 and 5")
	case in.Ratings != nil && *in.Ratings < 0:
		return invalid("ratings must not be negative")
	}
	return nil
}

// args returns the column values of in, in placeWriteColumns order.
func (in *PlaceInput) args() []any {
	isFnB := true
	if in.IsFnB != nil {
		isFnB = *in.IsFnB
	}

	// types is NOT NULL: a missing array is an empty one, not SQL NULL.
	types := in.Types
	if types == nil {
		types = []string{}
	}

	return []any{
		strings.TrimSpace(in.PlaceKey), strings.TrimSpace(in.Name), in.AddressLine,
		jsonbOrNil(in.Address), in.District, in.Province, in.Latitude, in.Longitude,
		in.PhoneNumber, in.Website, in.Stars, in.Ratings, in.PriceLevel, types,
		jsonbOrNil(in.OpenHours), jsonbOrNil(in.Images), in.Category, isFnB,
		in.LastTaskID,
	}
}

// jsonbOrNil maps an omitted JSON field to SQL NULL instead of the JSON
// literal null, so "no opening hours" is absent rather than stored.
func jsonbOrNil(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	return raw
}

func invalid(message string) error {
	return handler.ErrorWithMessage(handler.ECValidationError, message, nil)
}
