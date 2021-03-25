package utils

// CloseFunc is an io.Closer that can be easily constructed from a simple function.
type CloseFunc func() error

func (this CloseFunc) Close() error {
	return this()
}
