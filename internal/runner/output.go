package runner

import "fmt"

type Output struct {
	error          error
	errorMessage   string
	successMessage string
}

func NewRunOutput(error error, errorMessage string, successMessage string) Output {
	return Output{
		error:          error,
		errorMessage:   errorMessage,
		successMessage: successMessage,
	}
}

func (o Output) HasError() bool {
	return o.error != nil
}

func (o Output) GetErrorMessage() string {
	return fmt.Sprintf("%s: %s", o.errorMessage, o.error)
}

func (o Output) GetSuccessMessage() string {
	return o.successMessage
}
