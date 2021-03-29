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
	"strings"
	"time"
)

func GetAWS(config config.Config, ctx context.Context, endpoint string) ([]byte, error) {
	httpClientConfig := config.HTTPClientConfig
	client2, err := pconfig.NewClientFromConfig(httpClientConfig, "fetch_json", true)
	if err != nil {
		fmt.Println("Error generating HTTP client")
		return nil, err
	}
	bucket := "strategic-blue-reports-cern/"

	t := time.Now()
	time1 := t.Format("200601")
	time2 := t.AddDate(0, 1, 0).Format("200601")
	timePeriod := string(time1) + "01-" + string(time2) + "01"

	manifestFile := getDataFromS3File(bucket, "/sb-cern-aws/"+timePeriod+"/sb-cern-aws-Manifest.json")

	pathFromManifest := getField(manifestFile)
	fileContent := getDataFromS3File(bucket, pathFromManifest)

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

func getDataFromS3File(bucket string, path string) *os.File {
	file, err := os.Create("test.csv")
	if err != nil {
		log.Fatal(err)
	}

	// Get AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2"),
		Credentials: credentials.NewStaticCredentials(
			"---",
			"---",
			""),
	})

	downloader := s3manager.NewDownloader(sess)

	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(path),
		})
	if err != nil {
		exitErrorf("Unable to download item %q, %v", path, err)
	}

	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")

	return file

}
func parseColumn(file *os.File) []map[string]interface{} {
	d := map[string]interface{}{}
	var previous string

	csvr, err := gzip.NewReader(file)
	if err != nil {
		fmt.Print("foo\x00\x00")
		log.Fatal(err)
	}
	fmt.Print("foo\x00\x00")
	defer csvr.Close()
	var final = []map[string]interface{}{}

	var usageAccountID string
	var unblendedCost string
	// select just the columns we want

	cr := csv.NewReader(csvr)

	for {
		previous = usageAccountID
		row, err := cr.Read()
		// Stop at EOF.
		if err == io.EOF {
			break
		}
		if !verifyID(previous, usageAccountID) && len(d) > 0 {

			final = append(final, d)
			d = map[string]interface{}{} //TODO mudar 'para ter varios como no goole
		}

		for range row {
			//TODO add more fields
			usageAccountID = row[8]
			unblendedCost = row[22]

			d["projectid"] = usageAccountID
			d["amountSpent"] = unblendedCost

		}
	}
	//TODO  match and add name

	final = append(final, d)
	return final
}

func getField(file *os.File) string {
	// Open our jsonFile
	jsonFile, err := os.Open(file.Name())
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]interface{}
	json.Unmarshal([]byte(byteValue), &result)

	fmt.Println(result["reportKeys"])

	b, _ := json.Marshal(result["reportKeys"])

	s2 := string(b)
	s2 = strings.TrimRight(s2, "\\\"]\"")

	str3 := s2
	str3 = strings.TrimLeft(s2, "\"[\\")

	return str3

}
