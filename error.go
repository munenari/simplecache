package simplecache

import (
	"fmt"
)

var ErrNotFound = fmt.Errorf("cache: not found")

type ErrInvalidType struct {
	Got      any
	Expected any
}

func (x *ErrInvalidType) Error() string {
	return fmt.Sprintf("cache: got [%T] but not [%T]", x.Got, x.Expected)
}

var _ error = &ErrInvalidType{}
