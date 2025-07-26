package api

import (
	"net/http"
	"strconv"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"github.com/gorilla/mux"
)

// Patron represents an active subscriber.
type Patron struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func buildPatronsRouter(r *mux.Router) {
	r.HandleFunc("", handleListPatrons).Methods("GET")
}

func handleListPatrons(
	w http.ResponseWriter,
	r *http.Request,
) {

	// TODO: validate Authorization header

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 100
	}

	rows, err := database.QueryCustomers(limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, &APIError{"500", "list error"})
		return
	}

	patrons := []Patron{}
	for _, c := range rows {
		updated := c.Created
		if c.Updated.Valid {
			updated = c.Updated.Int64
		}
		patrons = append(patrons, Patron{
			ID:        c.ID,
			Email:     c.Email,
			Name:      c.Name,
			CreatedAt: time.Unix(c.Created, 0).Format(time.RFC3339),
			UpdatedAt: time.Unix(updated, 0).Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, APIResponse{nil, patrons})
}
