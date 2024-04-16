package exporter

import (
	"encoding/json"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

func ProbeHandler(w http.ResponseWriter, r *http.Request, c *AWSClient) {
	bucket := r.URL.Query().Get("bucket")
	if bucket == "" {
		http.Error(w, "bucket parameter is missing", http.StatusBadRequest)
		return
	}

	prefix := r.URL.Query().Get("prefix")
	delimiter := r.URL.Query().Get("delimiter")

	exporter := &Exporter{
		bucket:    bucket,
		prefix:    prefix,
		delimiter: delimiter,
		svc:       *c,
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(exporter)

	// Serve
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

type discoveryTarget struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

// DiscoveryHandler Setup HTTP service discovery for Prometheus
func DiscoveryHandler(w http.ResponseWriter, r *http.Request, c *AWSClient) {
	result, err := c.s3.ListBuckets(&c.s3.bucketsInputs)
	if err != nil {
		log.Errorln(err)
		http.Error(w, "error listing buckets", http.StatusInternalServerError)
		return
	}

	targets := []discoveryTarget{}
	for _, b := range result.Buckets {
		name := c.StringValue(b.Name)
		if name != "" {
			t := discoveryTarget{
				Targets: []string{r.Host},
				Labels: map[string]string{
					"__param_bucket": name,
				},
			}
			targets = append(targets, t)
		}
	}

	data, err := json.Marshal(targets)
	if err != nil {
		http.Error(w, "error marshalling json", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(data); err != nil {
		log.Fatalln(err)
	}
}
