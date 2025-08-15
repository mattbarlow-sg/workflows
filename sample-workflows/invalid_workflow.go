package workflows

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"go.temporal.io/sdk/workflow"
)

// BadWorkflow has multiple validation issues
func BadWorkflow(ctx workflow.Context, input string) (string, error) {
	// ISSUE: Using time.Now() instead of workflow.Now()
	currentTime := time.Now()
	println("Current time:", currentTime.String())

	// ISSUE: Using math/rand without workflow.SideEffect
	randomNum := rand.Intn(100)
	println("Random number:", randomNum)

	// ISSUE: Using native goroutine instead of workflow.Go
	go func() {
		println("Background task running")
	}()

	// ISSUE: Using native channel instead of workflow.Channel
	ch := make(chan string)

	// ISSUE: Using select instead of workflow.Selector
	select {
	case msg := <-ch:
		println(msg)
	default:
		println("No message")
	}

	// ISSUE: Reading environment variable (non-deterministic)
	apiKey := os.Getenv("API_KEY")
	println("API Key:", apiKey)

	// ISSUE: Map iteration without sorting (non-deterministic order)
	data := map[string]int{"a": 1, "b": 2, "c": 3}
	for key, value := range data {
		println(key, value)
	}

	// Activity with issues
	var result string
	err := workflow.ExecuteActivity(ctx, bad_activity, input).Get(ctx, &result)
	if err != nil {
		return "", err
	}

	return result, nil
}

// ISSUE: Activity with bad naming (should be PascalCase)
func bad_activity(ctx context.Context, input string) (string, error) {
	return "result", nil
}

// ISSUE: Activity missing context parameter
func ProcessDataActivity(input string) (string, error) {
	return "processed", nil
}

// ISSUE: Activity missing error return
func NotifyUserActivity(ctx context.Context, userID string) string {
	return "notified"
}

// HumanApprovalWorkflow without proper escalation
func HumanApprovalWorkflow(ctx workflow.Context, requestID string) error {
	// ISSUE: Human task without escalation policy
	var approved bool
	signalChan := workflow.GetSignalChannel(ctx, "approval-signal")
	signalChan.Receive(ctx, &approved)

	if !approved {
		return fmt.Errorf("request %s was rejected", requestID)
	}

	return nil
}
