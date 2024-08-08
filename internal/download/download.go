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
	defer cl.Conn.Close()
	defer wg.Done()

	for !q.IsEmpty() {
		pieceIndex, err := q.Dequeue()
		if err != nil {
			log.Warnf("Failed to dequeue piece: %v", err)
			q.Enqueue(pieceIndex)
			return
		}

		if !cl.HasPiece(pieceIndex) {
			log.Debugf("Client does not have piece %d, re-enqueueing", pieceIndex)
			q.Enqueue(pieceIndex)
			continue
		}

		pieceProgress, err := DownloadPiece(cl, pieceIndex, t.GetPieceLength(pieceIndex), t.GetPieceHash(pieceIndex))
		if err != nil {
			log.Warnf("Failed to download piece %d: %v", pieceIndex, err)
			q.Enqueue(pieceIndex)
			return
		}

		t.AddPiece(pieceProgress.Data, pieceIndex)
		log.Infof("Successfully downloaded and added piece %d", pieceIndex)
	}
}

func DownloadPiece(cl *client.Client, pieceIndex, pieceSize int, pieceHash [20]byte) (PieceProgress, error) {
	log.Infof("Starting download of piece: Index=%d, Size=%d", pieceIndex, pieceSize)
	var zero PieceProgress
	p := PieceProgress{
		Index: pieceIndex,
		Size:  pieceSize,
		Hash:  pieceHash,
		Data:  make([]byte, pieceSize),
	}

	log.Debug("Sending interested message")
	cl.SendInterested()

	var offset, end int
	for p.Downloaded < p.Size {
		log.Debugf("Download progress: %d/%d bytes", p.Downloaded, p.Size)
		if !cl.Choked {
			blockSize := BLOCK_SIZE
			if blockSize > p.Size-p.Downloaded {
				blockSize = p.Size - p.Downloaded
			}
			offset = p.Downloaded
			log.Debugf("Requesting block: Offset=%d, BlockSize=%d", offset, blockSize)
			cl.SendRequest(p.Index, offset, blockSize)
		} else {
			log.Debug("Client is choked, waiting for unchoke message")
		}

		data, err := read(cl)
		if err != nil {
			log.Errorf("Error reading data: %v", err)
			return zero, err
		}

		if data == nil {
			log.Debug("No data received, continuing")
			continue
		}

		end = offset + len(data)
		log.Tracef("Copying block to data: Offset=%d, End=%d, BlockSize=%d", offset, end, len(data))
		copy(p.Data[offset:end], data)
		p.Downloaded += len(data)
	}

	log.Debug("Validating piece hash")
	err := p.validatePiece()
	if err != nil {
		log.Errorf("Piece hash validation failed: %v", err)
		return zero, err
	}

	log.Info("Hashes match, piece download complete")
	fmt.Printf("Downloaded piece %d\n", p.Index)
	return p, nil
}

func read(cl *client.Client) ([]byte, error) {
	if cl == nil {
		panic("Client is nil")
	}

	log.Debug("Reading message from client")
	msg, err := cl.Read()
	if err != nil {
		log.Errorf("Error reading message: %v", err)
		return nil, err
	}

	if msg == nil {
		log.Debug("Received nil message")
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
		log.Debug("Received Interested message")

	case client.MSG_NOT_INTERESTED:
		log.Debug("Received Not Interested message")

	case client.MSG_HAVE:
		log.Debug("Received Have message")
		cl.AddPiece(msg)

	case client.MSG_BITFIELD:
		log.Debug("Received Bitfield message")

	case client.MSG_REQUEST:
		log.Debug("Received Request message")

	case client.MSG_PIECE:
		log.Debug("Received Piece message")
		return msg.Payload[8:], nil

	case client.MSG_CANCEL:
		log.Debug("Received Cancel message")

	default:
		log.Warnf("Received unknown message ID: %d", msg.MessageID)
		return nil, fmt.Errorf("received unknown message")
	}

	return nil, nil
}
