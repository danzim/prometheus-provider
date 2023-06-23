package prometheus

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/danzim/prometheus-provider/pkg/utils"
	"k8s.io/klog/v2"
)

/* const (
	baseURL  = "https://prometheus-k8s-openshift-monitoring.apps.test-cluster.ocp.dz-co.de"
	resource = "/api/v1/query"
) */

func RequestUsageRatio(ns string) (float64, float64, error) {

	var values map[string]float64
	values = make(map[string]float64)
	var value float64

	/* queries := map[string]string{
		"cpuUsage":   "quantile_over_time(0.9, pod:container_cpu_usage:sum[7d])",
		"memUsage":   "quantile_over_time(0.9, namespace:container_memory_usage_bytes:sum[7d])",
		"cpuRequest": "namespace_cpu:kube_pod_container_resource_requests:sum",
		"memRequest": "namespace_memory:kube_pod_container_resource_requests:sum",
	} */

	queries := map[string]string{
		"cpuUsage":   utils.AppConfig.Prometheus.Query.CPU.Usage,
		"memUsage":   utils.AppConfig.Prometheus.Query.Memory.Usage,
		"cpuRequest": utils.AppConfig.Prometheus.Query.CPU.Request,
		"memRequest": utils.AppConfig.Prometheus.Query.Memory.Request,
	}

	for k, v := range queries {
		klog.InfoS(fmt.Sprintf("Query Prometheus: %s", v))
		strValue, err := prometheusQuery(v)
		if err != nil {
			klog.ErrorS(err, fmt.Sprintf("unable to request %s", k))
			return 0, 0, err
		}

		float64Value, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			klog.ErrorS(err, fmt.Sprintf("unable to convert string to integer [%s]", k))
			return 0, 0, err
		}
		if strings.HasPrefix(k, "cpu") {
			value = float64Value * 1000
		} else if strings.HasPrefix(k, "mem") {
			value = float64Value / 1024 / 1024
		} else {
			err := errors.New("internal error: value 0")
			panic(err)
		}

		values[k] = value
	}

	cpuRequestRatio := values["cpuUsage"] / values["cpuRequest"] * 100
	memRequestRatio := values["memUsage"] / values["memRequest"] * 100

	return cpuRequestRatio, memRequestRatio, nil

}

func prometheusQuery(query string) (string, error) {

	var body PrometheusResponse

	params := url.Values{}
	params.Add("query", query)

	u, err := url.ParseRequestURI(utils.AppConfig.Prometheus.URL)
	if err != nil {
		klog.ErrorS(err, "unable to parse prometheus query request URI")
		return "", err
	}
	u.Path = utils.AppConfig.Prometheus.Resource
	u.RawQuery = params.Encode()
	urlStr := fmt.Sprintf("%v", u)

	res, err := http.Get(urlStr)
	if err != nil {
		klog.ErrorS(err, "unable to query prometheus api")
		return "", err
	}

	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		klog.ErrorS(err, "unable to read response body")
		return "", err
	}

	klog.InfoS(fmt.Sprintf("Prometheus return body: %s", b))

	err = json.Unmarshal(b, &body)
	if err != nil {
		klog.ErrorS(err, "unable to unmarshal body")
		return "", err
	}

	value := fmt.Sprintf("%v", body.Data.Result[0].Value[1])

	//fmt.Println(res.Status)

	/* var value string

	switch query {
	case "quantile_over_time(0.9, pod:container_cpu_usage:sum[7d])":
		value = "0.006724902031765679"
	case "quantile_over_time(0.9, namespace:container_memory_usage_bytes:sum[7d])":
		value = "1761214464"
	case "namespace_cpu:kube_pod_container_resource_requests:sum":
		value = "0.45"
	case "namespace_memory:kube_pod_container_resource_requests:sum":
		value = "3221225473"
	} */

	return value, nil
}
