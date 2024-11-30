package stripe

import "github.com/stripe/stripe-go/v81"

var ENDPOINT_SECRET string
var updateRequests chan updateRequest
var pageRebuildRequests chan<- int

func Init(key string, endpointSecret string, pageRebuildC chan<- int) {
	stripe.Key = key
	ENDPOINT_SECRET = endpointSecret

	pageRebuildRequests = pageRebuildC
	updateRequests = make(chan updateRequest, 32)
	go scheduleResourceUpdates(updateRequests)
}
