package client

import (
	"encoding/binary"
	"fmt"
	"io"
	"karlan/torrent/internal/utils"
	"net"
	"strconv"
	"strings"
	"time"
)

// A Client is a TCP connection with a peer
type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield *Bitfield
	IP       net.IP
	Port     uint16
	PeerID   [20]byte
}

func New(ip net.IP, port uint16) Client {
	utils.LogAndPrintf("Creating new client: IP=%v, Port=%v\n", ip, port)
	return Client{Choked: true, IP: ip, Port: port}
}

func (c *Client) Address() string {
	address := fmt.Sprintf("%v:%v", c.IP, c.Port)
	utils.LogAndPrintf("Client address: %s\n", address)
	return address
}

func (c *Client) Init(infoHash, peerID []byte) error {
	utils.LogAndPrintf("Initializing client: InfoHash=%x, PeerID=%x\n", infoHash, peerID)
	handshake := newHandshake(infoHash, peerID)
	handshake.log()

	response, conn, err := sendHandshake(c.Address(), handshake)
	if err != nil {
		utils.LogAndPrintf("Error during handshake: %v\n", err)
		return err
	}
	response.log()

	c.Conn = conn
	c.PeerID = response.PeerID
	utils.LogAndPrintf("Connected to peer: PeerID=%x\n", c.PeerID)

	c.Bitfield, err = c.receiveBitfield()
	if err != nil {
		utils.LogAndPrintf("Error receiving bitfield: %v\n", err)
		return err
	}
	utils.LogAndPrintln("Bitfield received successfully")
	return nil
}

// Peer messages consist of a message length prefix (4 bytes), message id (1 byte) and a payload (variable size).
func (c *Client) receiveBitfield() (*Bitfield, error) {
	utils.LogAndPrintln("Receiving bitfield from peer")
	conn := c.Conn

	// Set timeout, Keep alive message is sent every 2 minutes
	conn.SetReadDeadline(time.Now().Add(2 * time.Minute))
	defer conn.SetReadDeadline(time.Time{})

	buffer := make([]byte, 4)
	if _, err := io.ReadFull(conn, buffer); err != nil {
		return nil, fmt.Errorf("cannot read message length: %v", err)
	}

	messageLength := binary.BigEndian.Uint32(buffer)
	utils.LogAndPrintf("Bitfield message length: %d\n", messageLength)
	buffer = make([]byte, messageLength)
	if _, err := io.ReadFull(conn, buffer); err != nil {
		return nil, fmt.Errorf("cannot read message payload: %v", err)
	}

	messageID := MessageID(buffer[0])
	utils.LogAndPrintf("Received message ID: %v\n", messageID)
	if messageID != MSG_BITFIELD {
		return nil, fmt.Errorf("expected bitfield but got message id %v", messageID)
	}
	var bitfield Bitfield = buffer[1:]
	utils.LogAndPrintf("Bitfield data: %x\n", bitfield)

	return &bitfield, nil
}

// Peer messages consist of a message length prefix (4 bytes), message id (1 byte) and a payload (variable size).
func (c *Client) Read() (*Message, error) {
	utils.LogAndPrintln("Reading message from client")
	conn := c.Conn

	// Set timeout, Keep alive message is sent every 2 minutes
	conn.SetReadDeadline(time.Now().Add(2 * time.Minute))
	defer conn.SetReadDeadline(time.Time{})

	buffer := make([]byte, 4)
	if _, err := io.ReadFull(conn, buffer); err != nil {
		return nil, fmt.Errorf("cannot read message length: %v", err)
	}

	messageLength := binary.BigEndian.Uint32(buffer)
	utils.LogAndPrintf("Message length: %d\n", messageLength)
	if messageLength == 0 {
		utils.LogAndPrintln("Received keep-alive message")
		return nil, nil
	}

	buffer = make([]byte, messageLength)
	if _, err := io.ReadFull(conn, buffer); err != nil {
		return nil, fmt.Errorf("cannot read message payload: %v", err)
	}

	messageID := MessageID(buffer[0])
	payload := buffer[1:]
	utils.LogAndPrintf("Received message ID: %v, Payload length: %d\n", messageID, len(payload))

	peer_message := &Message{MessageID: messageID, Payload: payload}
	return peer_message, nil
}

func (c *Client) Send(message *Message) error {
	utils.LogAndPrintf("Sending message to peer. Message ID: %v\n", message.MessageID.String())
	if c.Conn == nil {
		utils.LogAndPrintln("Connection is nil, cannot send message")
		return fmt.Errorf("connection is nil")
	}
	_, err := c.Conn.Write(message.Serialize())
	if err != nil {
		utils.LogAndPrintf("Error sending message: %v\n", err)
		return err
	}

	utils.LogAndPrintln("Message sent successfully")
	return nil
}

func (c *Client) HasPiece(index int) bool {
	hasPiece := c.Bitfield.HasPiece(index)
	utils.LogAndPrintf("Checking if client has piece %d: %v\n", index, hasPiece)
	return hasPiece
}

func (c *Client) AddPiece(message *Message) {
	index := binary.BigEndian.Uint32(message.Payload)
	utils.LogAndPrintf("Adding piece index %d to bitfield\n", index)
	c.Bitfield.AddPiece(int(index))
}

func (c *Client) SendKeepAlive() {
	utils.LogAndPrintln("Sending keep-alive message")
	msg := make([]byte, 4)
	_, err := c.Conn.Write(msg)
	if err != nil {
		utils.LogAndPrintf("Error sending keep-alive message: %v\n", err)
	}
}

func (c *Client) SendChoke() {
	utils.LogAndPrintln("Sending choke message")
	msg := Message{MessageID: MSG_CHOKE}
	c.Send(&msg)
}

func (c *Client) SendUnchoke() {
	utils.LogAndPrintln("Sending unchoke message")
	msg := Message{MessageID: MSG_UNCHOKE}
	c.Send(&msg)
}

func (c *Client) SendInterested() {
	utils.LogAndPrintln("Sending interested message")
	msg := Message{MessageID: MSG_INTERESTED}
	c.Send(&msg)
}

func (c *Client) SendNotInterested() {
	utils.LogAndPrintln("Sending not interested message")
	msg := Message{MessageID: MSG_NOT_INTERESTED}
	c.Send(&msg)
}

func (c *Client) SendHave(pieceIndex int) {
	utils.LogAndPrintf("Sending have message for piece index %d\n", pieceIndex)
	msg := Message{MessageID: MSG_HAVE}
	msg.FormatHave(pieceIndex)
	c.Send(&msg)
}

func (c *Client) SendRequest(pieceIndex, offset, blockSize int) {
	utils.LogAndPrintf("Sending request for piece index %d, offset %d, block size %d\n", pieceIndex, offset, blockSize)
	msg := Message{MessageID: MSG_REQUEST}
	msg.FormatRequest(pieceIndex, offset, blockSize)
	c.Send(&msg)
}

// stringToPeerAddress converts a string in the format "IP:Port" to a PeerAddress.
func StringToClient(addr string) (Client, error) {
	// Split the string into IP and port
	var zero Client
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return zero, fmt.Errorf("address must be in the format IP:Port")
	}

	// Parse the IP
	ip := net.ParseIP(parts[0])
	if ip == nil {
		return zero, fmt.Errorf("invalid IP address")
	}

	// Parse the port
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return zero, fmt.Errorf("invalid port number")
	}

	// Check if the port is within the valid range
	if port < 0 || port > 65535 {
		return zero, fmt.Errorf("port number must be between 0 and 65535")
	}

	return New(ip, uint16(port)), nil
}
