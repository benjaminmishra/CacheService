package server

import "fmt"

// Structured error that helps track errors from multiple protocols
type ProtocolError struct {
	Protocol string
	Address  string
	Err      error
}

func (pe ProtocolError) Error() string {
	return fmt.Sprintf("%s server on %s: %v", pe.Protocol, pe.Address, pe.Err)
}

func (pe ProtocolError) Unwrap() error {
	return pe.Err
}
