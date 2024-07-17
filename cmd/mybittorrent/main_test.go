package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// Test function
func TestDecode(t *testing.T) {

	//Decoding string
	result := strings.TrimSpace(getStandardOutput(func() { decodeBencodedString("5:hello") }))
	expected := `"hello"`
	if result != expected {
		t.Errorf("Result was incorrect, got: %s, want: %s.", result, expected)
	}

	//Decoding number
	result = strings.TrimSpace(getStandardOutput(func() { decodeBencodedString("i52e") }))
	expected = "52"
	if result != expected {
		t.Errorf("Result was incorrect, got: %s, want: %s.", result, expected)
	}

	//Decoding number
	result = strings.TrimSpace(getStandardOutput(func() { decodeBencodedString("i-52e") }))
	expected = "-52"
	if result != expected {
		t.Errorf("Result was incorrect, got: %s, want: %s.", result, expected)
	}

	//Decoding list
	result = strings.TrimSpace(getStandardOutput(func() { decodeBencodedString("l5:helloi52ee") }))
	expected = `["hello",52]`
	if result != expected {
		t.Errorf("Result was incorrect, got: %s, want: %s.", result, expected)
	}

	//Decoding dictionary
	result = strings.TrimSpace(getStandardOutput(func() { decodeBencodedString("d3:foo3:bar5:helloi52ee") }))
	expected = `{"foo":"bar","hello":52}`
	if result != expected {
		t.Errorf("Result was incorrect, got: %s, want: %s.", result, expected)
	}
}

func TestInfo(t *testing.T) {
	value := func() {
		printTorrentInfo("../../sample.torrent")
	}
	output := getStandardOutput(value)

	expectedTrackerURL := "http://bittorrent-test-tracker.codecrafters.io/announce"
	if !strings.Contains(output, expectedTrackerURL) {
		t.Errorf("Expected output to contain: %s", expectedTrackerURL)
	}

	expectedLength := "Length: 92063"
	if !strings.Contains(output, expectedLength) {
		t.Errorf("Expected output to contain: %s", expectedLength)
	}

	expectedInfoHash := "Info Hash: d69f91e6b2ae4c542468d1073a71d4ea13879a7f"
	if !strings.Contains(output, expectedInfoHash) {
		t.Errorf("Expected output to contain: %s", expectedInfoHash)
	}

	expectedPieceLength := "Piece Length: 32768"
	if !strings.Contains(output, expectedPieceLength) {
		t.Errorf("Expected output to contain: %s", expectedPieceLength)
	}

	expectedPieceHashes := []string{
		"Piece Hashes:",
		"e876f67a2a8886e8f36b136726c30fa29703022d",
		"6e2275e604a0766656736e81ff10b55204ad8d35",
		"f00d937a0213df1982bc8d097227ad9e909acc17",
	}

	for _, hash := range expectedPieceHashes {
		if !strings.Contains(output, hash) {
			t.Errorf("Expected output to contain: %s", hash)
		}
	}
}

func TestPeers(t *testing.T) {
	value := func() {
		printPeers("../../sample.torrent")
	}
	output := getStandardOutput(value)

	expectedIP := "178.62.82.89"
	if !strings.Contains(output, expectedIP) {
		t.Errorf("Expected output to contain: %s", expectedIP)
	}
	expectedIP = "165.232.33.77"
	if !strings.Contains(output, expectedIP) {
		t.Errorf("Expected output to contain: %s", expectedIP)
	}
	expectedIP = "178.62.85.20"
	if !strings.Contains(output, expectedIP) {
		t.Errorf("Expected output to contain: %s", expectedIP)
	}
}

func getStandardOutput(value func()) string {
	// Redirect stdout to capture the output
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	print()

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	value()

	// back to normal state
	w.Close()
	os.Stdout = old // restoring the real stdout
	output := <-outC
	return output
}
