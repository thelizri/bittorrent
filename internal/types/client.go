package types

import "net"

// A Client is a TCP connection with a peer
type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield Bitfield
	Address  PeerAddress
	PeerID   []byte
}

func NewClient(conn net.Conn, bitfield Bitfield, address PeerAddress, peerid []byte) *Client {
	client := &Client{Conn: conn, Choked: true, Bitfield: bitfield, Address: address, PeerID: peerid}
	return client
}
