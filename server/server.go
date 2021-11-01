package main

/***
* This file implements status server.
* Status server uses `Badger` structure to collect test-result information.
***/

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/smallkirby/skbctf-status/badge"
	"go.uber.org/zap"
)

func get_port() int {
	// priority is command-line > ENVVAR.
	var port_num int = 0
	port := flag.Int("port", 8080, "Port number this badge server listens to. (can be specified also by $BADGEPORT envvar.)")
	flag.Parse()

	// first, assign command-line value even if it's not specified
	port_num = *port

	// next, check envvar
	port_str := os.Getenv("BADGEPORT")
	if len(port_str) != 0 {
		if port_num_tmp, err := strconv.Atoi(port_str); err == nil {
			port_num = port_num_tmp
		}
	}

	// lastly, overwrite with command-line value if specified
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "port":
			port_num = *port
		}
	})

	return port_num
}

func main() {
	// force running in production mode
	gin.SetMode(gin.ReleaseMode)

	// init logger
	_logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to init logger.")
	}
	logger := _logger.Sugar()
	logger.Debug("Logger init.")

	// get Badger
	dbuser := os.Getenv("DBUSER")
	dbpass := os.Getenv("DBPASS")
	dbhost := os.Getenv("DBHOST")
	dbname := os.Getenv("DBNAME")
	badger, err := badge.NewBadger(dbuser, dbpass, dbhost, dbname)
	if err != nil {
		logger.Fatalf("%v", err)
	}

	// init server
	server := gin.Default()

	// ** set endpoints **

	// health check EP
	server.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// badge EP
	server.GET("/badge/:challid", func(c *gin.Context) {
		// convert parameter into int
		challid_str := c.Params.ByName("challid")
		challid, err := strconv.Atoi(challid_str)
		if err != nil {
			c.String(http.StatusBadRequest, "Specified challenge ID is invalid: %s.", challid_str)
			return
		}

		// fetch record
		badge_url, err := badger.GetBadge(challid)
		if err != nil {
			logger.Warnf("%v", err)
			c.String(http.StatusInternalServerError, "Something went to bad when fetching test result for %d.", challid)
			return
		}

		c.Header("Cache-Control", "max-age=60, public, immutable, must-revalidate")
		c.Redirect(http.StatusFound, badge_url)
		return
	})

	// default error badge
	server.GET("/badge/error", func(c *gin.Context) {
		c.Header("Cache-Control", "max-age=60, public, immutable, must-revalidate")
		c.Redirect(http.StatusFound, "https://img.shields.io/badge/error-status_fetching_fails-red")
	})

	// Run server
	port := get_port()
	port_str := fmt.Sprintf(":%v", port)
	logger.Infof("Badge server running on %s.", port_str)
	server.Run(port_str)
}
