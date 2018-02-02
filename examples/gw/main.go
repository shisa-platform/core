package main

import (
	"flag"
	"log"

	"go.uber.org/zap"
)

const (
	goodbyeServiceAddrEnv = "GOODBYE_SERVICE_ADDR"
	idpServiceAddrEnv     = "IDP_SERVICE_ADDR"
)

func main() {
	addr := flag.String("addr", "", "gateway service address")
	debugAddr := flag.String("debugaddr", ":0", "debug service address")
	healthcheckAddr := flag.String("healthcheckaddr", ":0", "healthcheck service address")

	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("error initializing logger: %v", err)
	}

	defer logger.Sync()

	serve(logger, *addr, *debugAddr, *healthcheckAddr)
}
