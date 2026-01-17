package libol

import "fmt"

func NewErr(message string, v ...any) error {
	return fmt.Errorf(message, v...)
}
