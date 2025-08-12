package entity

// custom error

type InvalidArgumentError struct {
	message string
}

func NewInvalidArgumentError(message string) error {
	return &InvalidArgumentError{message}
}
func (e *InvalidArgumentError) Error() string {
	return e.message
}

type ResourceAlreadyExistsError struct {
	message string
}

func NewResourceAlreadyExistsError(message string) error {
	return &ResourceAlreadyExistsError{message}
}
func (e *ResourceAlreadyExistsError) Error() string {
	return e.message
}
