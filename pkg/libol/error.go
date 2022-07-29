package libol

import "fmt"

func NewErr(message string, v ...interface{}) error {
	return fmt.Errorf(message, v...)
}
