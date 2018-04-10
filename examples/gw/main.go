package main

import (
	"flag"
	"log"
	"os"

	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"
)

const (
	serviceName = "gateway"
)

func main() {
	addr := flag.String("addr", ":0", "gateway service address")
	debugAddr := flag.String("debugaddr", ":0", "debug service address")
	healthcheckAddr := flag.String("healthcheckaddr", ":0", "healthcheck service address")

	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("error initializing logger: %s", err.Error())
	}

	defer logger.Sync()

	cfg := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
	}

	if value := os.Getenv("JAEGER_AGENT_HOST"); value != "" {
		cfg.Reporter = &jaegercfg.ReporterConfig{
			LocalAgentHostPort: value + ":6831",
		}
	}

	closer, err := cfg.InitGlobalTracer(serviceName)
	if err != nil {
		log.Fatalf("error initializing jaeger tracer: %s", err.Error())
	}
	defer closer.Close()

	serve(logger, *addr, *debugAddr, *healthcheckAddr)
}
