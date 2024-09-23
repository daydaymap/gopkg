package errs

import (
	"errors"
	"fmt"
	"io"
)

var (
	Success = "success"

	RetOk      int32 = 200
	RetUnknown int32 = 99999
)

type Error struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`

	cause error
	stack stackTrace
}

func New(code int32, format string, params ...interface{}) error {
	msg := fmt.Sprintf(format, params...)
	err := &Error{
		Code: code,
		Msg:  msg,
	}
	if traceable {
		err.stack = callers()
	}
	return err
}

func (e *Error) Error() string {
	if e == nil {
		return Success
	}

	if e.cause != nil {
		return fmt.Sprintf("code:%d, msg:%s, caused by %s",
			e.Code, e.Msg, e.cause.Error())
	}
	return fmt.Sprintf("code: %d, msg: %s", e.Code, e.Msg)
}

func (e *Error) Format(s fmt.State, verb rune) {
	var stackTrace stackTrace
	defer func() {
		if stackTrace != nil {
			stackTrace.Format(s, verb)
		}
	}()

	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = fmt.Fprintf(s, "code: %d, msg: %s", e.Code, e.Msg)
			if e.stack != nil {
				stackTrace = e.stack
			}
			if e.Unwrap() != nil {
				_, _ = fmt.Fprintf(s, "\nCause by %+v", e.Unwrap())
			}
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, e.Error())
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", e.Error())
	default:
		_, _ = fmt.Fprintf(s, "%%!%c(errs.Error=%s)", verb, e.Error())
	}
}

func (e *Error) Unwrap() error {
	return e.cause
}

func Wrap(err error, code int32, msg string) error {
	if err == nil {
		return nil
	}
	wrapErr := &Error{
		Code:  code,
		Msg:   msg,
		cause: err,
	}
	var e *Error
	if traceable && !errors.As(err, &e) {
		wrapErr.stack = callers()
	}
	return wrapErr
}

func Wrapf(err error, code int32, format string, params ...interface{}) error {
	if err == nil {
		return nil
	}
	msg := fmt.Sprintf(format, params...)
	wrapErr := &Error{
		Code:  code,
		Msg:   msg,
		cause: err,
	}
	var e *Error
	if traceable && !errors.As(err, &e) {
		wrapErr.stack = callers()
	}
	return wrapErr
}

func Code(e error) int32 {
	if e == nil {
		return RetOk
	}
	err, ok := e.(*Error)
	if !ok && !errors.As(e, &err) {
		return RetUnknown
	}
	if err == nil {
		return RetOk
	}
	return err.Code
}

func Msg(e error) string {
	if e == nil {
		return Success
	}
	err, ok := e.(*Error)
	if !ok && !errors.As(e, &err) {
		return Success
	}
	if err.Unwrap() != nil {
		return fmt.Sprintf("%s: %s", err.Msg, err.Error())
	}
	return err.Msg
}
