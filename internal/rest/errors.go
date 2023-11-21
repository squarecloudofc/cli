package rest

import (
	"errors"
	"fmt"
)

var (
	ErrAppNotFound   = errors.New("application not found, verify your application ID and try again")
	ErrAccessDenied  = errors.New("your access is denied, make login with your api token using \"squarecloud login\" or verify if you have access for this action")
	ErrUserNotFound  = errors.New("user not found, verify your user ID and try again")
	ErrInvalidBuffer = errors.New("unable to send buffer")
	ErrInvalidFile   = errors.New("unable to send the file")
	ErrCommitError   = errors.New("unable to commit to your application")
	ErrDelayNow      = errors.New("you are in rate limit, try again later")
)

func ParseError[T any](e *ApiResponse[T]) (err error) {
	switch e.Code {
	case "APP_NOT_FOUND":
		err = ErrAppNotFound
	case "USER_NOT_FOUND":
		err = ErrUserNotFound
	case "ACCESS_DENIED":
		err = ErrAccessDenied
	case "INVALID_FILE":
		err = ErrInvalidFile
	case "INVALID_BUFFER":
		err = ErrInvalidBuffer
	case "COMMIT_ERROR":
		err = ErrCommitError
	case "DELAY_NOW":
		err = ErrDelayNow
	default:
		err = fmt.Errorf("square cloud retorned error %s", e.Code)
	}

	return
}
