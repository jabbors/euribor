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
function loadData(maturity, name) {
    $.getJSON('URLPLACEHOLDER/rates/app/hs/'+maturity, function (data) {
        // Create the chart
        Highcharts.stockChart('container', {

            rangeSelector: {
                selected: 1
            },

            title: {
                text: 'Euribor ' + name
            },

            series: [{
                name: maturity,
                data: data,
                tooltip: {
                    valueDecimals: 3
                }
            }]
        });
    });
};

window.onload=loadData('3m', '3 months');
</script>
<h3>Euribor rates</h3>
<div>
<ul>
<li onClick="loadData('1w', '1 week')"><font color="blue">1 week</font></li>
<li onClick="loadData('2w', '2 weeks')"><font color="blue">2 weeks</font></li>
<li onClick="loadData('1m', '1 month')"><font color="blue">1 month</font></li>
<li onClick="loadData('2m', '2 months')"><font color="blue">2 months</font></li>
<li onClick="loadData('3m', '3 months')"><font color="blue">3 months</font></li>
<li onClick="loadData('6m', '6 months')"><font color="blue">6 months</font></li>
<li onClick="loadData('9m', '9 months')"><font color="blue">9 months</font></li>
<li onClick="loadData('12m', '12 months')"><font color="blue">12 months</font></li>
</ul>
<div>
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
