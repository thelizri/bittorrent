package network

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	"karlan/torrent/internal/types"
)

// Constants representing different message IDs used in the BitTorrent protocol.
const (
	BITFIELD   byte = 5 // Contains a bitfield representing the pieces the sender has
	INTERESTED byte = 2 // Indicates the sender wants to download pieces from the recipient
	UNCHOKE    byte = 1 // Indicates the sender will now allow the receiver to request pieces
	REQUEST    byte = 6 // Requests a specific piece of data
	PIECE      byte = 7 // Contains the actual data of the piece being sent
)

const BLOCK_SIZE int = 16 * 1024

// Handshake
const protocolString string = "BitTorrent protocol"
const reservedBytes int = 8

func PerformHandshake(torrent *types.Torrent, peer *types.PeerAddress) ([]byte, net.Conn, error) {
	// make handshake with peer
	handshake := createHandshake(torrent.InfoHash, torrent.PeerID)
	response, conn, err := sendHandshake(peer.GetAddress(), handshake)
	return response, conn, err
}

func createHandshake(infoHash []byte, peerID []byte) []byte {
	handshake := make([]byte, 0, 1+len(protocolString)+reservedBytes+len(infoHash)+len(peerID))
	handshake = append(handshake, byte(len(protocolString)))
	handshake = append(handshake, []byte(protocolString)...)
	reserve := make([]byte, reservedBytes)
	handshake = append(handshake, reserve...)
	handshake = append(handshake, infoHash...)
	handshake = append(handshake, peerID...)
	return handshake
}

func sendHandshake(address string, handshake []byte) ([]byte, net.Conn, error) {
	// Step 1: Establish a TCP connection to the peer
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting to peer:", err)
		conn.Close()
		return nil, nil, err
	}

	// Step 2: Send Message
	_, err = conn.Write(handshake)
	if err != nil {
		fmt.Println("Error sending handshake:", err)
		conn.Close()
		return nil, nil, err
	}

	// Step 3: Receive response
	response := make([]byte, len(handshake))
	_, err = io.ReadFull(conn, response)
	if err != nil {
		fmt.Println("Error receiving handshake:", err)
		conn.Close()
		return nil, nil, err
	}

	return response, conn, nil
}

// Peer Messages
func SendMessageToPeer(conn net.Conn, message *types.PeerMessage) {
	_, err := conn.Write(message.GetBytes())
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}

// Peer messages consist of a message length prefix (4 bytes), message id (1 byte) and a payload (variable size).
func ListenForPeerMessage(conn net.Conn, expectedMessageID byte) (types.PeerMessage, error) {
	for {
		//Set timeout
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		buffer := make([]byte, 4)
		if _, err := io.ReadFull(conn, buffer); err != nil {
			return types.PeerMessage{}, fmt.Errorf("cannot read message payload: %v", err)
		}

		messageLength := binary.BigEndian.Uint32(buffer)
		buffer = make([]byte, messageLength)
		if _, err := io.ReadFull(conn, buffer); err != nil {
			return types.PeerMessage{}, fmt.Errorf("cannot read message payload: %v", err)
		}

		messageID := buffer[0]
		if messageID != expectedMessageID {
			fmt.Printf("Unexpected Message ID: %v, Expected: %v\n", messageID, expectedMessageID)
			continue
		}
		payload := buffer[1:]

		peer_message := types.PeerMessage{MessageID: messageID, Payload: payload}
		return peer_message, nil
	}
}
