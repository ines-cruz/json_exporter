package providers

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/athena"
)

const table = "testexporterdb.testtable"
const db = "testexporterdb"
const outputBucket = "s3://strategic-blue-reports-cern/sb-cern-aws/"

func GetAWS() { //TODO will have to return  ([]byte, error)   that will be our data
	awscfg := &aws.Config{}
	awscfg.WithRegion("us-east-2")
	// Create the session that the service will use.
	sess := session.Must(session.NewSession(awscfg))

	svc := athena.New(sess, aws.NewConfig().WithRegion("us-east-2"))
	var s athena.StartQueryExecutionInput
	s.SetQueryString("SELECT * FROM \"testexporterdb\".\"testtable\";")

	var q athena.QueryExecutionContext
	q.SetDatabase(db)
	s.SetQueryExecutionContext(&q)

	var r athena.ResultConfiguration
	r.SetOutputLocation(outputBucket)
	s.SetResultConfiguration(&r)

	result, err := svc.StartQueryExecution(&s)
	if err != nil {
		fmt.Println(err)
		//	return nil, err
	}
	fmt.Println("StartQueryExecution result:")
	fmt.Println(result.GoString())

	var qri athena.GetQueryExecutionInput
	qri.SetQueryExecutionId(*result.QueryExecutionId)

	var qrop *athena.GetQueryExecutionOutput
	duration := time.Duration(2) * time.Second // Pause for 2 seconds

	for {
		qrop, err = svc.GetQueryExecution(&qri)
		if err != nil {
			fmt.Println(err)
			//return nil, err
		}
		if *qrop.QueryExecution.Status.State != "RUNNING" {
			break
		}
		fmt.Println("waiting.")
		time.Sleep(duration)

	}
	if *qrop.QueryExecution.Status.State == "SUCCEEDED" {

		var ip athena.GetQueryResultsInput
		ip.SetQueryExecutionId(*result.QueryExecutionId)

		op, err := svc.GetQueryResults(&ip)
		if err != nil {
			fmt.Println(err)
			//return nil, err
		}
		fmt.Printf("%+v", op)
	} else {
		fmt.Println(*qrop.QueryExecution.Status.State)

	}
}
