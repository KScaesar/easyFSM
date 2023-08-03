package easyFSM

import "fmt"

var (
	ErrStateNotMatch = fmt.Errorf("state not match")
	ErrEventNotExist = fmt.Errorf("event not exist")
)
