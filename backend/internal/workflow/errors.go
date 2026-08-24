package workflow

import "fmt"

type GraphValidationError struct {
	Errors []ValidationError
}

func (e *GraphValidationError) Error() string {
	return fmt.Sprintf("invalid workflow graph: %d errors(s)", len(e.Errors))
}
