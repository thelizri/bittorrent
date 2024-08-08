package client

import (
	"encoding/hex"
	"fmt"
	"io"

	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

const BLOCK_SIZE int = 16 * 1024

// Handshake constants
const protocolLength byte = 19
const protocolString string = "BitTorrent protocol"
const reservedBytes int = 8
const handshakeLength int = 68

// Handshake represents a BitTorrent handshake
type Handshake struct {
	Length   byte
	Protocol string
	Reserved [8]byte
	InfoHash [20]byte
	PeerID   [20]byte
}

func (h *Handshake) log() {
	log.Info(h.string())
}

// String returns a formatted string of the Handshake struct
func (h *Handshake) string() string {
	return fmt.Sprintf("Handshake{\n  Length: %d,\n  Protocol: %s,\n  Reserved: %s,\n  InfoHash: %s,\n  PeerID: %s\n}",
		h.Length,
		h.Protocol,
		hex.EncodeToString(h.Reserved[:]),
		hex.EncodeToString(h.InfoHash[:]),
		hex.EncodeToString(h.PeerID[:]),
	)
}

// newHandshake creates a new handshake message
func newHandshake(infoHash, peerID [20]byte) *Handshake {
	log.Info("Creating new handshake: InfoHash=%s, PeerID=%s\n", hex.EncodeToString(infoHash[:]), hex.EncodeToString(peerID[:]))
	hs := &Handshake{
		Length:   protocolLength,
		Protocol: protocolString,
	}
	hs.InfoHash = infoHash
	hs.PeerID = peerID
	return hs
}

// serialize converts the handshake into a byte slice
// <length:1><protocol id:19><reserved bytes:8><info hash:20><peer id:20>
func (h *Handshake) serialize() []byte {
	log.Info("Serializing handshake")
	buf := make([]byte, handshakeLength)
	buf[0] = h.Length
	curr := 1
	curr += copy(buf[curr:], []byte(h.Protocol))
	curr += copy(buf[curr:], h.Reserved[:])
	curr += copy(buf[curr:], h.InfoHash[:])
	curr += copy(buf[curr:], h.PeerID[:])
	log.Info("Serialized handshake: %x\n", buf)
	return buf
}

// parseHandshake parses a handshake message from a byte slice
func parseHandshake(buf []byte) (*Handshake, error) {
	log.Info("Parsing handshake: %x\n", buf)
	if len(buf) < handshakeLength {
		return nil, fmt.Errorf("invalid handshake length: %d", len(buf))
	}

	h := &Handshake{}
	h.Length = buf[0]
	log.Info("Parsed Length: %d\n", h.Length)
	if int(h.Length) != len(protocolString) {
		return nil, fmt.Errorf("invalid protocol string length")
	}

	h.Protocol = string(buf[1 : 1+int(h.Length)])
	log.Info("Parsed Protocol: %s\n", h.Protocol)
	if h.Protocol != protocolString {
		return nil, fmt.Errorf("invalid protocol string: %s", h.Protocol)
	}

	copy(h.Reserved[:], buf[1+int(h.Length):1+int(h.Length)+reservedBytes])
	copy(h.InfoHash[:], buf[1+int(h.Length)+reservedBytes:21+int(h.Length)+reservedBytes])
	copy(h.PeerID[:], buf[21+int(h.Length)+reservedBytes:handshakeLength])

	log.Info("Parsed Handshake: %s\n", h.string())
	return h, nil
}

func sendHandshake(address string, handshake *Handshake) (*Handshake, net.Conn, error) {
	log.Info("Sending handshake to address: %s\n", address)

	// Step 1: Establish a TCP connection to the peer
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		log.Info("Error establishing TCP connection: %s", err)
		return nil, nil, err
	}
	log.Info("TCP connection established with peer")

	// Step 2: Send Message
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})
	_, err = conn.Write(handshake.serialize())
	if err != nil {
		log.Info("Error sending handshake: %s", err)
		conn.Close()
		return nil, nil, err
	}
	log.Info("Handshake sent successfully")

	// Step 3: Receive response
	responseBuf := make([]byte, handshakeLength)
	_, err = io.ReadFull(conn, responseBuf)
	if err != nil {
		log.Info("Error receiving handshake: %s", err)
		conn.Close()
		return nil, nil, err
	}
	log.Info("Received handshake response: %x\n", responseBuf)

	// Step 4: Parse handshake
	response, err := parseHandshake(responseBuf)
	if err != nil {
		conn.Close()
		log.Info("Error parsing handshake")
		return nil, nil, err
	}
	log.Info("Handshake parsed successfully")

	return response, conn, nil
}
