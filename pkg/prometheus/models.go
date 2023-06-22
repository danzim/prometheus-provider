package prometheus

type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
				Name      string `json:"__name__"`
				Namespace string `json:"namespace"`
			} `json:"metric"`
			Value []any `json:"value"`
		} `json:"result"`
	} `json:"data"`
}
