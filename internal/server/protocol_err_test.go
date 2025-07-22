package server

import (
	"fmt"
	"testing"
)

func TestProtocolError(t *testing.T) {

	err := ProtocolError{Protocol: "http", Address: ":8080", Err: fmt.Errorf("failure")}
	if err.Error() == "" {
		t.Fatalf("unexpected empty error")
	}

	if unwrap := err.Unwrap(); unwrap == nil {
		t.Fatalf("expected wrapped error")
	}
}
