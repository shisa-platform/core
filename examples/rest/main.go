package main

import (
	"expvar"
	"flag"
	"log"
	"os"
	"time"

	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"
)

const (
	timeFormat  = "2006-01-02T15:04:05+00:00"
	serviceName = "rest"
)

var goodbye = expvar.NewMap(serviceName)

func main() {
	start := time.Now().UTC()

	startTime := new(expvar.String)
	startTime.Set(start.Format(timeFormat))
	goodbye.Set("start-time", startTime)
	goodbye.Set("uptime", expvar.Func(func() interface{} {
		now := time.Now().UTC()
		return now.Sub(start).String()
	}))

	addr := flag.String("addr", ":0", "service address")
	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("initializing logger: %s", err)
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

	serve(logger, *addr)
}
