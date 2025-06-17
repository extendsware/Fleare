package errors

import (
	"fmt"
)

var (
	KeyNotFoundError      = "KeyNotFoundError"
	InvalidCommandError   = "InvalidCommandError"
	InvalidArgsError      = "InvalidArgumentError"
	InvalidValueError     = "InvalidValueError"
	InvalidKeyError       = "InvalidKeyError"
	InvalidMapKeyError    = "InvalidMapKeyError"
	InvalidCharacterError = "InvalidCharacterError"
	InvalidIndexError     = "InvalidIndexError"
	InvalidReferenceError = "InvalidReferenceError"
)

var (
	MsgInvalidCommandError = fmt.Sprintf("%v: The system could not process your request due to an unrecognized command input", InvalidCommandError)
)
