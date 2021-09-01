package commando

// StopOnError is a helper which stops processing a file on the first
// encountered error.
func StopOnError(err error) error {
	return err
}
