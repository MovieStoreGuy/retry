package retry

type recover struct {
	msg string
	can bool
}

func (r recover) Error() string {
	return r.msg
}

// CanRecover will check the error to if it is a NoRecover error
// and if it is, the internal values are exposed and returned back.
// Any other err is assumed to be recoverable.
func CanRecover(err error) (bool, error) {
	if r, ok := err.(recover); ok {
		return r.can, r
	}
	return true, err
}

// NoRecover returns a wrapped error that sets a flag saying
// it is not possible to recover from this error.
func NoRecover(msg string) error {
	return recover{msg: msg, can: false}
}
