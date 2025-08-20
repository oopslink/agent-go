// Package utils provides utility functions for the agent-go framework.
// This file contains HTTP client utilities for making HTTP requests.
package utils

import (
	"crypto/tls"
	"net/http"
)

// CreateHTTPClient creates an HTTP client with optional TLS certificate verification.
// If skipVerify is true, the client will skip TLS certificate verification,
// which is useful for development or testing environments with self-signed certificates.
// If skipVerify is false, the client uses the default HTTP client with full verification.
func CreateHTTPClient(skipVerify bool) *http.Client {
	if !skipVerify {
		return http.DefaultClient
	}
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment, // Use proxy settings from environment variables
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // Skip TLS certificate verification
			},
		},
	}
}
