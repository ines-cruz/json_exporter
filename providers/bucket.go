package providers

import (
	"context"
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
	// 1) Define your bucket and item names
	bucket := "strategic-blue-reports-cern"
	item := "sb-cern-aws/"

	fileContent := getDataFromS3File(bucket, item)

	jsonData, _ := json.Marshal(fileContent)

	var ex = extractData(jsonData)
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
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil

}

func getDataFromS3File(bucket string, s3File string) string {
	//the only writable directory in the lambda is /tmp
	file, err := os.Create("/tmp/" + s3File)
	if err != nil {
		fmt.Errorf("Unable to open file %q, %v", s3File, err)
	}

	defer file.Close()
	// Get AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2"),
		Credentials: credentials.NewStaticCredentials(
			"----------",
			"----------",
			""),
	})
	if err != nil {
		fmt.Println("Error creating the session", err)
		//return
	}

	downloader := s3manager.NewDownloader(sess)

	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(s3File),
		})
	if err != nil {
		fmt.Errorf("Unable to download s3File %q, %v", s3File, err)
	}

	dat, err := ioutil.ReadFile(file.Name())

	if err != nil {
		fmt.Errorf("Cannot read the file")
	}

	return string(dat)

}

func extractData(data []byte) []map[string]interface{} {

	//TODO and add name
	d := map[string]interface{}{}
	var final = []map[string]interface{}{}

	for i := 0; i < len(data); i++ {

		d["x"] = data[i]
	}
	//	d["groupid"] = MatchNametoID()

	final = append(final, d)
	return final
}
