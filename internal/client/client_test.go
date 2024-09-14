package client

import (
	"fmt"
	"net"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const (
	CorrectIpStr   = "178.62.82.89"
	CorrectPortStr = "51470"
)

var (
	correctAddrStr = fmt.Sprintf("%s:%s", CorrectIpStr, CorrectPortStr)
	correctIp      = net.ParseIP(CorrectIpStr)
	correctPort, _ = strconv.Atoi(CorrectPortStr)
)

func TestNew(t *testing.T) {
	have := New(correctIp, uint16(correctPort))
	want := &Client{Choked: true, IP: correctIp, Port: uint16(correctPort)}

	if !have.IP.Equal(want.IP) {
		t.Errorf("IP inequality: have %s, expected %s", have.IP.String(), want.IP.String())
	}

	if have.Choked != want.Choked {
		t.Errorf("Choked inequality: have %v, expected %v", have.Choked, want.Choked)
	}

	if have.Port != want.Port {
		t.Errorf("Port inequality: have %d, expected %d", have.Port, want.Port)
	}

	if have.Conn != want.Conn {
		t.Errorf("Conn should be uninitialized but is initialized")
	}

	if have.PeerID != want.PeerID {
		t.Errorf("PeerID should be uninitialized but is initialized")
	}

	if have.Bitfield != want.Bitfield {
		t.Errorf("Bitfield should be uninitialized but is initialized")
	}
}

func TestStringToClient(t *testing.T) {
	tests := []struct {
		addr string
		want Client
	}{
		{correctAddrStr, New(correctIp, uint16(correctPort))},
	}

	failingTests := []struct {
		addr, expectedError string
	}{
		{"178.62.82.89.51470", "address must be in the format IP:Port"},
		{":51470", "invalid IP address"},
		{"178.62.82.89:", "invalid port number"},
		{"178.62.82.89:-1", "port number must be between 0 and 65535"},
		{"178.62.82.89:65536", "port number must be between 0 and 65535"},
	}

	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			have, err := StringToClient(tt.addr)
			if err != nil {
				t.Errorf("Error: %s", err.Error())
			}

			if !cmp.Equal(have, tt.want) {
				t.Errorf("Client inequality: have %v, expected %v", have, tt.want)
			}
		})
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
