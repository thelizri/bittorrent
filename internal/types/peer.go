package types

import (
	"encoding/binary"
	"fmt"
	"net"
)

// Holds the address of a peer in the torrent network
type PeerAddress struct {
	IP   net.IP
	Port uint16
}

func (p *PeerAddress) Print() {
	fmt.Printf("%v:%v\n", p.IP, p.Port)
}

func (p *PeerAddress) GetAddress() string {
	return fmt.Sprintf("%v:%v", p.IP, p.Port)
}

// PeerMessage represents a message exchanged between peers in the BitTorrent protocol.
type PeerMessage struct {
	MessageLength uint32 // Length of the message
	MessageID     byte   // ID of the message type
	Payload       []byte // The payload of the message
}

func (p *PeerMessage) setLength() {
	p.MessageLength = uint32(1 + len(p.Payload))
}

func (p *PeerMessage) GetBytes() []byte {
	p.setLength()

	buffer := make([]byte, 5+len(p.Payload))
	binary.BigEndian.PutUint32(buffer, p.MessageLength)
	buffer[4] = p.MessageID
	copy(buffer[5:], p.Payload)

	return buffer
}
