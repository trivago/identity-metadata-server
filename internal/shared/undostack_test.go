package shared

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUndoStack(t *testing.T) {
	assert := assert.New(t)

	stack := NewUndoStack()
	assert.NotNil(stack, "NewUndoStack should return a non-nil UndoStack")

	callOrder := []int{}

	func1Called := false
	func1 := func() error {
		func1Called = true
		callOrder = append(callOrder, 1)
		return nil
	}
	func2Called := false
	func2 := func() error {
		func2Called = true
		callOrder = append(callOrder, 2)
		return nil
	}
	func3Called := false
	func3 := func() error {
		func3Called = true
		callOrder = append(callOrder, 3)
		return nil
	}

	stack.Push(func1)
	stack.Push(func2)
	stack.Push(func3)
	assert.Equal(3, len(*stack), "UndoStack should have 3 items after pushing 3 functions")

	undoFunc, ok := stack.Pop()
	assert.True(ok, "Pop should return true when stack is not empty")
	assert.NotNil(undoFunc, "Pop should return a non-nil function")
	assert.False(func3Called, "func3 should not have been called yet")

	err := stack.Rollback()
	assert.NoError(err, "Rollback should not return an error")
	assert.False(func3Called, "func3 should not have been called during rollback")
	assert.True(func2Called, "func2 should have been called during rollback")
	assert.True(func1Called, "func1 should have been called during rollback")

	assert.Equal([]int{2, 1}, callOrder, "Call order should be 2, 1 after rollback")

	assert.True(stack.IsEmpty(), "UndoStack should be empty after rollback")
}

func TestUndStackErrorHandling(t *testing.T) {
	assert := assert.New(t)

	stack := NewUndoStack()

	// Test rollback with no functions
	err := stack.Rollback()
	assert.NoError(err, "Rollback should not return an error when stack is empty")

	funcWithoutErrorCalled := false
	funcWithoutError := func() error {
		funcWithoutErrorCalled = true
		return nil
	}

	// Push a function that returns an error
	funcWithError := func() error {
		return errors.New("test error")
	}

	stack.Push(funcWithError)
	stack.Push(funcWithoutError)

	// Rollback should return the error from the function
	err = stack.Rollback()
	assert.Error(err, "Rollback should return an error from the undo function")
	assert.Equal("test error", err.Error(), "Error message should match the expected error")
	assert.True(funcWithoutErrorCalled, "funcWithoutError should have been called during rollback")
	assert.True(stack.IsEmpty(), "UndoStack should be empty after rollback with error")

	// Rollback should call all functions, even if one returns an error
	funcWithoutErrorCalled = false
	stack.Push(funcWithoutError)
	stack.Push(funcWithError)
	err = stack.Rollback()

	assert.Error(err, "Rollback should return an error from the undo function")
	assert.Equal("test error", err.Error(), "Error message should match the expected error")
	assert.True(funcWithoutErrorCalled, "funcWithoutError should have been called during rollback")
	assert.True(stack.IsEmpty(), "UndoStack should be empty after rollback with error")

	// Test error collection when multiple functions return errors
	stack.Clear() // Clear the stack before the next test

	stack.Push(funcWithError)
	stack.Push(funcWithError) // Push the same error function twice
	stack.Push(funcWithoutError)

	err = stack.Rollback()
	assert.Error(err, "Rollback should return an error from the undo functions")
	assert.Equal("test error\ntest error", err.Error(), "Error message should contain both errors")
	assert.True(funcWithoutErrorCalled, "funcWithoutError should have been called during rollback")
	assert.True(stack.IsEmpty(), "UndoStack should be empty after rollback with multiple errors")

	// Test RollbackIfError when passing a nil error (no rollback should occur)
	stack.Clear() // Clear the stack before the next test

	funcWithoutErrorCalled = false
	stack.Push(funcWithoutError)

	err = stack.RollbackIfError(nil)
	assert.NoError(err, "RollbackIfError should not return an error when no error is provided")
	assert.False(stack.IsEmpty(), "UndoStack should not be empty after RollbackIfError with nil error")
	assert.False(funcWithoutErrorCalled, "funcWithoutError should not have been called during RollbackIfError with nil error")

	// Test RollbackFromError when passing a nil error (rollback should still work)
	err = stack.RollbackFromError(nil)
	assert.NoError(err, "RollbackFromError should not return an error when no error is provided and no function returns an error")
	assert.True(stack.IsEmpty(), "UndoStack should be empty after RollbackFromError with nil error")
	assert.True(funcWithoutErrorCalled, "funcWithoutError should have been called during RollbackFromError with nil error")
}
