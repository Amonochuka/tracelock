package httpdir

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func parseIDParam(r *http.Request, param string) (int, error) {
	return strconv.Atoi(chi.URLParam(r, param))
}

//helps in display of records from a certain point
//defaults set are 20 and 0
func parsePagination(r *http.Request) (limit, offset int) {
	limit = 20
	offset = 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}
	return
}
