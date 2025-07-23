package stripe

import (
	"github.com/gorilla/mux"
	"github.com/stripe/stripe-go/v82"
)

var ENDPOINT_SECRET string
var updateRequests chan updateRequest

func Init(key string, endpointSecret string) {
	stripe.Key = key
	ENDPOINT_SECRET = endpointSecret

	updateRequests = make(chan updateRequest, 32)
	go scheduleResourceUpdates(updateRequests)
}

func BuildRouter(r *mux.Router) {

	r.HandleFunc("/webhook", HandleWebhook)
}
