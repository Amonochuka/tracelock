package main

import (
	"net/http"
	"testing"
)

func TestNewHTTPServerSetsTimeouts(t *testing.T) {
	server := newHTTPServer(":8080", http.NewServeMux())

	if server.ReadHeaderTimeout != serverReadHeaderTimeout {
		t.Fatalf("expected ReadHeaderTimeout to be %v, got %v", serverReadHeaderTimeout, server.ReadHeaderTimeout)
	}
	if server.ReadTimeout != serverReadTimeout {
		t.Fatalf("expected ReadTimeout to be %v, got %v", serverReadTimeout, server.ReadTimeout)
	}
	if server.WriteTimeout != serverWriteTimeout {
		t.Fatalf("expected WriteTimeout to be %v, got %v", serverWriteTimeout, server.WriteTimeout)
	}
	if server.IdleTimeout != serverIdleTimeout {
		t.Fatalf("expected IdleTimeout to be %v, got %v", serverIdleTimeout, server.IdleTimeout)
	}
}
