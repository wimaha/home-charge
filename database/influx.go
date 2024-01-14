package database

import (
	"context"
	"fmt"
	"strconv"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type InfluxClient struct {
	ip                   string
	token                string
	organisation         string
	productionTotalQuery string
}

type Query struct {
	ProductionTotal string
}

func NewInfluxClient(ip string, port string, token string, organisation string, productionTotalQuery string) *InfluxClient {
	return &InfluxClient{
		// Create a new client using an InfluxDB server base URL and an authentication token
		ip:                   fmt.Sprintf("http://%s:%s", ip, port),
		token:                token,
		organisation:         organisation,
		productionTotalQuery: productionTotalQuery,
	}
}

func (i *InfluxClient) InfluxProductionTotal() (float64, error) {
	var retValue float64 = -1
	var retError error = nil
	// Create a new client using an InfluxDB server base URL and an authentication token
	client := influxdb2.NewClient(i.ip, i.token)
	// Get query client
	queryAPI := client.QueryAPI(i.organisation)
	// get QueryTableResult
	result, err := queryAPI.Query(context.Background(), i.productionTotalQuery)
	if err == nil {
		// Iterate over query response
		for result.Next() {
			// Access data
			valueString := fmt.Sprintf("%v", result.Record().Value())
			// Parse the string value to float64
			valueFloat, err := strconv.ParseFloat(valueString, 64)
			if err != nil {
				retError = fmt.Errorf("error parsing value to float64: %v", err)
			} else {
				retValue = valueFloat
			}
		}
		// check for an error
		if result.Err() != nil {
			retError = fmt.Errorf("query parsing error: %s", result.Err().Error())
		}
	} else {
		retError = fmt.Errorf("query error: %s", err)
	}
	// Ensures background processes finishes
	client.Close()

	//fmt.Printf("Value: %f\n", retValue)

	return retValue, retError
}
