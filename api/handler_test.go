package api_test

import (
	"crypto/tls"
	"testing"
)

const (
	port       = "4430"
	localHost  = "localhost"
	remoteName = "go-git-test"
)

var certs []tls.Certificate

func TestCreateRepo(t *testing.T) {
	t.Skipf("Not yet implemented")
}
