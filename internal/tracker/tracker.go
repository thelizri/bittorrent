package tracker

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"

	"karlan/torrent/internal/bencode"
	"karlan/torrent/internal/client"
	"karlan/torrent/internal/torrent"
)

func GET(torrent *torrent.Torrent) (int, []client.Client) {
	baseURL := createTrackerRequest(torrent)
	body := sendTrackerRequest(baseURL)
	interval, peers := parseResponse(body)
	return interval, peers
}

func createTrackerRequest(torrent *torrent.Torrent) *url.URL {
	// Prepare query parameters for the tracker request
	params := url.Values{}
	params.Add("info_hash", string(torrent.InfoHash[:]))
	params.Add("peer_id", string(torrent.PeerID[:])) // This should ideally be dynamically generated or unique
	params.Add("port", fmt.Sprintf("%d", torrent.Port))
	params.Add("uploaded", fmt.Sprintf("%d", torrent.Uploaded))
	params.Add("downloaded", fmt.Sprintf("%d", torrent.Downloaded))
	params.Add("left", fmt.Sprintf("%d", torrent.Left))
	params.Add("compact", "1") // Indicates compact response format is preferred

	// Build URL with query parameters
	baseURL, err := url.Parse(torrent.Announce)
	if err != nil {
		log.Errorf("Error parsing announce URL: %v", err)
		return nil
	}
	baseURL.RawQuery = params.Encode()

	log.Debugf("Created tracker request URL: %s", baseURL.String())
	return baseURL
}

func sendTrackerRequest(baseURL *url.URL) []byte {
	// Make GET request
	resp, err := http.Get(baseURL.String())
	if err != nil {
		log.Errorf("Error making GET request: %v", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error reading response body: %v", err)
		return nil
	}

	log.Debugf("Received tracker response: %x", body)
	return body
}

func parseResponse(body []byte) (int, []client.Client) {
	// Decoding Body
	encoding, _, err := bencode.Decode(string(body), 0)
	if err != nil {
		log.Errorf("Error decoding body: %v", err)
		return 0, nil
	}

	response, ok := encoding.(map[string]interface{})
	if !ok {
		log.Error("Decoded response is not a map[string]interface{}")
		return 0, nil
	}

	interval, ok := response["interval"].(int)
	if !ok {
		log.Error("Error: 'interval' field is missing or not an int")
		return 0, nil
	}

	peersStr, ok := response["peers"].(string)
	if !ok {
		log.Error("Error: 'peers' field is missing or not a string")
		return 0, nil
	}

	bytes := []byte(peersStr)
	if len(bytes)%6 != 0 {
		log.Error("Error: 'peers' string length is not a multiple of 6")
		return 0, nil
	}

	peers := make([]client.Client, 0, len(bytes)/6)
	for i := 0; i < len(bytes)/6; i++ {
		peerBytes := bytes[i*6 : (i+1)*6]
		log.Tracef("Extracting peer %d: %s", i, hex.EncodeToString(peerBytes))
		ip, port, err := parsePeers(peerBytes)
		if err != nil {
			log.Warnf("Error converting bytes to PeerAddress: %v", err)
			continue
		}
		peers = append(peers, client.New(ip, port))
	}

	log.Infof("Successfully extracted %d peers", len(peers))
	return interval, peers
}

// bytesToPeerAddress converts a 6-byte array to an IP address and port.
func parsePeers(data []byte) (net.IP, uint16, error) {
	if len(data) != 6 {
		return nil, 0, fmt.Errorf("data must be exactly 6 bytes long")
	}

	ip := net.IP(data[:4])
	port := binary.BigEndian.Uint16(data[4:])

	log.Debugf("Parsed peer: IP=%s, Port=%d", ip.String(), port)
	return ip, port, nil
}
