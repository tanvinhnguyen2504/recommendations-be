package places

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nguyentanvinh/recommendations-be/backend/internal/platform/config"
	"github.com/nguyentanvinh/recommendations-be/backend/internal/platform/handler"
	"github.com/nguyentanvinh/recommendations-be/backend/internal/platform/model"
)

type Handler struct {
	*handler.Handler
	service *Service
}

func NewHandler(base *handler.Handler, service *Service) *Handler {
	return &Handler{Handler: base, service: service}
}

// PageMetadata is the paging block of a list response. NextCursor is null on
// the last page; a frontend keeps scrolling while it is non-null and feeds it
// back as ?cursor=.
type PageMetadata struct {
	Limit      int     `json:"limit"`
	NextCursor *string `json:"next_cursor"`
}

// ListPlaces godoc
//
//	@Summary		List places
//	@Description	Keyset paged. Omit cursor for the first page, then send back metadata.next_cursor until it is null.
//	@Tags			places
//	@Produce		json
//	@Param			q			query		string	false	"Case-insensitive match on name"
//	@Param			province	query		string	false	"Exact province"
//	@Param			district	query		string	false	"Exact district"
//	@Param			category	query		string	false	"food or drink"
//	@Param			limit		query		int		false	"Page size (1-200, default 50)"
//	@Param			cursor		query		string	false	"metadata.next_cursor from the previous page"
//	@Success		200			{object}	config.SuccessResponse
//	@Failure		400			{object}	handler.ErrorResponse
//	@Router			/places [get]
func (h *Handler) ListPlaces(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	// offset was removed in favour of the cursor. Rejecting it loudly beats
	// ignoring it and serving page one to a client that asked for page three.
	if query.Has("offset") {
		h.ReturnWithError(w, r, invalid("offset is no longer supported; page with cursor instead"))
		return
	}

	limit, err := intParam(query.Get("limit"))
	if err != nil {
		h.ReturnWithError(w, r, invalid("limit must be an integer"))
		return
	}

	params := ListParams{
		Query:    query.Get("q"),
		Province: query.Get("province"),
		District: query.Get("district"),
		Category: query.Get("category"),
		CursorParams: model.CursorParams{
			Limit:  limit,
			Cursor: query.Get("cursor"),
		},
	}

	page, err := h.service.List(r.Context(), &params)
	if err != nil {
		h.ReturnWithError(w, r, err)
		return
	}

	h.ReturnWithJSON(w, http.StatusOK, config.SuccessResponse{
		Count: &page.Total,
		Data:  page.Items,
		Metadata: PageMetadata{
			// params was passed by pointer, so the service's defaults are
			// visible here and this reports the page actually served.
			Limit:      params.Limit,
			NextCursor: page.NextCursor,
		},
	})
}

// GetPlaceByID godoc
//
//	@Summary	Get one place
//	@Tags		places
//	@Produce	json
//	@Param		id	path		int	true	"Place id"
//	@Success	200	{object}	config.SuccessResponse
//	@Failure	404	{object}	handler.ErrorResponse
//	@Router		/places/{id} [get]
func (h *Handler) GetPlaceByID(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		h.ReturnWithError(w, r, err)
		return
	}

	place, err := h.service.Get(r.Context(), id)
	if err != nil {
		h.ReturnWithError(w, r, err)
		return
	}

	h.ReturnWithJSON(w, http.StatusOK, config.SuccessResponse{Data: place})
}

// CreatePlace godoc
//
//	@Summary	Create a place
//	@Tags		places
//	@Accept		json
//	@Produce	json
//	@Param		body	body		PlaceInput	true	"Place"
//	@Success	201		{object}	config.SuccessResponse
//	@Failure	400		{object}	handler.ErrorResponse
//	@Router		/places [post]
func (h *Handler) CreatePlace(w http.ResponseWriter, r *http.Request) {
	in, err := h.decodeInput(w, r)
	if err != nil {
		h.ReturnWithError(w, r, err)
		return
	}

	place, err := h.service.Create(r.Context(), in)
	if err != nil {
		h.ReturnWithError(w, r, err)
		return
	}

	w.Header().Set("Location", r.URL.Path+"/"+strconv.FormatInt(place.ID, 10))
	h.ReturnWithJSON(w, http.StatusCreated, config.SuccessResponse{Data: place})
}

// UpdatePlace godoc
//
//	@Summary		Replace a place
//	@Description	PUT is a full replacement: any field left out of the body is written as NULL.
//	@Tags			places
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int			true	"Place id"
//	@Param			body	body		PlaceInput	true	"Place"
//	@Success		200		{object}	config.SuccessResponse
//	@Failure		400		{object}	handler.ErrorResponse
//	@Failure		404		{object}	handler.ErrorResponse
//	@Router			/places/{id} [put]
func (h *Handler) UpdatePlace(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		h.ReturnWithError(w, r, err)
		return
	}

	in, err := h.decodeInput(w, r)
	if err != nil {
		h.ReturnWithError(w, r, err)
		return
	}

	place, err := h.service.Update(r.Context(), id, in)
	if err != nil {
		h.ReturnWithError(w, r, err)
		return
	}

	h.ReturnWithJSON(w, http.StatusOK, config.SuccessResponse{Data: place})
}

// DeletePlace godoc
//
//	@Summary	Delete a place
//	@Tags		places
//	@Produce	json
//	@Param		id	path	int	true	"Place id"
//	@Success	204	"No Content"
//	@Failure	404	{object}	handler.ErrorResponse
//	@Router		/places/{id} [delete]
func (h *Handler) DeletePlace(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		h.ReturnWithError(w, r, err)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		h.ReturnWithError(w, r, err)
		return
	}

	h.ReturnWithJSON(w, http.StatusNoContent, nil)
}

// decodeInput reads a request body. A malformed body is the client's mistake,
// so the decoder error is reported as a validation failure rather than a 500.
// What counts as a *valid* place is the service's call, not this layer's.
func (h *Handler) decodeInput(w http.ResponseWriter, r *http.Request) (*PlaceInput, error) {
	var in PlaceInput
	if err := h.DecodeJSON(w, r, &in); err != nil {
		return nil, handler.ErrorWithMessage(handler.ECValidationError,
			"The request body is not valid JSON for a place.", err)
	}
	return &in, nil
}

// pathID reads {id}. The route pattern already restricts it to digits, so a
// parse failure here means the value overflowed int64.
func pathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil || id < 1 {
		return 0, invalid("id must be a positive integer")
	}
	return id, nil
}

// intParam parses an optional integer query parameter, 0 when absent.
func intParam(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}
	return strconv.Atoi(raw)
}
