package types

import (
	"fmt"
	"log"
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

func (p *PeerAddress) Log() {
	log.Printf("%v:%v\n", p.IP, p.Port)
}

func (p *PeerAddress) GetAddress() string {
	return fmt.Sprintf("%v:%v", p.IP, p.Port)
}

// PeerMessage represents a message exchanged between peers in the BitTorrent protocol.
