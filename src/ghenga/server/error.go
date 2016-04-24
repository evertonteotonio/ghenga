package server

// Error bundles an HTTP status code and an error.
type Error interface {
	error
	Status() int
}

// StatusError bundles an HTTP status code with an error.
type StatusError struct {
	Err  error
	Code int
}

// Status returns the HTTP status for this error
func (err StatusError) Status() int {
	return err.Code
}

func (err StatusError) Error() string {
	return err.Err.Error()
}
