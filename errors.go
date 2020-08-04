package retry

import "fmt"

type failed int8

var named = map[failed]string{
	noIssue:  "no-issue",
	exceeded: "exceeded",
	abort:    "aborted",
}

const (
	noIssue failed = iota
	exceeded
	abort
)

func (f failed) String() string {
	if s, ok := named[f]; ok {
		return s
	}
	return `unknown`
}

type failedRetries struct {
	msg  string
	fail failed
}

func (fr *failedRetries) Error() string {
	return fmt.Sprintf(`[retry:%v] %s`, fr.fail, fr.msg)
}

// ExceededRetries wraps the message passed and returns
// an error that be read by the error handler within the retry client.
func ExceededRetries(msg string) error {
	return &failedRetries{msg: msg, fail: exceeded}
}

// AbortedRetries wraps the message passed and returns
// an error that be read by the error handler within the retry client.
func AbortedRetries(msg string) error {
	return &failedRetries{msg: msg, fail: abort}
}

// HasAborted checks the error to validate
// if the executed function has notified that it should
// abort now.
func HasAborted(err error) bool {
	if e, ok := err.(*failedRetries); ok {
		return e.fail == abort
	}
	return false
}

// HasExceeded checks error to validate
// if the exectued function has notified that it exceeded the limit.
func HasExceeded(err error) bool {
	if e, ok := err.(*failedRetries); ok {
		return e.fail == exceeded
	}
	return false
}
