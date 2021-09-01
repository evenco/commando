package commando

import "context"

// StopOnError is a helper which stops processing a file on the first
// encountered error.
func StopOnError(_ context.Context, err error) error {
	return err
}
