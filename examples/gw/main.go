package main

import (
	"expvar"
	"flag"
	"log"
	"time"

	"go.uber.org/zap"
)

const (
	timeFormat            = "2006-01-02T15:04:05+00:00"
	goodbyeServiceAddrEnv = "GOODBYE_SERVICE_ADDR"
	idpServiceAddrEnv     = "IDP_SERVICE_ADDR"
	gateway               = expvar.NewMap("gateway")
)

func main() {
	now := time.Now().UTC().Format(timeFormat)
	startTime := new(expvar.String)
	startTime.Set(now)
	gateway.Set("start-time", startTime)

	addr := flag.String("addr", "", "gateway service address")
	debugAddr := flag.String("debugaddr", ":0", "debug service address")
	healthcheckAddr := flag.String("healthcheckaddr", ":0", "healthcheck service address")

	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("error initializing logger: %v", err)
	}

	defer logger.Sync()

	serve(addr, debugAddr, healthcheckAddr)
}
