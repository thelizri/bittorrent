package client

import (
	"fmt"
	"net"
	"strconv"
	"testing"
)

const (
	CorrectIpStr = "178.62.82.89"
	CorrectPortStr = "51470"
)

var (
	correctAddrStr = fmt.Sprintf("%s:%s", CorrectIpStr, CorrectPortStr)
	correctIp = net.ParseIP(CorrectIpStr)
	correctPort, _ = strconv.Atoi(CorrectPortStr)
)

func TestStringToClient(t *testing.T) {
	failingTests := []struct {
		addr, expectedError string
	}{
		{"178.62.82.89.51470", "address must be in the format IP:Port"},
		{":51470", "invalid IP address"},
		{"178.62.82.89:", "invalid port number"},
		{"178.62.82.89:-1", "port number must be between 0 and 65535"},
		{"178.62.82.89:65536", "port number must be between 0 and 65535"},
	}

	for _, tt := range failingTests {
		t.Run(fmt.Sprintf("%s throws error", tt.addr), func(t *testing.T) {
			_, err := StringToClient(tt.addr)
			if err.Error() != tt.expectedError {
				t.Errorf("Expected error '%s' but none was thrown", tt.expectedError)
			}
		})
	}
}
