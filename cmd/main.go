// Copyright 2020 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package cmd
import (
	"context"
	"net/http"
	"fmt"
	"github.com/go-kit/kit/log/level"
	"github.com/ines-cruz/json_exporter/config"
	"github.com/ines-cruz/json_exporter/jsonexporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
)

func Run() {
	promlogConfig := &promlog.Config{}
		logger := promlog.New(promlogConfig)
		config, err := config.LoadConfig("examples/config.yml")
		if err != nil {
			level.Error(logger).Log("msg", "Error loading config", "err", err) //nolint:errcheck
		}

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/probe", func(w http.ResponseWriter, req *http.Request) {
		probeHandler(w, req, config)
	})

	if err := http.ListenAndServe(":7979", nil); err != nil {
		level.Error(logger).Log("msg", "failed to start the server", "err", err) //nolint:errcheck
	}
}

func probeHandler(w http.ResponseWriter, r *http.Request, config config.Config) {

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	r = r.WithContext(ctx)

	registry := prometheus.NewPedanticRegistry()
	metrics, err := jsonexporter.CreateMetricsList( config)
	if err != nil {
		fmt.Println("Failed to create metrics list from config")
	}

	jsonMetricCollector := jsonexporter.JsonMetricCollector{JsonMetrics: metrics}
	target := r.URL.Query().Get("target")
	if target == "" {
		http.Error(w, "Target parameter is missing", http.StatusBadRequest)
		return
	}
	data, err := jsonexporter.FetchJson(ctx, target, config)
	if err != nil {
		http.Error(w, "Failed to fetch JSON response. TARGET: "+target+", ERROR: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	jsonMetricCollector.Data = data

	registry.MustRegister(jsonMetricCollector)
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)

}
