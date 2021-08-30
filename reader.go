package gocsv

// Reader is an interface over csv.Reader, which allows swapping the
// implementation, if necessary.
type Reader interface {
	Read() ([]string, error)
}
