package stripe

import "github.com/stripe/stripe-go/v81"

var ENDPOINT_SECRET string
var updateResourceC chan updateRequest
var requestPageRebuild chan<- int

func Init(key string, endpointSecret string, pageBuildC chan<- int) {
	stripe.Key = key
	ENDPOINT_SECRET = endpointSecret

	requestPageRebuild = pageBuildC
	updateResourceC = make(chan updateRequest, 32)
	go scheduleResourceUpdates(updateResourceC)
}
