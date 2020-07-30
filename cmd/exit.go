package cmd

import (
	"fmt"
	"os"
	"strings"
)

const (
	CodeFailed = 1
	// 2: reserved
	CodeInvalidArgs = 3
	// 4: CodeInvalidEnv
	// 5: CodeNotFound
	CodeFailedDetect           = 100
	CodeFailedDetectWithErrors = 101
	CodeDetectError            = 102
	CodeFailedBuild            = 7
	CodeFailedLaunch           = 8
	// 9: CodeFailedUpdate
	CodeFailedSave               = 10
	CodeIncompatiblePlatformAPI  = 11
	CodeIncompatibleBuildpackAPI = 12
)

type ErrorFail struct {
	Err    error
	Code   int
	Action []string
}

func (e *ErrorFail) Error() string {
	message := "failed to " + strings.Join(e.Action, " ")
	if e.Err == nil {
		return message
	}
	return fmt.Sprintf("%s: %s", message, e.Err)
}

func FailCode(code int, action ...string) *ErrorFail {
	return FailErrCode(nil, code, action...)
}

func FailErr(err error, action ...string) *ErrorFail {
	code := CodeFailed
	if err, ok := err.(*ErrorFail); ok {
		code = err.Code
	}
	return FailErrCode(err, code, action...)
}

func FailErrCode(err error, code int, action ...string) *ErrorFail {
	return &ErrorFail{Err: err, Code: code, Action: action}
}

func Exit(err error) {
	if err == nil {
		os.Exit(0)
	}
	DefaultLogger.Errorf("%s\n", err)
	if err, ok := err.(*ErrorFail); ok {
		os.Exit(err.Code)
	}
	os.Exit(CodeFailed)
}

func ExitWithVersion() {
	DefaultLogger.Infof(buildVersion())
	os.Exit(0)
}
