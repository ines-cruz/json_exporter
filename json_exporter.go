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

package main

import (
	"github.com/prometheus-community/json_exporter/harness"
	"github.com/prometheus-community/json_exporter/jsonexporter"
	"context"
	"fmt"
	"strings"
	"google.golang.org/api/iterator"
	"encoding/json"
	"io/ioutil"
	"cloud.google.com/go/bigquery"
	"google.golang.org/api/option"
	"os"
	"io"
	"math"
	"strconv"
)

func main() {

	ctx := context.Background()
//Create client
//Name of the Google BigQuery DB
//credentials in example folder
		client, err :=bigquery.NewClient(context.Background(),  "cobalt-aria-281116", option.WithCredentialsFile("example/credentials.json"))
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
		thisMap["values"]=ex

		file, _ := json.MarshalIndent( thisMap, "", "")

		_ = ioutil.WriteFile("example/output.json", file, 0644)

	opts := harness.NewExporterOpts("json_exporter", jsonexporter.Version)
	opts.Usage = "[OPTIONS] HTTP_ENDPOINT CONFIG_PATH"
	opts.Init = jsonexporter.Init
	harness.Main(opts)
}

func query(ctx context.Context, client *bigquery.Client) (*bigquery.RowIterator, error) {

	q := client.Query(`
		SELECT cost,  sku.description, system_labels,
		FROM `+"`MyBilling.gcp_billing_export_v1_0190AC_ADCAC5_4B89E9` "				)

	return q.Read(ctx)
}


// printResults prints results from a query to the _ public dataset.
func printResults(w io.Writer, iter *bigquery.RowIterator) map[string]float64{
	var  sumCost float64
	var sumGPU float64
	var sumTot int
	var sumCores float64
	var sumMem float64
	var sumVM float64
	d := map[string]float64{}


	for {

		sumTot=1+sumTot

		var row []bigquery.Value
		err := iter.Next(&row)

		if err == iterator.Done {
			break
		}

		var	ex float64 =sumCost
		sumCost=row[0].(float64)
		sumCost = math.Round(sumCost*100)/100 + ex
		system_labels:=row[2].([]bigquery.Value)

		if(len(system_labels) > 0){

			sumCores=getCores(system_labels, sumCores)
			sumVM=getVM(system_labels,  sumVM)
			sumMem=getMem(system_labels,  sumMem)

		}
		sku := row[1].(string)

		if strings.Contains(sku, "GPU"){
			sumGPU=1+sumGPU
		}

	}

	d["sumCost"]=  sumCost
	d["sumGPU"]= sumGPU
	d["sumCores"]=  sumCores
	d["sumMem"]=  sumMem
	d["sumVM"]=  sumVM


	return d
}
func getCores(sys []bigquery.Value, sumCores float64) float64{
	var valCores =fmt.Sprint(sys[0])

	if strings.Contains(valCores,"cores" ) {
		s := strings.Fields(valCores)
		sCores:= strings.Replace(s[1], "]","", -1)

		example, err :=  strconv.ParseFloat(fmt.Sprint(sCores),  64)
		if(err== nil){
			// TODO: Handle error.
		}
		sumCores=example+sumCores
	}
	return sumCores
}
func getVM(sys []bigquery.Value, sumVM float64) float64{
	var valVM =fmt.Sprint(sys[1])

	if strings.Contains(valVM,"machine_spec" ) {
		sumVM= sumVM+1
	}
	return sumVM
}


func getMem(sys []bigquery.Value, sumMem float64 ) float64{
	var valMem =fmt.Sprint(sys[2])

	if strings.Contains(valMem,"memory" ) {

		example, err := strconv.ParseFloat(fmt.Sprint(valMem[len(valMem)-5:len(valMem)-1]), 64)
		if(err== nil){
			// TODO: Handle error.
		}
		sumMem=example+sumMem
	}
	return sumMem
}
