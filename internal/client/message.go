package client

import (
	"encoding/binary"

	log "github.com/sirupsen/logrus"
)

// Constants representing different message IDs used in the BitTorrent protocol.
type MessageID byte

const (
	MSG_CHOKE          MessageID = 0 // Indicates the sender will not send any more pieces
	MSG_UNCHOKE        MessageID = 1 // Indicates the sender will now allow the receiver to request pieces
	MSG_INTERESTED     MessageID = 2 // Indicates the sender wants to download pieces from the recipient
	MSG_NOT_INTERESTED MessageID = 3 // Indicates the sender does not want to download pieces from the recipient
	MSG_HAVE           MessageID = 4 // Indicates the sender has downloaded a specific piece
	MSG_BITFIELD       MessageID = 5 // Contains a bitfield representing the pieces the sender has
	MSG_REQUEST        MessageID = 6 // Requests a specific piece of data
	MSG_PIECE          MessageID = 7 // Contains the actual data of the piece being sent
	MSG_CANCEL         MessageID = 8 // Cancels a previously sent request
)

func (id MessageID) String() string {
	switch id {
	case 0:
		return "Choke"
	case 1:
		return "Unchoke"
	case 2:
		return "Interested"
	case 3:
		return "Not Interested"
	case 4:
		return "Have"
	case 5:
		return "Bitfield"
	case 6:
		return "Request"
	case 7:
		return "Piece"
	case 8:
		return "Cancel"
	default:
		return "Unknown"
	}
}

type Message struct {
	MessageID MessageID // ID of the message type
	Payload   []byte    // The payload of the message
}

func (p *Message) length() uint32 {
	return uint32(1 + len(p.Payload))
}

// <length:4><messageid:1><payload:variable>
func (p *Message) Serialize() []byte {
	buffer := make([]byte, 5+len(p.Payload))
	binary.BigEndian.PutUint32(buffer, p.length())
	buffer[4] = byte(p.MessageID)

	if p.Payload != nil {
		copy(buffer[5:], p.Payload)
	}

	return buffer
}

func (p *Message) Log() {
	log.Printf("Length: %v, ID: %v", p.length(), p.MessageID)
}

func (p *Message) FormatHave(index int) {
	p.Payload = make([]byte, 4)
	binary.BigEndian.PutUint32(p.Payload, uint32(index))
}

func (p *Message) FormatRequest(pieceIndex, offset, blockSize int) {
	p.Payload = make([]byte, 12)
	binary.BigEndian.PutUint32(p.Payload[0:4], uint32(pieceIndex))
	binary.BigEndian.PutUint32(p.Payload[4:8], uint32(offset))
	binary.BigEndian.PutUint32(p.Payload[8:12], uint32(blockSize))
}

func (p *Message) FormatPiece(pieceIndex, offset int, piece []byte) {
	p.Payload = make([]byte, 8+len(p.Payload))
	binary.BigEndian.PutUint32(p.Payload[0:4], uint32(pieceIndex))
	binary.BigEndian.PutUint32(p.Payload[4:8], uint32(offset))
	copy(p.Payload[8:], piece)
}

func (p *Message) FormatCancel(pieceIndex, offset, blockSize int) {
	p.Payload = make([]byte, 12)
	binary.BigEndian.PutUint32(p.Payload[0:4], uint32(pieceIndex))
	binary.BigEndian.PutUint32(p.Payload[4:8], uint32(offset))
	binary.BigEndian.PutUint32(p.Payload[8:12], uint32(blockSize))
}
