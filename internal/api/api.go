package api

import (
	"fmt"
	"net/http"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"github.com/gorilla/mux"
)

func BuildRouter(r *mux.Router) {

	// r.HandleFunc("/example", func(w http.ResponseWriter, r *http.Request) {})
	r.HandleFunc("/patrons/count", func(w http.ResponseWriter, r *http.Request) {
		summary, err := database.QuerySubscriptionSummary()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "{ \"count\": %d }", summary.Count)
	})
}
