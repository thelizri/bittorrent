package network

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"karlan/torrent/internal/types"
)

func PerformTrackerRequest(torrent *types.Torrent) []byte {
	baseURL := createTrackerRequest(torrent)
	return sendTrackerRequest(baseURL)
}

func createTrackerRequest(torrent *types.Torrent) *url.URL {
	// Prepare query parameters for the tracker request
	params := url.Values{}
	params.Add("info_hash", string(torrent.InfoHash))
	params.Add("peer_id", string(torrent.PeerID)) // This should ideally be dynamically generated or unique
	params.Add("port", fmt.Sprintf("%d", torrent.Port))
	params.Add("uploaded", fmt.Sprintf("%d", torrent.Uploaded))
	params.Add("downloaded", fmt.Sprintf("%d", torrent.Downloaded))
	params.Add("left", fmt.Sprintf("%d", torrent.Left))
	params.Add("compact", "1") // Indicates compact response format is preferred

	// Build URL with query parameters
	baseURL, err := url.Parse(torrent.Announce)
	if err != nil {
		fmt.Println("Error parsing announce URL:", err)
		return nil
	}
	baseURL.RawQuery = params.Encode()

	return baseURL
}

func sendTrackerRequest(baseURL *url.URL) []byte {
	// Make GET request
	resp, err := http.Get(baseURL.String())
	if err != nil {
		fmt.Println("Error making GET request:", err)
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil
	}
	return body
}
