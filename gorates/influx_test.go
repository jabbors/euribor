package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestStructUnmarshal(t *testing.T) {
	responseStr := "{\"results\":[{\"series\":[{\"name\":\"rates\",\"columns\":[\"time\",\"maturity\",\"value\"],\"values\":[[\"2016-10-06T00:00:00Z\",\"12m\",-0.07325],[\"2016-10-06T00:00:00Z\",\"9m\",-0.1325],[\"2016-10-06T00:00:00Z\",\"6m\",-0.2105],[\"2016-10-06T00:00:00Z\",\"3m\",-0.31275],[\"2016-10-06T00:00:00Z\",\"2w\",-0.3725],[\"2016-10-06T00:00:00Z\",\"2m\",-0.339],[\"2016-10-06T00:00:00Z\",\"1w\",-0.38075000000000003],[\"2016-10-06T00:00:00Z\",\"1m\",-0.3715]]}]}]}"

	response := influxResponse{}

	err := json.Unmarshal([]byte(responseStr), &response)
	if err != nil {
		panic(err)
	}

	if len(response.Results) != 1 {
		t.Errorf("expected 1 results, got %d", len(response.Results))
	}
	if len(response.Results[0].Series) != 1 {
		t.Errorf("expected 1 series, got %d", len(response.Results[0].Series))
	}
	if response.Results[0].Series[0].Name != "rates" {
		t.Errorf("expected 'rates' as name, got '%s'", response.Results[0].Series[0].Name)
	}
	if len(response.Results[0].Series[0].Columns) != 3 {
		t.Errorf("expected 3 columns, got %d", len(response.Results[0].Series[0].Columns))
	}
	if len(response.Results[0].Series[0].Values) != 8 {
		t.Errorf("expected 8 values, got %d", len(response.Results[0].Series[0].Values))
	}
}

func TestTransformInfluxValueToRate(t *testing.T) {
	// ["2016-10-06T00:00:00Z","12m",-0.07325]
	var value []interface{}
	value = append(value, "2016-10-06T00:00:00Z")
	value = append(value, "12m")
	value = append(value, -0.07325)

	maturity, rate, err := transformInfluxValueToRate(value)
	if err != nil {
		panic(err)
	}
	if maturity != "12m" {
		t.Errorf("incorrect maturity")
	}
	if rate.Date.Format("2006-01-02") != "2016-10-06" || rate.Value != -0.07325 {
		t.Errorf("incorrect rate")
	}
}

func TestReflectInterface(t *testing.T) {
	var values []interface{}
	var value []interface{}
	value = append(value, "2016-10-06T00:00:00Z")
	value = append(value, "12m")
	value = append(value, -0.07325)
	values = append(values, value)
	values = append(values, value)

	if reflect.TypeOf(values).Kind() != reflect.Slice {
		t.Errorf("no slice, should be slice")
	}

	for _, v := range values {
		if reflect.TypeOf(v).Kind() != reflect.Slice {
			t.Errorf("no slice, should be slice")
		}
	}
}
