package error

type LError struct {
	err     error
	ErrFunc ErrorFunc
}

func Default() *LError {
	return &LError{}
}

func (e *LError) Error() string {
	return e.err.Error()
}

func (e *LError) Put(err error) {
	e.check(err)
}

func (e *LError) check(err error) {
	if err != nil {
		e.err = err
		panic(e)
	}
}

type ErrorFunc func(e *LError)

func (e *LError) Result(errFunc ErrorFunc) {
	e.ErrFunc = errFunc
}

func (e *LError) ExecResult() {
	e.ErrFunc(e)
}
