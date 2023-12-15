package easyFSM

import "fmt"

var (
	ErrStateNotMatch   = fmt.Errorf("state not match")
	ErrEventNotDefined = fmt.Errorf("event not defined")
)
