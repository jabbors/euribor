package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"time"
)

const influxHost = "http://127.0.0.1:8086"
const influxDB = "euribor"
const influxMesurement = "rates"

type influxResponse struct {
	Results []influxResult `json:"results"`
}

type influxResult struct {
	Series []influxSerie `json:"series"`
}

type influxSerie struct {
	Name    string        `json:"name"`
	Columns []string      `json:"columns"`
	Values  []interface{} `json:"values"`
}

// queryInflux queries the influx database for downsampled and current rates
func queryInflux(retention string) influxSerie {
	query := fmt.Sprintf("SELECT * FROM %s.%s", retention, influxMesurement)
	v := url.Values{}
	v.Set("db", influxDB)
	v.Set("q", query)

	uri := influxHost + "/query?" + v.Encode()

	response, err := http.Get(uri)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	// decode response
	resp := influxResponse{}
	err = json.Unmarshal(contents, &resp)
	if err != nil {
		panic(err)
	}

	if len(resp.Results) > 0 && len(resp.Results[0].Series) > 0 {
		return resp.Results[0].Series[0]
	}
	return influxSerie{}
}

// transformInfluxValueToRate transforms a value obtained from the
// influx database to a JSON serializeable rate struct
func transformInfluxValueToRate(value []interface{}) (string, rate, error) {
	if len(value) != 3 {
		return "", rate{}, fmt.Errorf("incorrect number of fields")
	}
	m := ""
	r := rate{}
	// parse date
	if reflect.TypeOf(value[0]).Kind() == reflect.String {
		d := reflect.ValueOf(value[0]).Interface().(string)
		date, err := time.Parse("2006-01-02T15:04:05Z", d)
		if err != nil {
			return m, r, nil
		}
		r.Date = date
	}
	// parse maturity
	if reflect.TypeOf(value[1]).Kind() == reflect.String {
		m = reflect.ValueOf(value[1]).Interface().(string)
	}
	// parse rate value
	if reflect.TypeOf(value[2]).Kind() == reflect.Float64 {
		f := reflect.ValueOf(value[2]).Interface().(float64)
		r.Value = f
	}
	return m, r, nil
}
