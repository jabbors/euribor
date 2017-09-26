package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

const (
	apiKey = "to-be-invented"
)

var (
	errUnknownRetention = errors.New("unknown retention")
	errUnknownMaturity  = errors.New("unknown maturity")
	errMarshalError     = errors.New("marhshal failed")
	errInvalidYear      = errors.New("invalid year")
	errOutOfRange       = errors.New("time is out of range")
)

func sourceIP(r *http.Request) string {
	var ip string
	header := r.Header.Get("X-Forwarded-For")
	if header != "" {
		ips := strings.Split(header, ",")
		ip = strings.TrimSpace(ips[0])
	} else {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	return ip
}

// RequestHandler should be implemented by handlers that does not require authentication
type RequestHandler func(*http.Request, httprouter.Params) (string, int, error)

// NoAuthHandler wraps a handler with default stuff so each handler does
// not have to re-implement that same functionality
func NoAuthHandler(handler RequestHandler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		// Measure time spent executing shit
		start := time.Now()

		// Pass to the real handler
		response, statusCode, err := handler(r, params)

		// Logs [source IP] [request method] [request URL] [HTTP status] [time spent serving request]
		log.Printf("%v\t \"%v - %v\"\t%v\t%v", sourceIP(r), r.Method, r.RequestURI, statusCode, time.Since(start))

		if err != nil {
			http.Error(w, err.Error(), statusCode)
			return
		}

		w.WriteHeader(statusCode)
		fmt.Fprintln(w, response)
	}
}

// AuthorisedRequestHandler should be implemented by all handlers that require authentication
type AuthorisedRequestHandler func(*http.Request, httprouter.Params) (string, int, error)

// AuthHandler wraps a handler with default authorisation stuff so each handler
// does not have to re-implement the same functionality
func AuthHandler(handler AuthorisedRequestHandler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		// Measure time spent executing shit
		start := time.Now()

		// Authorize request
		err := authorize(r)
		if err != nil {
			// Logs [source IP] [request method] [request URL] [HTTP status] [time spent serving request]
			log.Printf("%v\t \"%v - %v\"\t%v\t%v", sourceIP(r), r.Method, r.RequestURI, http.StatusUnauthorized, time.Since(start))
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Pass to the real handler
		response, statusCode, err := handler(r, params)

		// Logs [source IP] [request method] [request URL] [HTTP status] [time spent serving request]
		log.Printf("%v\t \"%v - %v\"\t%v\t%v", sourceIP(r), r.Method, r.RequestURI, statusCode, time.Since(start))

		if err != nil {
			// If we run into an error, throw it back to the client (as plain text)
			http.Error(w, err.Error(), statusCode)
			return
		}

		w.WriteHeader(statusCode)
		fmt.Fprintln(w, response)
	}
}

func authorize(r *http.Request) error {
	token := r.Header.Get("X-User-Token")
	if token == "" {
		return fmt.Errorf("Missing auth token")
	}

	if token != apiKey {
		return fmt.Errorf("Invalid token")
	}

	return nil
}

func versionMsg(msg string) string {
	return fmt.Sprintf("{\"version\":\"%s\"}", msg)
}

// index handler
func indexHandler(r *http.Request, _ httprouter.Params) (string, int, error) {
	return "Welcome to the Euribor rates service!", http.StatusOK, nil
}

// version handler
func versionHandler(r *http.Request, _ httprouter.Params) (string, int, error) {
	return versionMsg(version), http.StatusOK, nil
}

// influx handler
func influxHandler(r *http.Request, params httprouter.Params) (string, int, error) {
	retention := params.ByName("retention")
	if isValidRetention(retention) == false {
		return "", http.StatusBadRequest, errUnknownRetention
	}

	maturity := params.ByName("maturity")
	if isValidMaturity(maturity) == false {
		return "", http.StatusBadRequest, errUnknownMaturity
	}

	rates := []rate{}
	for _, r := range influxCache[retentionMap[retention]][maturity] {
		rates = append(rates, r)
	}

	jsonData, err := json.Marshal(rates)
	if err != nil {
		return "", http.StatusInternalServerError, errMarshalError
	}

	return string(jsonData), http.StatusOK, nil
}

// highstock handler
func highstockHandler(r *http.Request, params httprouter.Params) (string, int, error) {
	maturity := params.ByName("maturity")
	if isValidMaturity(maturity) == false {
		return "", http.StatusBadRequest, errUnknownMaturity
	}

	rates := rateSlice{}
	for _, r := range historyCache[maturity] {
		rates = append(rates, r)
	}

	return rates.Values(), http.StatusOK, nil
}

// history handler
func historyHandler(r *http.Request, params httprouter.Params) (string, int, error) {
	year, err := strconv.ParseInt(params.ByName("year"), 10, 32)
	if err != nil {
		return "", http.StatusBadRequest, errInvalidYear
	}
	if year < 2010 || year > int64(time.Now().Year()) {
		return "", http.StatusBadRequest, errOutOfRange
	}

	maturity := params.ByName("maturity")
	if isValidMaturity(maturity) == false {
		return "", http.StatusBadRequest, errUnknownMaturity
	}

	rates := []rate{}
	for _, r := range historyCache[maturity] {
		if int64(r.Date.Year()) != year {
			continue
		}
		rates = append(rates, r)
	}

	jsonData, err := json.Marshal(rates)
	if err != nil {
		return "", http.StatusInternalServerError, errMarshalError
	}

	return string(jsonData), http.StatusOK, nil
}

// webapp handler
func webappHandler(r *http.Request, params httprouter.Params) (string, int, error) {
	return renderWebapp(webRoot), http.StatusOK, nil
}
