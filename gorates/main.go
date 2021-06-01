package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/kelseyhightower/envconfig"
)

const (
	appName    = "gorates"
	timeFormat = "2006-01-02"
)

var (
	version string
)

// appConfig represents the configuration.
type appConfig struct {
	Host            string `default:"0.0.0.0" desc:"host to bind to"`
	Port            int    `default:"8080" desc:"port to bind to"`
	WebRoot         string `default:"" split_words:"true" desc:"root path of hosted behind a proxy"`
	DataPath        string `default:"." split_words:"true" desc:"path to data CSV files"`
	PushpulletToken string `split_words:"true" desc:"authorization token for pushbullet used when sending alerts"`
}

// parse options from the environment. Return an error if parsing fails.
func (a *appConfig) parse() {
	defaultUsage := flag.Usage
	flag.Usage = func() {
		// Show default usage for the app (lists flags, etc).
		defaultUsage()
		fmt.Fprint(os.Stderr, "\n")

		err := envconfig.Usage("", a)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n\n", err.Error())
			os.Exit(1)
		}
	}

	var verFlag bool
	flag.BoolVar(&verFlag, "version", false, "print version and exit")
	flag.Parse()

	// Print version and exit if -version flag is passed.
	if verFlag {
		fmt.Printf("gorates: version=%s\n", version)
		os.Exit(0)
	}

	err := envconfig.Process("", a)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n\n", err.Error())
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	config := &appConfig{}
	config.parse()

	monitorCh := make(chan bool)

	_ = NewMonitorService(config.PushpulletToken, monitorCh)
	_ = NewCacheService(config.DataPath, monitorCh)
	h := NewHandler(config.WebRoot)

	router := httprouter.New()
	router.RedirectTrailingSlash = true

	// routes for general info
	// TODO: add routes for list of supported retentions/maturities
	router.GET("/", NoAuthHandler(h.IndexHandler))
	router.GET("/version", NoAuthHandler(h.VersionHandler))

	// routes to serve the app
	router.GET("/rates/app/hs/:maturity", NoAuthHandler(h.HighstockHandler))

	// routes to serve historical queries
	router.GET("/rates/history/:year/:maturity", NoAuthHandler(h.HistoryHandler))

	// routes to serve the webapp
	router.GET("/webapp", NoAuthHandler(h.WebappHandler))

	// routes to manage alerts
	router.PUT("/alert/:email/:maturity/:limit", NoAuthHandler(h.AlertAddHandler))
	router.DELETE("/alert/:email/:maturity/:limit", NoAuthHandler(h.AlertRemoveHandler))
	router.GET("/alert/:email", NoAuthHandler(h.AlertListHandler))

	log.Printf("listening on %s:%d", config.Host, config.Port)
	log.Fatal(http.ListenAndServe(config.Host+":"+strconv.Itoa(config.Port), router))
}
