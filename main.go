package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"time"

	prometheus "github.com/danzim/prometheus-provider/pkg"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/open-policy-agent/frameworks/constraint/pkg/externaldata"
	"go.uber.org/zap"
)

var log logr.Logger

const (
	timeout    = 3 * time.Second
	apiVersion = "externaldata.gatekeeper.sh/v1alpha1"
)

type requestRatio struct {
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
}

func main() {
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("unable to initialize logger: %v", err))
	}
	log = zapr.NewLogger(zapLog)

	log.Info("starting server...")
	http.HandleFunc("/validate", validate)

	if err = http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}

}

func validate(w http.ResponseWriter, req *http.Request) {

	// only accept POST requests
	if req.Method != http.MethodPost {
		sendResponse(nil, "only POST is allowed", w)
		return
	}

	// read request body
	requestBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		sendResponse(nil, fmt.Sprintf("unable to read request body: %v", err), w)
		return
	}

	//ctx, cancel := context.WithTimeout(context.Background(), timeout)
	//defer cancel()

	// parse request body
	var providerRequest externaldata.ProviderRequest
	err = json.Unmarshal(requestBody, &providerRequest)
	if err != nil {
		sendResponse(nil, fmt.Sprintf("unable to unmarshal request body: %v", err), w)
		return
	}

	results := make([]externaldata.Item, 0)
	// iterate over all keys
	for _, key := range providerRequest.Request.Keys {

		cpuRequestRatio, memRequestRatio := prometheus.RequestUsageRatio(key)

		ratio := requestRatio{
			CPU:    math.Floor(cpuRequestRatio*100) / 100,
			Memory: math.Floor(memRequestRatio*100) / 100,
		}

		fmt.Println(ratio)

		results = append(results, externaldata.Item{
			Key:   key,
			Value: ratio,
		})

	}
	sendResponse(&results, "", w)
}

// sendResponse sends back the response to Gatekeeper.
func sendResponse(results *[]externaldata.Item, systemErr string, w http.ResponseWriter) {
	response := externaldata.ProviderResponse{
		APIVersion: apiVersion,
		Kind:       "ProviderResponse",
	}

	if results != nil {
		response.Response.Items = *results
	} else {
		response.Response.SystemError = systemErr
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

// func processTimeout(h http.HandlerFunc, duration time.Duration) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		ctx, cancel := context.WithTimeout(r.Context(), duration)
// 		defer cancel()

// 		r = r.WithContext(ctx)

// 		processDone := make(chan bool)
// 		go func() {
// 			h(w, r)
// 			processDone <- true
// 		}()

// 		select {
// 		case <-ctx.Done():
// 			sendResponse(nil, "operation timed out", w)
// 		case <-processDone:
// 		}
// 	}
// }
