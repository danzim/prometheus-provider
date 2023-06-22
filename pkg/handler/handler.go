package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"

	"github.com/danzim/prometheus-provider/pkg/prometheus"
	"github.com/danzim/prometheus-provider/pkg/utils"
	"github.com/open-policy-agent/frameworks/constraint/pkg/externaldata"
)

type requestRatio struct {
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
}

func Handler(w http.ResponseWriter, req *http.Request) {

	// only accept POST requests
	if req.Method != http.MethodPost {
		utils.SendResponse(nil, "only POST is allowed", w)
		return
	}

	// read request body
	requestBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		utils.SendResponse(nil, fmt.Sprintf("unable to read request body: %v", err), w)
		return
	}

	//ctx, cancel := context.WithTimeout(context.Background(), timeout)
	//defer cancel()

	// parse request body
	var providerRequest externaldata.ProviderRequest
	err = json.Unmarshal(requestBody, &providerRequest)
	if err != nil {
		utils.SendResponse(nil, fmt.Sprintf("unable to unmarshal request body: %v", err), w)
		return
	}

	results := make([]externaldata.Item, 0)
	// iterate over all keys
	for _, key := range providerRequest.Request.Keys {

		cpuRequestRatio, memRequestRatio, err := prometheus.RequestUsageRatio(key)
		if err != nil {
			utils.SendResponse(nil, fmt.Sprintf("unable to request ratio: %v", err), w)
			return
		}

		ratio := requestRatio{
			CPU:    math.Floor(cpuRequestRatio*100) / 100,
			Memory: math.Floor(memRequestRatio*100) / 100,
		}

		results = append(results, externaldata.Item{
			Key:   key,
			Value: ratio,
		})

	}
	utils.SendResponse(&results, "", w)
}
