package tracker

import (
	"encoding/binary"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func TestGET(t *testing.T) {
	type peer struct {
		ip   string
		port uint16
	}

	peersIpAndPort := []peer{
		{"165.232.111.122", 51494},
		{"161.35.47.237", 51480},
		{"139.59.169.165", 51465},
	}

	var wantPeerAddresses []string

	getTorrentAsBytes := func() []byte {
		var peersBytes []byte

		for _, p := range peersIpAndPort {
			peerBytes := create6ByteArray(p.ip, p.port, t)
			peersBytes = append(peersBytes, peerBytes[:]...)
			wantPeerAddresses = append(wantPeerAddresses, p.ip+":"+strconv.Itoa(int(p.port)))
		}

		baseString := "d8:completei3e10:incompletei0e8:intervali60e12:min intervali60e5:peers"
		peersLength := strconv.Itoa(len(peersBytes))
		trackerResponse := baseString + peersLength + ":"
		trackerResponseBytes := []byte(trackerResponse)
		finalBytes := append(trackerResponseBytes, peersBytes...)
		return append(finalBytes, 'e')
	}

	getTestServer := func(torrentAsBytes []byte) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(torrentAsBytes)
		}))
	}

	torrentAsBytes := getTorrentAsBytes()
	ts := getTestServer(torrentAsBytes)
	baseURL, err := url.Parse(ts.URL)

	if err != nil {
		t.Errorf("Cannot parse URL: %s", ts.URL)
	}

	body := sendTrackerRequest(ts.Client(), baseURL)
	interval, peers := parseResponse(body)

	t.Run("Verifying interval", func(t *testing.T) {
		have := interval
		want := 60

		if have != want {
			t.Errorf("have: %d, want: %d", have, want)
		}
	})

	t.Run("Verifying peers", func(t *testing.T) {
		for i, peer := range peers {
			have := peer.Address()
			want := wantPeerAddresses[i]

			if have != want {
				t.Errorf("have: %s, want: %s", have, want)
			}
		}
	})
}

func create6ByteArray(ipStr string, port uint16, t *testing.T) [PeerSize]byte {
	var data [PeerSize]byte

	ip := net.ParseIP(ipStr).To4()

	if ip == nil {
		t.Error("Invalid IP")
	}

	copy(data[:PeerIpBytesCount], ip)

	binary.BigEndian.PutUint16(data[PeerIpBytesCount:], port)

	return data
}
