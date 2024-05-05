package cli

type AuthError struct {
	err error
}

func (ae *AuthError) Error() string {
	return ae.err.Error()
}
