package metrics

import (
	"github.com/TheZeroSlave/zapsentry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	DSN         = "https://94c0444acd3e4507a29b5e9b4c59403d@o1428566.ingest.sentry.io/6778904"
	TRACER_NAME = "team1_auth"
)

var Tracer = otel.Tracer(TRACER_NAME)

func InitOtel() error {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(
		jaeger.WithEndpoint("http://jaeger-instance-collector.observability:14268/api/traces")),
	)
	if err != nil {
		return err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("demo_service_name"),
		)))

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return nil
}

func GetLogger(debug bool, dsn string, env string) (*zap.Logger, error) {
	var err error
	var l *zap.Logger

	if debug {
		l, err = zap.NewDevelopment()
	} else {
		l, err = zap.NewProduction()
	}

	l = initSentry(l, dsn, env)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = l.Sync()
	}()
	l.Debug("logger initialized in debug level")

	return l, err
}

func initSentry(log *zap.Logger, sentryAddress, environment string) *zap.Logger {
	if sentryAddress == "" {
		return log
	}

	cfg := zapsentry.Configuration{
		Level: zapcore.ErrorLevel,
		Tags: map[string]string{
			"environment": environment,
			"app":         "demoApp",
		},
	}

	core, err := zapsentry.NewCore(cfg, zapsentry.NewSentryClientFromDSN(sentryAddress))
	if err != nil {
		log.Warn("failed to init zap", zap.Error(err))
	}

	return zapsentry.AttachCoreToLogger(core, log)
}
