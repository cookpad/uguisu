package uguisu

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/getsentry/sentry-go"
)

func init() {
	dsn := strings.TrimSpace(os.Getenv("SENTRY_DSN"))
	if dsn == "" {
		return
	}
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		AttachStacktrace: true,
	}); err != nil {
		log.Printf("uguisu: sentry.Init: %v", err)
	}
}

func flushSentry() {
	sentry.Flush(2 * time.Second)
}

// captureSentryError sends err to Sentry when SENTRY_DSN is configured. Lambda context is attached when available.
func captureSentryError(ctx context.Context, err error) *sentry.EventID {
	if err == nil {
		return nil
	}
	var id *sentry.EventID
	sentry.WithScope(func(scope *sentry.Scope) {
		if lc, ok := lambdacontext.FromContext(ctx); ok {
			scope.SetTag("lambda_request_id", lc.AwsRequestID)
			scope.SetContext("lambda", map[string]interface{}{
				"request_id":           lc.AwsRequestID,
				"invoked_function_arn": lc.InvokedFunctionArn,
			})
		}
		id = sentry.CaptureException(err)
	})
	return id
}
