package tracer

import (
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/shiyongxi/go-common/logger"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/zipkin"
)

type (
	TraceConfig struct {
		Param       float64
		HostPort    string
		ServiceName string
		LogSpans    bool
	}
)

var (
	tracerClient opentracing.Tracer
)

func GetTracerClient() opentracing.Tracer {
	return tracerClient
}

func NewTracer(cfg *TraceConfig) opentracing.Tracer {
	configEnv, err := config.FromEnv()
	if err != nil {
		logger.Error(err)
	}

	if cfg.HostPort == "" {
		cfg.HostPort = configEnv.Reporter.LocalAgentHostPort
	}

	if cfg.Param <= 0 {
		cfg.Param = configEnv.Sampler.Param
	}

	traceCfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: cfg.Param,
		},

		Reporter: &config.ReporterConfig{
			LogSpans:           cfg.LogSpans,
			LocalAgentHostPort: cfg.HostPort,
		},

		ServiceName: cfg.ServiceName,
	}

	propagator := zipkin.NewZipkinB3HTTPHeaderPropagator()

	tracerClient, _, err = traceCfg.NewTracer(
		config.Logger(jaeger.StdLogger),
		config.Injector(opentracing.HTTPHeaders, propagator),
		config.Extractor(opentracing.HTTPHeaders, propagator),
		config.ZipkinSharedRPCSpan(true),
		config.MaxTagValueLength(256),
		config.PoolSpans(true),
	)
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}

	return tracerClient
}
