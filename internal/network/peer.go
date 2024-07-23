package network

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"karlan/torrent/internal/types"
	"karlan/torrent/internal/utils"
)

const BLOCK_SIZE int = 16 * 1024

// Handshake
const protocolString string = "BitTorrent protocol"
const reservedBytes int = 8

func ConnectToPeer(torrent *types.Torrent, peers []types.PeerAddress) net.Conn {
	for _, peer := range peers {
		_, conn, err := PerformHandshake(torrent, &peer)
		if err != nil {
			log.Printf("Error connecting to peer: %v\n", err)
			continue
		} else {
			utils.LogAndPrint(fmt.Sprintf("Established connection with: %v\n", peer.GetAddress()))
			return conn
		}
	}
	utils.LogAndPrint("Failed to establish connection with any peer.")
	os.Exit(1)
	return nil
}

func ConnectToPeers(torrent *types.Torrent, peers []types.PeerAddress) []net.Conn {
	result := make([]net.Conn, 0)
	for _, peer := range peers {
		_, conn, err := PerformHandshake(torrent, &peer)
		if err != nil {
			log.Printf("Error connecting to peer: %v\n", err)
			continue
		} else {
			fmt.Printf("Established connection with: %v\n", peer.GetAddress())
			result = append(result, conn)
		}
	}

	if len(result) > 0 {
		return result
	}

	utils.LogAndPrint("Failed to establish connection with any peer.")
	os.Exit(1)
	return nil
}

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
		log.Println("Error connecting to peer:", err)
		conn.Close()
		return nil, nil, err
	}

	// Step 2: Send Message
	_, err = conn.Write(handshake)
	if err != nil {
		log.Println("Error sending handshake:", err)
		conn.Close()
		return nil, nil, err
	}

	// Step 3: Receive response
	response := make([]byte, len(handshake))
	_, err = io.ReadFull(conn, response)
	if err != nil {
		log.Println("Error receiving handshake:", err)
		conn.Close()
		return nil, nil, err
	}

	return response, conn, nil
}

// Peer Messages
func SendMessageToPeer(conn net.Conn, message *types.Message) {
	_, err := conn.Write(message.GetBytes())
	if err != nil {
		log.Println("Error:", err)
		return
	}
}

// Peer messages consist of a message length prefix (4 bytes), message id (1 byte) and a payload (variable size).
func ListenForPeerMessage(conn net.Conn, expectedMessageID types.MessageID) (types.Message, error) {
	var zero types.Message
	for {
		//Set timeout
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		buffer := make([]byte, 4)
		if _, err := io.ReadFull(conn, buffer); err != nil {
			return zero, fmt.Errorf("cannot read message payload: %v", err)
		}

		messageLength := binary.BigEndian.Uint32(buffer)
		buffer = make([]byte, messageLength)
		if _, err := io.ReadFull(conn, buffer); err != nil {
			return zero, fmt.Errorf("cannot read message payload: %v", err)
		}

		messageID := types.MessageID(buffer[0])
		if messageID != expectedMessageID {
			fmt.Printf("Unexpected Message ID: %v, Expected: %v\n", messageID.String(), expectedMessageID.String())
			continue
		}
		payload := buffer[1:]

		peer_message := types.Message{MessageID: messageID, Payload: payload}
		return peer_message, nil
	}
}
