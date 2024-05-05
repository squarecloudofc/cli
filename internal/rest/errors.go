package rest

import (
	"errors"
	"fmt"
)

type RestError error

var (
	ErrAppNotFound   RestError = errors.New("application not found, verify your application ID and try again")
	ErrAccessDenied  RestError = errors.New("your access is denied, make login with your api token using \"squarecloud login\" or verify if you have access for this action")
	ErrUserNotFound  RestError = errors.New("user not found, verify your user ID and try again")
	ErrInvalidBuffer RestError = errors.New("unable to send buffer")
	ErrInvalidFile   RestError = errors.New("unable to send the file")
	ErrCommitError   RestError = errors.New("unable to commit to your application")
	ErrDelayNow      RestError = errors.New("you are in rate limit, try again later")
)

func ParseError(e *ApiResponse[any]) (err error) {
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
	case "DELAY_NOW", "RATELIMIT":
		err = ErrDelayNow
	default:
		err = fmt.Errorf("square cloud returned error %s", e.Code)
	}

	return
}
