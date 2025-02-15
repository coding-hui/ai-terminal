package debug

// Option configuration function type
type Option func(*Logger) error

// WithFormatter sets log format
func WithFormatter(f Formatter) Option {
	return func(l *Logger) error {
		l.formatter = f
		return nil
	}
}

// WithBufferSize sets buffer size
func WithBufferSize(size int) Option {
	return func(l *Logger) error {
		l.bufferSize = size
		return nil
	}
}
