package providers

import (
	"cloud.google.com/go/bigquery"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ines-cruz/json_exporter/config"
	"github.com/pkg/errors"
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
	"time"
)

func GetGoogle(config config.Config, ctx context.Context, endpoint string) ([]byte, error) {
	httpClientConfig := config.HTTPClientConfig
	client2, err := pconfig.NewClientFromConfig(httpClientConfig, "fetch_json", true, false)
	if err != nil {
		fmt.Println("Error generating HTTP client")
		return nil, err
	}
	// GCP
	//Create client
	//Name of the Google BigQuery DB
	env := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	client, err := bigquery.NewClient(ctx, "billing-cern", option.WithCredentialsFile(env))
	if err != nil {
		fmt.Printf("Client create error: %v\n", err)
	}

	row := client.Query(`SELECT cost,  sku.description, system_labels, project.id, usage_start_time, usage_end_time FROM sbsl_cern_billing_info.gcp_billing_export_v1_012C54_B3DAFC_973FAF WHERE project.id IS NOT NULL AND project.id!="billing-cern" ORDER BY project.id, usage_start_time`)
	rows, err := row.Read(ctx)
	if err != nil {
		fmt.Printf("Error2: %v\n", err)

	}
	var ex = printResults(rows)

	thisMap := make(map[string]([]map[string]interface{}))
	thisMap["values"] = ex

	file, _ := json.MarshalIndent(thisMap, "", "")

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
			fmt.Println("Failed to discard body", "err", err)
		}
		_ = resp.Body.Close()
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
func printResults(iter *bigquery.RowIterator) []map[string]interface{} {
	var sumCost float64
	var sumGPU float64
	var sumCores float64
	var sumMem float64
	var sumVM float64
	var projectid string
	var previous string
	var startTime time.Time
	var endTime time.Time
	var prevDate time.Time
	var data = []map[string]interface{}{}
	d := map[string]interface{}{}

	for {

		previous = projectid
		prevDate = startTime
		var row []bigquery.Value
		err := iter.Next(&row)

		if err == iterator.Done {
			break
		}
		if len(row) != 0 {
			startTime = row[4].(time.Time)

			projectid = row[3].(string)

			if (!verifyID(previous, projectid) && len(d) > 0) || (!checkMonth(startTime, prevDate) && len(d) > 0) {

				data = append(data, d)
				sumVM = 0
				sumMem = 0
				sumCost = 0
				sumCores = 0
				sumGPU = 0
				d = map[string]interface{}{}
			}
			sumCost = getSum(row, sumCost, projectid, previous)

			endTime = row[5].(time.Time)
			var valGPU = fmt.Sprint(row[1])
			if verifyStrings(valGPU, "GPU") {
				sumGPU = getDuration(projectid, previous, sumGPU, endTime, startTime)
			}
			system_labels := row[2].([]bigquery.Value)

			if len(system_labels) > 0 {

				var valCores = fmt.Sprint(system_labels[0])
				if verifyStrings(valCores, "cores") {
					sumCores = getDuration(projectid, previous, sumCores, endTime, startTime)
				}
				var valVM = fmt.Sprint(system_labels[1])
				if verifyStrings(valVM, "machine_spec") {
					sumVM = getVM(sumVM, previous, projectid)
				}
				var valMem = fmt.Sprint(system_labels[2])
				if verifyStrings(valMem, "memory") {
					sumMem = getMem(system_labels, sumMem, previous, projectid)
				}

			}

			d["month"] = startTime.Format("01-2006")
			d["amountSpent"] = math.Round(sumCost*100) / 100
			d["numberVM"] = sumVM
			d["memoryMB"] = sumMem
			d["CPUh"] = sumCores
			d["GPUh"] = sumGPU
			d["projectid"] = projectid

		}
	}
	data = append(data, d)
	return data
}

func checkMonth(startDate time.Time, prevDate time.Time) bool {

	return startDate.Month() == prevDate.Month() && startDate.Year() == prevDate.Year()

}
func checkDate(startDate time.Time) bool {

	return startDate.Month() == time.Now().Month() && startDate.Year() == time.Now().Year()

}
func verifyID(prev string, id string) bool {

	return strings.EqualFold(prev, id)
}
func verifyStrings(toCompare string, real string) bool {
	return strings.Contains(toCompare, real)
}

func getSum(row []bigquery.Value, sumCost float64, id string, prev string) float64 {

	if verifyID(prev, id) { //if we are still in the same project continue
		return row[0].(float64) + sumCost
	}
	return row[0].(float64)

}

func getDuration(projectid string, previous string, sum float64, endTime time.Time, startTime time.Time) float64 {

	if verifyID(previous, projectid) {
		diff := endTime.Sub(startTime).Hours() + sum
		return diff
	}
	return sum
}

func getVM(sumVM float64, prev string, id string) float64 {
	if verifyID(prev, id) {
		return sumVM + 1
	}
	return 0
}
func verifyMem(valMem string) float64 {
	example, err := strconv.ParseFloat(fmt.Sprint(valMem[len(valMem)-5:len(valMem)-1]), 64)
	if err == nil {
		// TODO: Handle error.
	}
	return example
}
func getMem(sys []bigquery.Value, sumMem float64, prev string, id string) float64 {
	if verifyID(prev, id) {
		return verifyMem(fmt.Sprint(sys[2])) + sumMem
	}
	return 0
}
