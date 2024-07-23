package utils

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"karlan/torrent/internal/bencode"
	"karlan/torrent/internal/types"
)

const LINE_SEPARATOR string = "-------------------------------------------------------------------------"

func LogSeparator() {
	log.Printf("\n%s\n", LINE_SEPARATOR)
}

func LogAndPrint(message string) {
	log.Println(message)
	fmt.Println(message)
}

// bytesToPeerAddress converts a 6-byte array to an IP address and port.
func BytesToPeerAddress(data []byte) (types.PeerAddress, error) {
	if len(data) != 6 {
		return types.PeerAddress{}, fmt.Errorf("data must be exactly 6 bytes long")
	}

	ip := net.IP(data[:4])
	port := binary.BigEndian.Uint16(data[4:])

	return types.PeerAddress{IP: ip, Port: port}, nil
}

// stringToPeerAddress converts a string in the format "IP:Port" to a PeerAddress.
func StringToPeerAddress(addr string) (types.PeerAddress, error) {
	// Split the string into IP and port
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return types.PeerAddress{}, fmt.Errorf("address must be in the format IP:Port")
	}

	// Parse the IP
	ip := net.ParseIP(parts[0])
	if ip == nil {
		return types.PeerAddress{}, fmt.Errorf("invalid IP address")
	}

	// Parse the port
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return types.PeerAddress{}, fmt.Errorf("invalid port number")
	}

	// Check if the port is within the valid range
	if port < 0 || port > 65535 {
		return types.PeerAddress{}, fmt.Errorf("port number must be between 0 and 65535")
	}

	return types.PeerAddress{IP: ip, Port: uint16(port)}, nil
}

func ExtractPeersFromResponse(body []byte) []types.PeerAddress {
	// Decoding Body
	encoding, _, err := bencode.Decode(string(body), 0)
	if err != nil {
		log.Println("Error decoding body:", err)
		return nil
	}

	response, ok := encoding.(map[string]interface{})
	if !ok {
		log.Println("Error: Decoded response is not a map[string]interface{}")
		return nil
	}

	_, ok = response["peers"]
	if !ok {
		log.Println("Error: 'peers' field not found")
		return nil
	}

	peersStr, ok := response["peers"].(string)
	if !ok {
		log.Println("Error: 'peers' field is not a string")
		return nil
	}

	bytes := []byte(peersStr)
	if len(bytes)%6 != 0 {
		log.Println("Error: 'peers' string length is not a multiple of 6")
		return nil
	}

	peers := make([]types.PeerAddress, 0, len(bytes)/6)
	for i := 0; i < len(bytes)/6; i++ {
		peerBytes := bytes[i*6 : (i+1)*6]
		log.Printf("Extracting peer %d: %s", i, hex.EncodeToString(peerBytes))
		peer, err := BytesToPeerAddress(peerBytes)
		if err != nil {
			log.Printf("Error converting bytes to PeerAddress: %v", err)
			continue
		}
		peers = append(peers, peer)
	}

	log.Printf("Successfully extracted %d peers", len(peers))
	return peers
}
