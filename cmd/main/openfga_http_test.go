package main

import (
	"net/http"
	"testing"
)

func TestNewNoPoolHTTPClient(t *testing.T) {
	c := newNoPoolHTTPClient()

	tr, ok := c.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}
	if !tr.DisableKeepAlives {
		t.Fatal("expected DisableKeepAlives to be true")
	}
	if tr == http.DefaultTransport {
		t.Fatal("transport must be a clone, not http.DefaultTransport itself")
	}
}
