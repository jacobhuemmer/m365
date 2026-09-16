package domain

import "fmt"

const (
	ExitOK       = 0
	ExitUsage    = 3
	ExitAuth     = 4
	ExitService  = 5
	ExitNotFound = 6
)

const (
	ClassUsage    = "usage"
	ClassAuth     = "auth"
	ClassService  = "service"
	ClassNotFound = "not_found"
)

type Error struct {
	Class   string
	Message string
	Hint    string
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *Error) ExitCode() int {
	if e == nil {
		return ExitOK
	}
	switch e.Class {
	case ClassUsage:
		return ExitUsage
	case ClassAuth:
		return ExitAuth
	case ClassService:
		return ExitService
	case ClassNotFound:
		return ExitNotFound
	default:
		return ExitUsage
	}
}

func Usage(msg string) *Error    { return &Error{Class: ClassUsage, Message: msg} }
func Auth(msg string) *Error     { return &Error{Class: ClassAuth, Message: msg} }
func Service(msg string) *Error  { return &Error{Class: ClassService, Message: msg} }
func NotFound(msg string) *Error { return &Error{Class: ClassNotFound, Message: msg} }

func Usagef(format string, args ...any) *Error { return Usage(fmt.Sprintf(format, args...)) }

func ExitOf(err error) int {
	if err == nil {
		return ExitOK
	}
	if de, ok := err.(*Error); ok {
		return de.ExitCode()
	}
	return ExitUsage
}

func ClassOf(err error) string {
	if de, ok := err.(*Error); ok {
		return de.Class
	}
	if err == nil {
		return ""
	}
	return ClassUsage
}
