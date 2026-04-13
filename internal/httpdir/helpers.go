package httpdir

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func parseIDParam(r *http.Request, param string) (int, error) {
	return strconv.Atoi(chi.URLParam(r, param))
}
