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
	//credentials in example folder
	api, exists := os.LookupEnv("key")
	if !exists {
		fmt.Println("No env variables")
	}

	client, err := bigquery.NewClient(context.Background(), "billing-cern", option.WithAPIKey(api))
	if err != nil {
		fmt.Println("bigquery.NewClient", err)
	}
	defer client.Close()

	rows, err := query(ctx, client)
	if err != nil {
		fmt.Println(err)
	}
	var ex = printResults(os.Stdout, rows)

	thisMap := make(map[string](map[string]float64))
	thisMap["values"] = ex

	file, _ := json.MarshalIndent(thisMap, "", "")

	_ = ioutil.WriteFile("examples/output.json", file, 0644)

	/*
	   		//Aws

	   		client := return boto3.client(
	           api,
	           aws_access_key_id=aux.configs["aws"]["accessKey"],
	           aws_secret_access_key=aux.configs["aws"]["secretKey"],
	           region_name="us-east-1" )
	   			defer client.Close()


	   			response = ce.get_cost_and_usage(
	   		        TimePeriod={
	   		            "Start": starting,
	   		            "End":  ending
	   		        },
	   		        Granularity="DAILY",
	   		        Metrics=["UnblendedCost"],
	   		        GroupBy=[
	   		            {
	   		                "Type": "DIMENSION",
	   		                "Key": "LINKED_ACCOUNT"
	   		            }
	   		        ],
	   		        Filter={
	   		            "Dimensions": {
	   		                "Key": "LINKED_ACCOUNT",
	   		                "Values": list(__account_mappings().keys())
	   		            }
	   		        }
	   		    )

	   		    return list(map(__extract_cost, __extract_accounts(response)))

	*/

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
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}
	data2, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data2, nil
}

func query(ctx context.Context, client *bigquery.Client) (*bigquery.RowIterator, error) {

	q := client.Query(`
		SELECT cost,  sku.description, system_labels,
		FROM ` + "`sbsl_cern_billing_info.gcp_billing_export_v1_012C54_B3DAFC_973FAF` ")

	return q.Read(ctx)
}

// printResults prints results from a query to the _ public dataset.
func printResults(w io.Writer, iter *bigquery.RowIterator) map[string]float64 {
	var sumCost float64
	var sumGPU float64
	var sumTot int
	var sumCores float64
	var sumMem float64
	var sumVM float64
	d := map[string]float64{}

	for {

		sumTot = 1 + sumTot

		var row []bigquery.Value
		err := iter.Next(&row)

		if err == iterator.Done {
			break
		}

		var ex float64 = sumCost
		sumCost = row[0].(float64)
		sumCost = math.Round(sumCost*100)/100 + ex
		system_labels := row[2].([]bigquery.Value)

		if len(system_labels) > 0 {

			sumCores = getCores(system_labels, sumCores)
			sumVM = getVM(system_labels, sumVM)
			sumMem = getMem(system_labels, sumMem)

		}
		sku := row[1].(string)

		if strings.Contains(sku, "GPU") {
			sumGPU = 1 + sumGPU
		}

	}

	d["sumCost"] = sumCost
	d["sumGPU"] = sumGPU
	d["sumCores"] = sumCores
	d["sumMem"] = sumMem
	d["sumVM"] = sumVM

	return d
}
func getCores(sys []bigquery.Value, sumCores float64) float64 {
	var valCores = fmt.Sprint(sys[0])

	if strings.Contains(valCores, "cores") {
		s := strings.Fields(valCores)
		sCores := strings.Replace(s[1], "]", "", -1)

		example, err := strconv.ParseFloat(fmt.Sprint(sCores), 64)
		if err == nil {
			// TODO: Handle error.
		}
		sumCores = example + sumCores
	}
	return sumCores
}
func getVM(sys []bigquery.Value, sumVM float64) float64 {
	var valVM = fmt.Sprint(sys[1])

	if strings.Contains(valVM, "machine_spec") {
		sumVM = sumVM + 1
	}
	return sumVM
}

func getMem(sys []bigquery.Value, sumMem float64) float64 {
	var valMem = fmt.Sprint(sys[2])

	if strings.Contains(valMem, "memory") {

		example, err := strconv.ParseFloat(fmt.Sprint(valMem[len(valMem)-5:len(valMem)-1]), 64)
		if err == nil {
			// TODO: Handle error.
		}
		sumMem = example + sumMem
	}
	return sumMem
}
