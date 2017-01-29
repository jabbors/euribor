package main

import (
	"net/http"
	"strings"
)

const indexPage = `
<html>
<head>
<meta charset="utf-8">
<meta http-equiv="X-UA-Compatible" content="IE=edge">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Euribor rates</title>
<script src="https://code.highcharts.com/stock/5.0.7/highstock.js"></script>
<script src="https://code.jquery.com/jquery-3.1.1.min.js"></script>
<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css" integrity="sha384-BVYiiSIFeK1dGmJRAkycuHAHRg32OmUcww7on3RYdg4Va+PmSTsz/K68vbdEjh4u" crossorigin="anonymous">
<link rel="apple-touch-icon" sizes="57x57" href="https://github.com/jabbors/euribor/raw/master/gorates/chart-icon-57x57.png" />
<link rel="apple-touch-icon" sizes="72x72" href="https://github.com/jabbors/euribor/raw/master/gorates/chart-icon-72x72.png" />
<link rel="apple-touch-icon" sizes="114x114" href="https://github.com/jabbors/euribor/raw/master/gorates/chart-icon-114x114.png" />
<link rel="apple-touch-icon" sizes="144x144" href="https://github.com/jabbors/euribor/raw/master/gorates/chart-icon-144x144.png" />
<style>
.icon-license {
  font-size: 1rem;
  font-style: italic;
}
</style>
</head>
<body>
<script>
function loadChart(maturity, name) {
	$.getJSON('URLPLACEHOLDER/rates/app/hs/'+maturity, function (data) {
		// Create the chart
		Highcharts.stockChart('container', {

			rangeSelector: {
				buttons: [{
					type: 'week',
					count: 1,
					text: '1w'
				}, {
					type: 'month',
					count: 1,
					text: '1m'
				}, {
					type: 'month',
					count: 3,
					text: '3m'
				}, {
					type: 'month',
					count: 6,
					text: '6m'
				}, {
					type: 'year',
					count: 1,
					text: '1y'
				}, {
					type: 'year',
					count: 2,
					text: '2y'
				}, {
					type: 'year',
					count: 6,
					text: '6y'
				}, {
					type: 'all',
					text: 'All'
				}],
				selected: 4
			},

			navigator: {
				enabled: false
			},

			scrollbar: {
				enabled: false
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

window.onload=loadChart('3m', '3 months');
</script>
<h3>Euribor rates</h3>
<div>
<select onChange="loadChart(this.options[this.selectedIndex].value, this.options[this.selectedIndex].text)">
<option value="1w">1 week</option>
<option value="2w">2 weeks</option>
<option value="1m">1 month</option>
<option value="2m">2 months</option>
<option value="3m">3 months</option>
<option value="6m">6 months</option>
<option value="9m">9 months</option>
<option value="12m">12 months</option>
</select>
</div>
<div id="container"></div>
<div class="icon-license">
Icons made by <a href="http://www.flaticon.com/authors/gregor-cresnar" title="Gregor Cresnar">Gregor Cresnar</a> from <a href="http://www.flaticon.com" title="Flaticon">www.flaticon.com</a> is licensed by <a href="http://creativecommons.org/licenses/by/3.0/" title="Creative Commons BY 3.0" target="_blank">CC 3.0 BY</a>
</div>
</body>
</html>`

func renderWebapp(baseURL string) string {
	return strings.Replace(indexPage, "URLPLACEHOLDER", baseURL, -1)
}

func baseURL(r *http.Request, webRoot string) string {
	protocol := "https"
	if r.TLS == nil {
		protocol = "http"
	}
	path := strings.Replace(strings.Replace(r.RequestURI, "/webapp/", "", -1), "/webapp", "", -1)

	return protocol + "://" + r.Host + webRoot + path
}
