package errno

import (
	"github.com/pkg/errors"
	"reflect"
)

type Error interface {
	error
}

type errorx struct {
	err error
}

func (e errorx) Error() string {
	return e.err.Error()
}

// New returns an error with the supplied message.
// New also records the stack trace at the point it was called.
func NewError(message string) Error {
	return &errorx{err: errors.New(message)}
}

// Errorf formats according to a format specifier and returns the string
// as a value that satisfies error.
// Errorf also records the stack trace at the point it was called.
func Errorf(format string, args ...interface{}) Error {
	return &errorx{err: errors.Errorf(format, args...)}
}

// WithStack annotates err with a stack trace at the point WithStack was called.
// If err is nil, WithStack returns nil.
func WithStack(err error) Error {
	if err == nil {
		return nil
	}
	// 如果是*Error类型,则已经包含了堆栈信息,因此直接返回.
	if Err, yes := reflect.ValueOf(err).Interface().(*errorx); yes {
		return Err
	}
	return &errorx{err: errors.WithStack(err)}
}

// Wrap returns an error annotating err with a stack trace
// at the point Wrap is called, and the supplied message.
// If err is nil, Wrap returns nil.
func Wrap(err error, message string) Error {
	if err == nil {
		return nil
	}
	if Err, yes := reflect.ValueOf(err).Interface().(*errorx); yes {
		err = Err.err
	}
	return &errorx{err: errors.Wrap(err, message)}
}

// Wrapf returns an error annotating err with a stack trace
// at the point Wrapf is called, and the format specifier.
// If err is nil, Wrapf returns nil.
func Wrapf(err error, format string, args ...interface{}) Error {
	if err == nil {
		return nil
	}
	if Err, yes := reflect.ValueOf(err).Interface().(*errorx); yes {
		err = Err.err
	}
	return &errorx{err: errors.Wrapf(err, format, args...)}
}

func Unwrap(err error) Error {
	if err == nil {
		return nil
	}
	if Err, yes := reflect.ValueOf(err).Interface().(*errorx); yes {
		err = Err.err
	}
	return &errorx{err: errors.Unwrap(err)}
}

func Cause(err error) Error {
	if err == nil {
		return nil
	}
	if Err, yes := reflect.ValueOf(err).Interface().(*errorx); yes {
		err = Err.err
	}
	return &errorx{err: errors.Cause(err)}
}
