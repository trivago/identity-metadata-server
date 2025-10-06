package shared

import "errors"

type UndoStack []func() error

// NewUndoStack creates a new UndoStack.
func NewUndoStack() *UndoStack {
	return &UndoStack{}
}

// Push adds a new undo function to the end of the stack.
// The undo function should return an error if it fails to undo an operation.
func (us *UndoStack) Push(undoFunc func() error) {
	*us = append(*us, undoFunc)
}

// Pop removes the last undo function from the stack and returns it.
// It returns false if the stack is empty.
func (us *UndoStack) Pop() (func() error, bool) {
	stackSize := len(*us)
	if stackSize == 0 {
		return nil, false
	}
	// Pop the last undo function
	undoFunc := (*us)[stackSize-1]
	*us = (*us)[:stackSize-1]
	return undoFunc, true
}

// IsEmpty checks if the undo stack is empty.
func (us *UndoStack) IsEmpty() bool {
	return len(*us) == 0
}

// Clear removes all undo functions from the stack without executing them.
func (us *UndoStack) Clear() {
	// Clear the stack by reinitializing it
	*us = UndoStack{}
}

// Rollback calls RollbackFromError with a nil error.
func (us *UndoStack) Rollback() error {
	return us.RollbackFromError(nil)
}

// RollbackIfError works like RollbackFromError, but only calls it if the
// provided error is not nil.
func (us *UndoStack) RollbackIfError(err error) error {
	if err != nil {
		return us.RollbackFromError(err)
	}
	return nil
}

// RollbackFromError executes all undo functions with the last one executed first.
// Errors returned by the undo functions will be collected. All undo functions will
// be called. The originial error is always returned first, followed by any errors
// from the undo functions.
// The stack is cleared after the rollback is complete, regardless of success
// or failure of single undo functions.
// Passing a nil error will not change the behavior, it will still execute all undo
// functions. If a nil error should not execute the undo functions, use
// RollbackIfError instead.
func (us *UndoStack) RollbackFromError(err error) error {
	for {
		undo, ok := us.Pop()
		switch {
		case !ok:
			return err
		case undo == nil:
			continue
		default:
			if undoErr := undo(); undoErr != nil {
				err = errors.Join(err, undoErr)
			}
		}
	}
}
