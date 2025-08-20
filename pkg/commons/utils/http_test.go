package utils

import (
	"net/http"
	"testing"
)

func TestCreateHTTPClient(t *testing.T) {
	// Test with skipVerify = false
	client := CreateHTTPClient(false)
	if client == nil {
		t.Error("CreateHTTPClient(false) returned nil")
	}

	// Should return the default client when skipVerify is false
	if client != http.DefaultClient {
		t.Error("CreateHTTPClient(false) should return http.DefaultClient")
	}
}

func TestCreateHTTPClientSkipVerify(t *testing.T) {
	// Test with skipVerify = true
	client := CreateHTTPClient(true)
	if client == nil {
		t.Error("CreateHTTPClient(true) returned nil")
	}

	// Should return a custom client when skipVerify is true
	if client == http.DefaultClient {
		t.Error("CreateHTTPClient(true) should not return http.DefaultClient")
	}

	// Check that the transport is properly configured
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Error("Client transport is not *http.Transport")
	}

	if transport.TLSClientConfig == nil {
		t.Error("TLSClientConfig is nil")
	}

	if !transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be true")
	}

	// Check that proxy is set to use environment
	if transport.Proxy == nil {
		t.Error("Proxy is nil")
	}
}

func TestCreateHTTPClientConsistency(t *testing.T) {
	// Test that multiple calls with the same parameters return the same client
	client1 := CreateHTTPClient(false)
	client2 := CreateHTTPClient(false)
	client3 := CreateHTTPClient(true)
	client4 := CreateHTTPClient(true)

	// Default clients should be the same
	if client1 != client2 {
		t.Error("Multiple calls to CreateHTTPClient(false) should return the same client")
	}

	// Custom clients should be different instances but with same configuration
	if client3 == client4 {
		t.Log("Note: Multiple calls to CreateHTTPClient(true) returned the same instance")
	}

	// Verify that custom clients have the same configuration
	transport3, ok3 := client3.Transport.(*http.Transport)
	transport4, ok4 := client4.Transport.(*http.Transport)

	if !ok3 || !ok4 {
		t.Error("Client transports are not *http.Transport")
	}

	if transport3.TLSClientConfig == nil || transport4.TLSClientConfig == nil {
		t.Error("TLSClientConfig is nil")
	}

	if transport3.TLSClientConfig.InsecureSkipVerify != transport4.TLSClientConfig.InsecureSkipVerify {
		t.Error("TLSClientConfig.InsecureSkipVerify should be the same for both clients")
	}
}

func TestCreateHTTPClientTransportProperties(t *testing.T) {
	client := CreateHTTPClient(true)
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Client transport is not *http.Transport")
	}

	// Test that the transport has the expected properties
	if transport.Proxy == nil {
		t.Error("Transport.Proxy should not be nil")
	}

	if transport.TLSClientConfig == nil {
		t.Error("Transport.TLSClientConfig should not be nil")
	}

	if !transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("Transport.TLSClientConfig.InsecureSkipVerify should be true")
	}
}

func TestCreateHTTPClientDefaultClient(t *testing.T) {
	// Test that the default client is properly configured
	client := CreateHTTPClient(false)

	// The default client should be the same as http.DefaultClient
	if client != http.DefaultClient {
		t.Error("CreateHTTPClient(false) should return http.DefaultClient")
	}

	// Test that we can make a request (this is more of an integration test)
	// We'll just check that the client is usable
	if client.Timeout != 0 {
		t.Logf("Default client timeout is set to %v", client.Timeout)
	}
}

func TestCreateHTTPClientCustomClient(t *testing.T) {
	// Test that the custom client is properly configured
	client := CreateHTTPClient(true)

	// The custom client should have a custom transport
	if client.Transport == nil {
		t.Error("Custom client should have a transport")
	}

	// Test that the custom client has the expected configuration
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Client transport is not *http.Transport")
	}

	// Verify TLS configuration
	if transport.TLSClientConfig == nil {
		t.Error("TLSClientConfig should not be nil")
	}

	if !transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be true")
	}

	// Verify proxy configuration
	if transport.Proxy == nil {
		t.Error("Proxy should not be nil")
	}
}
