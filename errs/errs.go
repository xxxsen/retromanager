package errs

import (
	"fmt"
	"retromanager/constants"
)

var (
	ErrOK = New(0, "success")
)

type IError interface {
	error
	Code() int64
	Message() string
}

type Error struct {
	code   int64
	msg    string
	err    error
	extmsg []string
}

func (e *Error) Error() string {
	return fmt.Sprintf("Error:[code:%d, msg:%s, err:[%v], extmsg:[%+v]]", e.code, e.msg, e.err, e.extmsg)
}

func (e *Error) Code() int64 {
	return e.code
}

func (e *Error) Message() string {
	return e.msg
}

func New(code int64, fmtter string, args ...interface{}) *Error {
	return Wrap(
		code,
		fmt.Sprintf(fmtter, args...),
		nil,
	)
}

func Wrap(code int64, msg string, err error) *Error {
	return &Error{
		code: code,
		msg:  msg,
		err:  err,
	}
}

func (e *Error) WithDebugMsg(fmtter string, args ...interface{}) *Error {
	e.extmsg = append(e.extmsg, fmt.Sprintf(fmtter, args...))
	return e
}

func IsErrOK(err IError) bool {
	if err == nil {
		return true
	}
	if err.Code() == 0 {
		return true
	}
	return false
}

func FromError(err error) IError {
	if err == nil {
		return nil
	}
	if e, ok := err.(IError); ok {
		return e
	}
	return Wrap(constants.ErrUnknown, "unknown error", err)
}
