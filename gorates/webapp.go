package main

import (
	"net/http"
	"strings"
)

const indexPage = `
<html>
<head>
<title>Euribor rates</title>
<script src="https://code.highcharts.com/stock/5.0.7/highstock.js"></script>
<script src="https://code.jquery.com/jquery-3.1.1.min.js"></script>
</head>
<body>
<script>
function loadData() {
    $.getJSON('URLPLACEHOLDER/rates/app/hs/1w', function (data) {
        // Create the chart
        Highcharts.stockChart('container', {

            rangeSelector: {
                selected: 1
            },

            title: {
                text: 'Euribor 1 week'
            },

            series: [{
                name: '1 week',
                data: data,
                tooltip: {
                    valueDecimals: 3
                }
            }]
        });
    });
};

window.onload=loadData;
</script>
<h3>Euribor rates</h3>
<div id="container" style="height: 400px; min-width: 310px"></div>
</body>
</html>`

func renderWebapp(baseURL string) string {
	return strings.Replace(indexPage, "URLPLACEHOLDER", baseURL, -1)
}

func baseURL(r *http.Request) string {
	protocol := "https"
	if r.TLS == nil {
		protocol = "http"
	}
	path := strings.Replace(strings.Replace(r.RequestURI, "/webapp/", "", -1), "/webapp", "", -1)

	return protocol + "://" + r.Host + path
}
