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

package jsonexporter

import (
	"cloud.google.com/go/bigquery"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ines-cruz/json_exporter/config"
	"github.com/kawamuray/jsonpath"
	"github.com/prometheus/client_golang/prometheus"
	pconfig "github.com/prometheus/common/config"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func MakeMetricName(parts ...string) string {
	return strings.Join(parts, "_")
}

func SanitizeValue(v *jsonpath.Result) (float64, error) {
	var value float64
	var boolValue bool
	var err error
	switch v.Type {
	case jsonpath.JsonNumber:
		value, err = parseValue(v.Value)
	case jsonpath.JsonString:
		// If it is a string, lets pull off the quotes and attempt to parse it as a number
		value, err = parseValue(v.Value[1 : len(v.Value)-1])
	case jsonpath.JsonNull:
		value = math.NaN()
	case jsonpath.JsonBool:
		if boolValue, err = strconv.ParseBool(string(v.Value)); boolValue {
			value = 1.0
		} else {
			value = 0.0
		}
	default:
		value, err = parseValue(v.Value)
	}
	if err != nil {
		// Should never happen.
		return -1.0, err
	}
	return value, err
}

func parseValue(bytes []byte) (float64, error) {
	value, err := strconv.ParseFloat(string(bytes), 64)
	if err != nil {
		return -1.0, fmt.Errorf("failed to parse value as float;value:<%s>", bytes)
	}
	return value, nil
}

func CreateMetricsList(c config.Config) ([]JsonMetric, error) {
	var metrics []JsonMetric
	for _, metric := range c.Metrics {
		switch metric.Type {
		case config.ValueScrape:
			constLabels := make(map[string]string)
			var variableLabels, variableLabelsValues []string
			for k, v := range metric.Labels {
				if len(v) < 1 || v[0] != '$' {
					// Static value
					constLabels[k] = v
				} else {
					variableLabels = append(variableLabels, k)
					variableLabelsValues = append(variableLabelsValues, v)
				}
			}
			jsonMetric := JsonMetric{
				Desc: prometheus.NewDesc(
					metric.Name,
					metric.Help,
					variableLabels,
					constLabels,
				),
				KeyJsonPath:     metric.Path,
				LabelsJsonPaths: variableLabelsValues,
			}
			metrics = append(metrics, jsonMetric)
		case config.ObjectScrape:
			for subName, valuePath := range metric.Values {
				name := MakeMetricName(metric.Name, subName)
				constLabels := make(map[string]string)
				var variableLabels, variableLabelsValues []string
				for k, v := range metric.Labels {
					if len(v) < 1 || v[0] != '$' {
						// Static value
						constLabels[k] = v
					} else {
						variableLabels = append(variableLabels, k)
						variableLabelsValues = append(variableLabelsValues, v)
					}
				}
				jsonMetric := JsonMetric{
					Desc: prometheus.NewDesc(
						name,
						metric.Help,
						variableLabels,
						constLabels,
					),
					KeyJsonPath:     metric.Path,
					ValueJsonPath:   valuePath,
					LabelsJsonPaths: variableLabelsValues,
				}
				metrics = append(metrics, jsonMetric)
			}
		default:
			return nil, fmt.Errorf("Unknown metric type: '%s', for metric: '%s'", metric.Type, metric.Name)
		}
	}
	return metrics, nil
}

func FetchJson(ctx context.Context, endpoint string, config config.Config) ([]byte, error) {
	httpClientConfig := config.HTTPClientConfig
	client2, err := pconfig.NewClientFromConfig(httpClientConfig, "fetch_json", true)
	if err != nil {
		fmt.Println("Error generating HTTP client")
		return nil, err
	}

	// GCP
	//Create client
	//Name of the Google BigQuery DB
	key := os.Getenv("key")
	table := os.Getenv("table")
	client, err := bigquery.NewClient(ctx, "billing-cern", option.WithAPIKey(key))
	if err != nil {
		fmt.Printf("Client create error: %v\n", err)
	}

	row := client.Query(`SELECT cost,  sku.description, system_labels, project.id
		FROM ` + table + "ORDER BY project.id")
	rows, err := row.Read(ctx)
	if err != nil {
		fmt.Printf("Error2: %v\n", err)

	}
	var ex = printResults(rows)

	thisMap := make(map[string]interface{})
	thisMap["values"] = ex

	file, _ := json.MarshalIndent(thisMap, "[]", "")

	_ = ioutil.WriteFile("examples/output.json", file, 0644)

	req, err := http.NewRequest("GET", endpoint, nil)
	req = req.WithContext(ctx)
	if err != nil {
		fmt.Println("Failed to create request")
		return nil, err
	}
	for key, value := range config.Headers {
		req.Header.Add(key, value)
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Add("Accept", "application/json")
	}
	resp, err := client2.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
			fmt.Println("Failed to discard body", "err", err) //nolint:errcheck
		}
		resp.Body.Close()
	}()

	if resp.StatusCode/100 != 2 {
		return nil, errors.New(resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// printResults prints results from a query to the _ public dataset.
func printResults(iter *bigquery.RowIterator) map[string]interface{} {
	var sumCost float64
	var sumGPU float64
	var sumCores float64
	var sumMem float64
	var sumVM float64
	var projectid string
	var previous string
	d := map[string]interface{}{}
	for {

		var row []bigquery.Value
		err := iter.Next(&row)

		if err == iterator.Done {
			break
		}
		projectid = row[3].(string)

		sumCost = getSum(row, sumCost, projectid, previous)

		system_labels := row[2].([]bigquery.Value)

		if len(system_labels) > 0 {

			sumCores = getCores(system_labels, sumCores, previous, projectid)
			sumVM = getVM(system_labels, sumVM, previous, projectid)
			sumMem = getMem(system_labels, sumMem, previous, projectid)

		}
		sku := row[1].(string)
		if projectid == previous && strings.Contains(sku, "GPU") { //if we are still in the same project continue
			sumGPU = 1 + sumGPU
		} else if strings.Contains(sku, "GPU") {
			sumGPU = 1
		} else {
			sumGPU = 0
		}

		previous = projectid
		projectid = row[3].(string)
	}

	d["sumCost"] = sumCost
	d["sumGPU"] = sumGPU
	d["sumCores"] = sumCores
	d["sumMem"] = sumMem
	d["sumVM"] = sumVM
	d["projectid"] = projectid
	return d
}
func getSum(row []bigquery.Value, sumCost float64, id string, prev string) float64 {
	var ex = sumCost

	if id == prev { //if we are still in the same project continue
		sumCost = row[0].(float64)
		sumCost = math.Round(sumCost*100)/100 + ex
	} else { //else start from 0
		sumCost = row[0].(float64)
		ex = 0
	}

	return sumCost
}

//TODO improve the code, many repetitions

func getCores(sys []bigquery.Value, sumCores float64, prev string, id string) float64 {
	var valCores = fmt.Sprint(sys[0])

	if strings.Contains(valCores, "cores") && id == prev {
		s := strings.Fields(valCores)
		sCores := strings.Replace(s[1], "]", "", -1)

		example, err := strconv.ParseFloat(fmt.Sprint(sCores), 64)
		if err == nil {
			// TODO: Handle error.
		}
		sumCores = example + sumCores
	} else if strings.Contains(valCores, "cores") {
		s := strings.Fields(valCores)
		sCores := strings.Replace(s[1], "]", "", -1)

		example, err := strconv.ParseFloat(fmt.Sprint(sCores), 64)
		if err == nil {
			// TODO: Handle error.
		}
		sumCores = example
	} else {
		sumCores = 0
	}
	return sumCores
}
func getVM(sys []bigquery.Value, sumVM float64, prev string, id string) float64 {
	var valVM = fmt.Sprint(sys[1])

	if strings.Contains(valVM, "machine_spec") && id == prev {
		sumVM = sumVM + 1
	} else if strings.Contains(valVM, "machine_spec") {
		sumVM = 1
	}
	return sumVM
}

func getMem(sys []bigquery.Value, sumMem float64, prev string, id string) float64 {
	var valMem = fmt.Sprint(sys[2])

	if strings.Contains(valMem, "memory") && id == prev {

		example, err := strconv.ParseFloat(fmt.Sprint(valMem[len(valMem)-5:len(valMem)-1]), 64)
		if err == nil {
			// TODO: Handle error.
		}
		sumMem = example + sumMem
	} else if strings.Contains(valMem, "memory") {
		sumMemNew, err := strconv.ParseFloat(fmt.Sprint(valMem[len(valMem)-5:len(valMem)-1]), 64)
		if err == nil {
			// TODO: Handle error.
		}
		sumMem = sumMemNew
	} else {
		sumMem = 0
	}
	return sumMem
}
