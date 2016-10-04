#!/bin/bash

fetch_page() {
    maturity=$1
    page="/euribor-rate-${maturity}.asp"
    file="/tmp/${page}"
    url="http://www.euribor-rates.eu/${page}"
    if [ ! -f ${file} ]
    then
        echo "fetching page ${url}"
        curl ${url} -o ${file}
    fi
}

parse_date() {
    maturity=$1
    page="/euribor-rate-${maturity}.asp"
    file="/tmp/${page}"
    cat ${file} | grep "TABLE" -A 8 -m 1 | grep -o -e "[0-9]\{2\}-[0-9]\{2\}-[0-9]\{4\}"
}

parse_rate() {
    maturity=$1
    page="/euribor-rate-${maturity}.asp"
    file="/tmp/${page}"
    cat ${file} | grep "TABLE" -A 8 -m 1 | grep -o -e "-\?[0-9]\{1,2\}\.[0-9]\{3\}"
}

last_inserted_time() {
    maturity=$(short_maturity_string $1)
    ts=$(influx -format csv -execute "SELECT * FROM euribor.week.rates WHERE maturity='${maturity}' ORDER BY time DESC LIMIT 1" | tail -n 1 | cut -d ',' -f2)
    date -d @$(( ts / 1000000000 )) +%m-%d-%Y
}

insert_data() {
    maturity=$(short_maturity_string $1)
    date=$2
    rate=$3
    ts=$(date_to_timestamp_nano ${date})
    echo "inserting data: INSERT rates,maturity=${maturity} value=${rate} ${ts} (${date})"
    influx -execute "INSERT INTO euribor.week rates,maturity=${maturity} value=${rate} ${ts}"
    influx -execute "INSERT INTO euribor.month rates,maturity=${maturity} value=${rate} ${ts}"
}

populate_csv_files() {
    maturity=$(short_maturity_string $1)
    date=$(date_normalized $2)
    rate=$3
    file="euribor-rates-${maturity}.csv"
    echo "append line' ${date},${rate}' to ${file}"
    echo "${date},${rate}" >> ${file}
}

short_maturity_string() {
    maturity=$1
    echo "${maturity}" | tr -d '-' | grep -o -e "[0-9]\{1,2\}[wm]"
}

date_to_timestamp_nano() {
    date=$1
    year=$(echo ${date} | cut -d '-' -f3)
    month=$(echo ${date} | cut -d '-' -f1)
    day=$(echo ${date} | cut -d '-' -f2)
    time="12:00:00"
    date -d "${year}${month}${day} ${time}" -u +%s%N
}

date_normalized() {
    date=$1
    year=$(echo ${date} | cut -d '-' -f3)
    month=$(echo ${date} | cut -d '-' -f1)
    day=$(echo ${date} | cut -d '-' -f2)
    echo "${year}-${month}-${day}"
}

rm -f /tmp/euribor*
# As of November 1st 2013 the number of Euribor rates was reduced to 8 (1-2 weeks, 1, 2, 3, 6, 9 and 12 months).
for maturity in 1-week 2-weeks 1-month 2-months 3-months 6-months 9-months 12-months
do
    echo "processing maturity ${maturity}"
    fetch_page ${maturity}
    date=$(parse_date ${maturity})
    if [ ${#date} -ne 10 ]
    then
        echo "no date found, something might be wrong"
        continue
    fi
    rate=$(parse_rate $maturity)
    last_inserted=$(last_inserted_time ${maturity})
    if [ "${last_inserted}" == "${date}" ]
    then
        echo "data for ${date} already inserted"
        continue
    fi
    insert_data ${maturity} ${date} ${rate}
    populate_csv_files ${maturity} ${date} ${rate}
done
