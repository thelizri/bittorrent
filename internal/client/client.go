package client

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
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
	log.Debugf("Creating new client: IP=%v, Port=%v", ip, port)
	return Client{Choked: true, IP: ip, Port: port}
}

func (c *Client) Address() string {
	address := fmt.Sprintf("%v:%v", c.IP, c.Port)
	log.Debugf("Client address: %s", address)
	return address
}

func (c *Client) Init(infoHash, peerID [20]byte) error {
	log.Infof("Initializing client: InfoHash=%x, PeerID=%x", infoHash, peerID)
	handshake := newHandshake(infoHash, peerID)
	handshake.log()

	response, conn, err := sendHandshake(c.Address(), handshake)
	if err != nil {
		log.Errorf("Error during handshake: %v", err)
		return err
	}
	response.log()

	c.Conn = conn
	c.PeerID = response.PeerID
	log.Infof("Connected to peer: PeerID=%x", c.PeerID)

	c.Bitfield, err = c.receiveBitfield()
	if err != nil {
		log.Errorf("Error receiving bitfield: %v", err)
		return err
	}
	log.Info("Bitfield received successfully")
	return nil
}

// Peer messages consist of a message length prefix (4 bytes), message id (1 byte), and a payload (variable size).
func (c *Client) receiveBitfield() (*Bitfield, error) {
	log.Debug("Receiving bitfield from peer")
	conn := c.Conn

	// Set timeout, Keep alive message is sent every 2 minutes
	conn.SetReadDeadline(time.Now().Add(2 * time.Minute))
	defer conn.SetReadDeadline(time.Time{})

	buffer := make([]byte, 4)
	if _, err := io.ReadFull(conn, buffer); err != nil {
		return nil, fmt.Errorf("cannot read message length: %v", err)
	}

	messageLength := binary.BigEndian.Uint32(buffer)
	log.Debugf("Bitfield message length: %d", messageLength)
	buffer = make([]byte, messageLength)
	if _, err := io.ReadFull(conn, buffer); err != nil {
		return nil, fmt.Errorf("cannot read message payload: %v", err)
	}

	messageID := MessageID(buffer[0])
	log.Debugf("Received message ID: %v", messageID)
	if messageID != MSG_BITFIELD {
		return nil, fmt.Errorf("expected bitfield but got message id %v", messageID)
	}
	var bitfield Bitfield = buffer[1:]
	log.Debugf("Bitfield data: %x", bitfield)

	return &bitfield, nil
}

// Peer messages consist of a message length prefix (4 bytes), message id (1 byte), and a payload (variable size).
func (c *Client) Read() (*Message, error) {
	log.Debug("Reading message from client")
	conn := c.Conn

	// Set timeout, Keep alive message is sent every 2 minutes
	conn.SetReadDeadline(time.Now().Add(2 * time.Minute))
	defer conn.SetReadDeadline(time.Time{})

	buffer := make([]byte, 4)
	if _, err := io.ReadFull(conn, buffer); err != nil {
		return nil, fmt.Errorf("cannot read message length: %v", err)
	}

	messageLength := binary.BigEndian.Uint32(buffer)
	log.Debugf("Message length: %d", messageLength)
	if messageLength == 0 {
		log.Debug("Received keep-alive message")
		return nil, nil
	}

	buffer = make([]byte, messageLength)
	if _, err := io.ReadFull(conn, buffer); err != nil {
		return nil, fmt.Errorf("cannot read message payload: %v", err)
	}

	messageID := MessageID(buffer[0])
	payload := buffer[1:]
	log.Debugf("Received message ID: %v, Payload length: %d", messageID, len(payload))

	peer_message := &Message{MessageID: messageID, Payload: payload}
	return peer_message, nil
}

func (c *Client) Send(message *Message) error {
	log.Debugf("Sending message to peer. Message ID: %v", message.MessageID.String())
	if c.Conn == nil {
		log.Error("Connection is nil, cannot send message")
		return fmt.Errorf("connection is nil")
	}
	_, err := c.Conn.Write(message.Serialize())
	if err != nil {
		log.Errorf("Error sending message: %v", err)
		return err
	}

	log.Debug("Message sent successfully")
	return nil
}

func (c *Client) HasPiece(index int) bool {
	hasPiece := c.Bitfield.HasPiece(index)
	log.Debugf("Checking if client has piece %d: %v", index, hasPiece)
	return hasPiece
}

func (c *Client) AddPiece(message *Message) {
	index := binary.BigEndian.Uint32(message.Payload)
	log.Debugf("Adding piece index %d to bitfield", index)
	c.Bitfield.AddPiece(int(index))
}

func (c *Client) SendKeepAlive() {
	log.Debug("Sending keep-alive message")
	msg := make([]byte, 4)
	_, err := c.Conn.Write(msg)
	if err != nil {
		log.Warnf("Error sending keep-alive message: %v", err)
	}
}

func (c *Client) SendChoke() {
	log.Info("Sending choke message")
	msg := Message{MessageID: MSG_CHOKE}
	c.Send(&msg)
}

func (c *Client) SendUnchoke() {
	log.Info("Sending unchoke message")
	msg := Message{MessageID: MSG_UNCHOKE}
	c.Send(&msg)
}

func (c *Client) SendInterested() {
	log.Info("Sending interested message")
	msg := Message{MessageID: MSG_INTERESTED}
	c.Send(&msg)
}

func (c *Client) SendNotInterested() {
	log.Info("Sending not interested message")
	msg := Message{MessageID: MSG_NOT_INTERESTED}
	c.Send(&msg)
}

func (c *Client) SendHave(pieceIndex int) {
	log.Infof("Sending have message for piece index %d", pieceIndex)
	msg := Message{MessageID: MSG_HAVE}
	msg.FormatHave(pieceIndex)
	c.Send(&msg)
}

func (c *Client) SendRequest(pieceIndex, offset, blockSize int) {
	log.Infof("Sending request for piece index %d, offset %d, block size %d", pieceIndex, offset, blockSize)
	msg := Message{MessageID: MSG_REQUEST}
	msg.FormatRequest(pieceIndex, offset, blockSize)
	c.Send(&msg)
}

// StringToClient converts a string in the format "IP:Port" to a Client.
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
