package util

import (
	"bytes"
	"fmt"
)

type Errors []error

func HasError(errs []error) bool {
	for _, e := range errs {
		if e != nil {
			return true
		}
	}
	return false
}

func MultiError(errs []error) error {
	var buf bytes.Buffer

	for i, err := range errs {
		if err != nil {
			if i != 0 {
				buf.WriteString("; ")
			}
			buf.WriteString(err.Error())
		}
	}

	return fmt.Errorf(buf.String())
}
