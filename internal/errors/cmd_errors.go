package errors

import (
	"errors"
	"fmt"
)

var (
	InvalidCommandError   = errors.New("InvalidCommandError")
	InvalidArgsError      = errors.New("InvalidArgumentError")
	InvalidValueError     = errors.New("InvalidValueError")
	InvalidKeyError       = errors.New("InvalidKeyError")
	InvalidMapKeyError    = errors.New("InvalidMapKeyError")
	InvalidCharacterError = errors.New("InvalidCharacterError")
)

var (
	MsgInvalidCommandError = fmt.Sprintf("%v: The system could not process your request due to an unrecognized command input", InvalidCommandError)
)
