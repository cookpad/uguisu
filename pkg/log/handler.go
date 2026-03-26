package log

import (
	"context"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

func Handler(ctx context.Context) slog.Handler {
	opts := &slog.HandlerOptions{}
	opts.Level = getLogLevel()

	h := slog.NewJSONHandler(os.Stdout, opts)
	if lc, ok := lambdacontext.FromContext(ctx); ok {
		return h.WithAttrs([]slog.Attr{slog.String("request_id", lc.AwsRequestID)})
	}
	return h
}

func getLogLevel() slog.Leveler {
	str, found := os.LookupEnv("LOG_LEVEL")

	// If no value is set, use Info as default Level
	if !found {
		return slog.LevelInfo
	}

	var l slog.Level
	err := l.UnmarshalText([]byte(str))

	// If invalid value is set, use Info as default Level
	if err != nil {
		return slog.LevelInfo
	}

	return l
}
