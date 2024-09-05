package tracker

import (
	"encoding/binary"
	"karlan/torrent/internal/torrent"
	"net"
	"net/http"
	"net/http/httptest"
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

	getTrackerAsBytes := func() []byte {
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

	getTestServer := func(trackerAsBytes []byte) *httptest.Server {
		getExpectedQueryHandler := func(expectedQuery string, handler http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				expected := r.URL.Query().Encode()

				if expected == expectedQuery {
					handler(w, r)
				} else {
					t.Error("Expected and actual query mismatch")
				}
			}
		}

		expectedQuery := "compact=1&downloaded=2048&info_hash=%01%02%03%04%05%06%07%08%09%0A%0B%0C%0D%0E%0F%10%11%12%13%14&left=4096&peer_id=%FF%EE%DD%CC%BB%AA%99%88wfUD3%22%11%00%99%88wf&port=6881&uploaded=1024"

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write(trackerAsBytes)
		})

		return httptest.NewServer(getExpectedQueryHandler(expectedQuery, handler))
	}

	trackerAsBytes := getTrackerAsBytes()
	ts := getTestServer(trackerAsBytes)
	torrent := &torrent.Torrent{
		Announce:   ts.URL,
		InfoHash:   [20]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14},
		PeerID:     [20]byte{0xFF, 0xEE, 0xDD, 0xCC, 0xBB, 0xAA, 0x99, 0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, 0x00, 0x99, 0x88, 0x77, 0x66},
		Port:       6881,
		Uploaded:   1024,
		Downloaded: 2048,
		Left:       4096,
	}
	baseURL := createTrackerRequest(torrent)
	// baseURL, err := url.Parse(ts.URL)
	//
	// if err != nil {
	// 	t.Errorf("Cannot parse URL: %s", ts.URL)
	// }

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
