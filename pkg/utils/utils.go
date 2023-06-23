package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/open-policy-agent/frameworks/constraint/pkg/externaldata"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

var AppConfig Config

const (
	apiVersion = "externaldata.gatekeeper.sh/v1beta1"
	kind       = "ProviderResponse"
)

func init() {
	currentDir, err := os.Getwd()
	if err != nil {
		klog.ErrorS(err, "unable to get current dir")
	}
	configFile := filepath.Join(currentDir, "config.yaml")
	err = loadConfig(configFile)
	if err != nil {
		klog.ErrorS(err, "unable to load config file")
	}

}

// sendResponse sends back the response to Gatekeeper.
func SendResponse(results *[]externaldata.Item, systemErr string, w http.ResponseWriter) {
	response := externaldata.ProviderResponse{
		APIVersion: apiVersion,
		Kind:       kind,
		Response: externaldata.Response{
			Idempotent: true, // mutation requires idempotent results
		},
	}

	if results != nil {
		response.Response.Items = *results
	} else {
		response.Response.SystemError = systemErr
	}

	klog.InfoS("sending response", "response", response)

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		klog.ErrorS(err, "unable to encode response")
		os.Exit(1)
	}
}

func loadConfig(file string) error {
	// read config file
	configData, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(configData, &AppConfig)
	if err != nil {
		return err
	}

	klog.InfoS("loaded config:")
	klog.InfoS(fmt.Sprintf("Prometheus URL: %s", AppConfig.Prometheus.URL))
	klog.InfoS(fmt.Sprintf("Prometheus Resource: %s", AppConfig.Prometheus.Resource))

	return nil

}
