package providers

import (
	"compress/gzip"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/ines-cruz/json_exporter/config"
	pconfig "github.com/prometheus/common/config"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func GetAWS(config config.Config, ctx context.Context, endpoint string) ([]byte, error) {
	httpClientConfig := config.HTTPClientConfig
	client2, err := pconfig.NewClientFromConfig(httpClientConfig, "fetch_json", true)
	if err != nil {
		fmt.Println("Error generating HTTP client")
		return nil, err
	}

	fileContent := getDataFromS3File()

	var ex = parseColumn(fileContent)
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
	defer resp.Body.Close()

	defer func() {
		if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
			fmt.Println("Failed to discard body", "err", err)
		}
		_ = resp.Body.Close()
	}()

	if resp.StatusCode/100 != 2 {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil

}

// function we use to display errors and exit.
func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func getDataFromS3File() *os.File {
	// 1) Define your bucket and item names
	bucket := "strategic-blue-reports-cern/"
	item := "sb-cern-aws/20201101-20201201/d5569030-de28-3c3b-a5ca-4c54d40ba416/sb-cern-aws-1.csv.gz"

	file, err := os.Create("sb-cern-aws-1.csv")
	if err != nil {
		log.Fatal(err)
	}

	// Get AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2"),
		Credentials: credentials.NewStaticCredentials(
			"----",
			"---",
			""),
	})

	downloader := s3manager.NewDownloader(sess)

	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(item),
		})
	if err != nil {
		exitErrorf("Unable to download item %q, %v", item, err)
	}

	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")

	return file

}
func parseColumn(file *os.File) []map[string]interface{} {
	d := map[string]interface{}{}

	csvr, err := gzip.NewReader(file)
	if err != nil {
		fmt.Print("foo\x00\x00")
		log.Fatal(err)
	}
	defer csvr.Close()
	var final = []map[string]interface{}{}

	var usageAccountID string
	//var unblendedCost string
	// select just the columns we want

	cr := csv.NewReader(csvr)
	rec, err := cr.Read()
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range rec {
		fmt.Println(v)

		usageAccountID = v

	}

	d["projectid"] = usageAccountID
	//	d["amountSpent"] = unblendedCost
	//TODO  add name

	final = append(final, d)
	return final
}
