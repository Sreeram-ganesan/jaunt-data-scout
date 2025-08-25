package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, event map[string]any) (map[string]any, error) {
	state := getenv("STATE_NAME", "mock")

	payload := map[string]any{
		"state":  state,
		"status": "ok",
	}

	// items: pass through up to 5 items if present
	if v, ok := event["items"]; ok {
		if arr, ok := v.([]any); ok {
			n := len(arr)
			if n > 5 {
				n = 5
			}
			payload["items"] = arr[:n]
		} else {
			payload["items"] = []any{}
		}
	} else {
		payload["items"] = []any{}
	}

	// new_unique_rate: default to 0.2 if not present or invalid
	if v, ok := event["new_unique_rate"]; ok {
		switch t := v.(type) {
		case float64:
			payload["new_unique_rate"] = t
		case float32:
			payload["new_unique_rate"] = float64(t)
		case int:
			payload["new_unique_rate"] = float64(t)
		case int32:
			payload["new_unique_rate"] = float64(t)
		case int64:
			payload["new_unique_rate"] = float64(t)
		default:
			payload["new_unique_rate"] = 0.2
		}
	} else {
		payload["new_unique_rate"] = 0.2
	}

	// Preserve job context
	for _, k := range []string{"job_id", "city", "s3_prefix", "budgets", "kill_switches", "early_stop", "timeouts"} {
		if v, ok := event[k]; ok {
			payload[k] = v
		}
	}

	log.Printf("mock handler '%s' processed event", state)
	return payload, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	lambda.Start(handler)
}
