package errs

type notFoundError struct {
	name string
}

func (e notFoundError) Error() string {
	return "не удалось найти " + e.name
}

func ErrNotFound(name string) error {
	err := notFoundError{name}

	return err
}

type invalidFormatError struct {
	name string
}

func (e invalidFormatError) Error() string {
	return "неверный формат: " + e.name
}

func ErrInvalidFormat(name string) error {
	err := invalidFormatError{name}

	return err
}
