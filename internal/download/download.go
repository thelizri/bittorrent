package download

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"karlan/torrent/internal/client"
	"karlan/torrent/internal/queue"
	"karlan/torrent/internal/torrent"
	"sync"

	log "github.com/sirupsen/logrus"
)

type PieceProgress struct {
	Downloaded int
	Uploaded   int
	Index      int
	Size       int
	Hash       [20]byte
	Data       []byte
}

func (p *PieceProgress) validatePiece() error {
	hash := sha1.Sum(p.Data)
	if !bytes.Equal(hash[:], p.Hash[:]) {
		return fmt.Errorf("Hashes are not matching:\n Expected: %x,\n Received: %x.\n", p.Hash, hash)
	}
	return nil
}

const BLOCK_SIZE int = 16 * 1024

func DownloadFile(cl *client.Client, t *torrent.Torrent, q *queue.Queue, wg *sync.WaitGroup) {
	// Dequeue
	// Check if client has piece
	// If not enqueue
	// Request piece
	// If error, return nothing
	// Return piece via channel
	// Go back to top

	defer cl.Conn.Close()
	defer wg.Done()

	for !q.IsEmpty() {
		pieceIndex, err := q.Dequeue()
		if err != nil {
			log.Info("Download File: %s", err)
			q.Enqueue(pieceIndex)
			return
		}

		if !cl.HasPiece(pieceIndex) {
			q.Enqueue(pieceIndex)
			continue
		}

		pieceProgress, err := DownloadPiece(cl, pieceIndex, t.GetPieceLength(pieceIndex), t.GetPieceHash(pieceIndex))
		if err != nil {
			log.Info("Download File: %s", err)
			q.Enqueue(pieceIndex)
			return
		}

		t.AddPiece(pieceProgress.Data, pieceIndex)
	}
}

func DownloadPiece(cl *client.Client, pieceIndex, pieceSize int, pieceHash [20]byte) (PieceProgress, error) {
	log.Info("Starting download of piece: Index=%d, Size=%d\n", pieceIndex, pieceSize)
	var zero PieceProgress
	p := PieceProgress{
		Index: pieceIndex,
		Size:  pieceSize,
		Hash:  pieceHash,
		Data:  make([]byte, pieceSize),
	}

	log.Info("Sending interested message")
	cl.SendInterested()

	var offset, end int
	for p.Downloaded < p.Size {
		log.Info("Download progress: %d/%d bytes\n", p.Downloaded, p.Size)
		if !cl.Choked {
			blockSize := BLOCK_SIZE
			if blockSize > p.Size-p.Downloaded {
				blockSize = p.Size - p.Downloaded
			}
			offset = p.Downloaded
			log.Info("Requesting block: Offset=%d, BlockSize=%d\n", offset, blockSize)
			cl.SendRequest(p.Index, offset, blockSize)
		} else {
			log.Info("Client is choked, waiting for unchoke message")
		}

		data, err := read(cl)
		if err != nil {
			log.Info("Error reading data: %v\n", err)
			return zero, err
		}

		if data == nil {
			log.Info("No data received, continuing")
			continue
		}

		end = offset + len(data)
		log.Info("Copying block to data: Offset=%d, End=%d, BlockSize=%d\n", offset, end, len(data))
		copy(p.Data[offset:end], data)
		p.Downloaded += len(data)
	}

	log.Info("Validating piece hash")
	err := p.validatePiece()
	if err != nil {
		log.Info("Piece hash validation failed: %v\n", err)
		return zero, err
	}

	log.Info("Hashes match, piece download complete")
	return p, nil
}

func read(cl *client.Client) ([]byte, error) {
	if cl == nil {
		panic("Client is nil")
	}

	log.Info("Reading message from client")
	msg, err := cl.Read()
	if err != nil {
		log.Info("Error reading message: %v\n", err)
		return nil, err
	}

	if msg == nil {
		return nil, nil
	}

	switch msg.MessageID {
	case client.MSG_CHOKE:
		log.Info("Received Choke message")
		cl.Choked = true

	case client.MSG_UNCHOKE:
		log.Info("Received Unchoke message")
		cl.Choked = false

	case client.MSG_INTERESTED:
		log.Info("Received Interested message")
		// Handle interested message

	case client.MSG_NOT_INTERESTED:
		log.Info("Received Not Interested message")
		// Handle not interested message

	case client.MSG_HAVE:
		log.Info("Received Have message")
		cl.AddPiece(msg)

	case client.MSG_BITFIELD:
		log.Info("Received Bitfield message")
		// Handle bitfield message if necessary

	case client.MSG_REQUEST:
		log.Info("Received Request message")
		// Handle request message if necessary

	case client.MSG_PIECE:
		log.Info("Received Piece message")
		return msg.Payload[8:], nil

	case client.MSG_CANCEL:
		log.Info("Received Cancel message")
		// Handle cancel message if necessary

	default:
		log.Info("Received unknown message ID: %d\n", msg.MessageID)
		return nil, fmt.Errorf("received unknown message\n")
	}

	return nil, nil
}
