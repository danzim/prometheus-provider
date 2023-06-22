package prometheus

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"k8s.io/klog/v2"
)

const (
	baseURL  = "https://prometheus-k8s-openshift-monitoring.apps.test-cluster.ocp.dz-co.de"
	resource = "/api/v1/query"
)

func RequestUsageRatio(ns string) (float64, float64) {

	var values map[string]float64
	values = make(map[string]float64)
	var value float64

	queries := map[string]string{
		"cpuUsage":   "quantile_over_time(0.9, pod:container_cpu_usage:sum[7d])",
		"memUsage":   "quantile_over_time(0.9, namespace:container_memory_usage_bytes:sum[7d])",
		"cpuRequest": "namespace_cpu:kube_pod_container_resource_requests:sum",
		"memRequest": "namespace_memory:kube_pod_container_resource_requests:sum",
	}

	for k, v := range queries {
		strValue, err := prometheusQuery(v)
		if err != nil {
			klog.ErrorS(err, fmt.Sprintf("unable to request %s", k))
		}
		fmt.Println(strValue)
		float64Value, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			klog.ErrorS(err, fmt.Sprintf("unable to convert string to integer [%s]", k))
		}
		if strings.HasPrefix(k, "cpu") {
			value = float64Value * 1000
		} else if strings.HasPrefix(k, "mem") {
			value = float64Value / 1024 / 1024
		} else {
			err := errors.New("internal error: value 0")
			panic(err)
		}
		fmt.Println(value)

		values[k] = value
	}

	fmt.Println(values["cpuUsage"])
	fmt.Println(values["cpuRequest"])
	fmt.Println(values["memUsage"])
	fmt.Println(values["memRequest"])

	cpuRequestRatio := values["cpuUsage"] / values["cpuRequest"] * 100
	memRequestRatio := values["memUsage"] / values["memRequest"] * 100

	fmt.Println(cpuRequestRatio)
	fmt.Println(memRequestRatio)

	return cpuRequestRatio, memRequestRatio

}

func prometheusQuery(query string) (string, error) {

	// var body PrometheusResponse

	// params := url.Values{}

	// params.Add("query", query)

	// u, err := url.ParseRequestURI(baseURL)
	// if err != nil {
	// 	log.Error(err, "unable to parse prometheus query request URI")
	// }
	// u.Path = resource
	// u.RawQuery = params.Encode()
	// urlStr := fmt.Sprintf("%v", u)

	// res, err := http.Get(urlStr)
	// if err != nil {
	// 	log.Error(err, "unable to query prometheus api")
	// }

	// defer res.Body.Close()

	// b, err := io.ReadAll(res.Body)
	// if err != nil {
	// 	log.Error(err, "unable to read response body")
	// }

	// err = json.Unmarshal(b, &body)
	// if err != nil {
	// 	log.Error(err, "unable to unmarshal body")
	// }

	// value := fmt.Sprintf("%v", body.Data.Result[0].Value[1])

	// fmt.Println(res.Status)

	var value string

	switch query {
	case "quantile_over_time(0.9, pod:container_cpu_usage:sum[7d])":
		value = "0.006724902031765679"
	case "quantile_over_time(0.9, namespace:container_memory_usage_bytes:sum[7d])":
		value = "1761214464"
	case "namespace_cpu:kube_pod_container_resource_requests:sum":
		value = "0.45"
	case "namespace_memory:kube_pod_container_resource_requests:sum":
		value = "3221225473"
	}

	return value, nil
}
