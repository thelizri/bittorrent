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
	log.Debug(h.string())
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
	log.Debugf("Creating new handshake: InfoHash=%s, PeerID=%s", hex.EncodeToString(infoHash[:]), hex.EncodeToString(peerID[:]))
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
	log.Debug("Serializing handshake")
	buf := make([]byte, handshakeLength)
	buf[0] = h.Length
	curr := 1
	curr += copy(buf[curr:], []byte(h.Protocol))
	curr += copy(buf[curr:], h.Reserved[:])
	curr += copy(buf[curr:], h.InfoHash[:])
	curr += copy(buf[curr:], h.PeerID[:])
	log.Tracef("Serialized handshake: %x", buf)
	return buf
}

// parseHandshake parses a handshake message from a byte slice
func parseHandshake(buf []byte) (*Handshake, error) {
	log.Debugf("Parsing handshake: %x", buf)
	if len(buf) < handshakeLength {
		return nil, fmt.Errorf("invalid handshake length: %d", len(buf))
	}

	h := &Handshake{}
	h.Length = buf[0]
	log.Debugf("Parsed Length: %d", h.Length)
	if int(h.Length) != len(protocolString) {
		return nil, fmt.Errorf("invalid protocol string length")
	}

	h.Protocol = string(buf[1 : 1+int(h.Length)])
	log.Debugf("Parsed Protocol: %s", h.Protocol)
	if h.Protocol != protocolString {
		return nil, fmt.Errorf("invalid protocol string: %s", h.Protocol)
	}

	copy(h.Reserved[:], buf[1+int(h.Length):1+int(h.Length)+reservedBytes])
	copy(h.InfoHash[:], buf[1+int(h.Length)+reservedBytes:21+int(h.Length)+reservedBytes])
	copy(h.PeerID[:], buf[21+int(h.Length)+reservedBytes:handshakeLength])

	log.Debugf("Parsed Handshake: %s", h.string())
	return h, nil
}

func sendHandshake(address string, handshake *Handshake) (*Handshake, net.Conn, error) {
	log.Infof("Sending handshake to address: %s", address)

	// Step 1: Establish a TCP connection to the peer
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		log.Errorf("Error establishing TCP connection: %v", err)
		return nil, nil, err
	}
	log.Info("TCP connection established with peer")

	// Step 2: Send Message
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})
	_, err = conn.Write(handshake.serialize())
	if err != nil {
		log.Errorf("Error sending handshake: %v", err)
		conn.Close()
		return nil, nil, err
	}
	log.Info("Handshake sent successfully")

	// Step 3: Receive response
	responseBuf := make([]byte, handshakeLength)
	_, err = io.ReadFull(conn, responseBuf)
	if err != nil {
		log.Errorf("Error receiving handshake: %v", err)
		conn.Close()
		return nil, nil, err
	}
	log.Info("Received handshake response")

	// Step 4: Parse handshake
	response, err := parseHandshake(responseBuf)
	if err != nil {
		conn.Close()
		log.Error("Error parsing handshake")
		return nil, nil, err
	}
	log.Info("Handshake parsed successfully")

	return response, conn, nil
}
