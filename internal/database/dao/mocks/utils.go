package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// handleMockReturn is a utility function to handle mock return values that can be either
// direct values or functions that need to be invoked with the original arguments.
// This reduces code duplication across mock DAOs.
func handleMockReturn[T any](args mock.Arguments, ctx context.Context, fnArgs ...interface{}) (T, error) {
	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, ...interface{}) (T, error)); ok {
		return fn(ctx, fnArgs...)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		var zero T
		return zero, args.Error(1)
	}
	return args.Get(0).(T), args.Error(1)
}

// invokeIfFunction checks if a mock return value is a function and invokes it if so
func invokeIfFunction[T any](returnValue interface{}, ctx context.Context, args ...interface{}) (T, error) {
	if fn, ok := returnValue.(func(context.Context, ...interface{}) (T, error)); ok {
		return fn(ctx, args...)
	}

	var zero T
	return zero, nil
}
