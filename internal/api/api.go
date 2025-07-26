package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

func BuildRouter(r *mux.Router) {
	buildMetricsRouter(r.PathPrefix("/metrics").Subrouter())
	buildLedgerRouter(r.PathPrefix("/ledger").Subrouter())
	buildPatronsRouter(r.PathPrefix("/patrons").Subrouter())
	buildHealthRouter(r.PathPrefix("/health").Subrouter())
}

/*
   How do the funds work? The community fund is a "bank account", where money goes in and money goes out and there's a total balance. The "sustaining" fund is maybe less like this? It's more about figuring out how much of a percentage of the spending we covered. So I guess we still want all the transactions, and all of the "input", and we'll just do a different calculation on it. We need all the "spending" categorized, all the "pledges" split, and then we can query buckets of months to see what the averages are.

   The key component is that both of these "funds" are contributed to by taking income split out of pateron payments. Maybe the difference is that one gets "reduced" by spending from the account, and the other does not. Actually, maybe the general fund *does* get spent from: going towards base costs. I guess whether or not we show that on the website is a different question. Okay, so these funds take money from pateron payments, and then get spent from. That means that I want a *separate* thing for tracking overall expenses? I don't want to replicate an entire double entry accounting system here. But I'm wanting to be able to show how the pateron income compares to the rest, and showcase categorized expenses. So we've got a "fund", which is an abstract bucket that money goes in and out of, and then a "ledger", which has income and expenses. Actually, these are both ledgers.

   Okay, so it's really just `GET+POST /ledger/{name}/transactions`. When a payment comes in, it automatically creates transactions into the community/general fund. We can manually `POST` to these ledgers to "spend" from them. Similarly, we can post income/expenses to the "balance" ledger for overall finances. These will need a label (?). Because, the goal is to know roughly where the income came from, and where the money went. The labels are for this.

   So for my idea: I'll be able to see the amount of money left (and total) that has gone to community fund. I can also determine how much income we've received (and where), and what we've spent (and where) to figure out the impact of the pateron. This should be good.
*/

type APIResponse struct {
	Error *APIError `json:"error"`
	Data  any       `json:"data"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}
