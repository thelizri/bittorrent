package types

import (
	"encoding/binary"
	"log"
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

func (p *Message) GetLength() uint32 {
	return uint32(1 + len(p.Payload))
}

func (p *Message) GetBytes() []byte {
	buffer := make([]byte, 5+len(p.Payload))
	binary.BigEndian.PutUint32(buffer, p.GetLength())
	buffer[4] = byte(p.MessageID)
	copy(buffer[5:], p.Payload)

	return buffer
}

func (p *Message) Log() {
	log.Printf("Length: %v, ID: %v", p.GetLength(), p.MessageID)
}
