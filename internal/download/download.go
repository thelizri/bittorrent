package download

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"karlan/torrent/internal/client"
	"karlan/torrent/internal/queue"
	"karlan/torrent/internal/torrent"
	"karlan/torrent/internal/utils"
	"sync"
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
			utils.LogAndPrintf("Download File: %s", err)
			q.Enqueue(pieceIndex)
			return
		}

		if !cl.HasPiece(pieceIndex) {
			q.Enqueue(pieceIndex)
			continue
		}

		pieceProgress, err := DownloadPiece(cl, pieceIndex, t.GetPieceLength(pieceIndex), t.GetPieceHash(pieceIndex))
		if err != nil {
			utils.LogAndPrintf("Download File: %s", err)
			q.Enqueue(pieceIndex)
			return
		}

		t.AddPiece(pieceProgress.Data, pieceIndex)
	}
}

func DownloadPiece(cl *client.Client, pieceIndex, pieceSize int, pieceHash [20]byte) (PieceProgress, error) {
	utils.LogAndPrintf("Starting download of piece: Index=%d, Size=%d\n", pieceIndex, pieceSize)
	var zero PieceProgress
	p := PieceProgress{
		Index: pieceIndex,
		Size:  pieceSize,
		Hash:  pieceHash,
		Data:  make([]byte, pieceSize),
	}

	utils.LogAndPrintln("Sending interested message")
	cl.SendInterested()

	var offset, end int
	for p.Downloaded < p.Size {
		utils.LogAndPrintf("Download progress: %d/%d bytes\n", p.Downloaded, p.Size)
		if !cl.Choked {
			blockSize := BLOCK_SIZE
			if blockSize > p.Size-p.Downloaded {
				blockSize = p.Size - p.Downloaded
			}
			offset = p.Downloaded
			utils.LogAndPrintf("Requesting block: Offset=%d, BlockSize=%d\n", offset, blockSize)
			cl.SendRequest(p.Index, offset, blockSize)
		} else {
			utils.LogAndPrintln("Client is choked, waiting for unchoke message")
		}

		data, err := read(cl)
		if err != nil {
			utils.LogAndPrintf("Error reading data: %v\n", err)
			return zero, err
		}

		if data == nil {
			utils.LogAndPrintln("No data received, continuing")
			continue
		}

		end = offset + len(data)
		utils.LogAndPrintf("Copying block to data: Offset=%d, End=%d, BlockSize=%d\n", offset, end, len(data))
		copy(p.Data[offset:end], data)
		p.Downloaded += len(data)
	}

	utils.LogAndPrintln("Validating piece hash")
	err := p.validatePiece()
	if err != nil {
		utils.LogAndPrintf("Piece hash validation failed: %v\n", err)
		return zero, err
	}

	utils.LogAndPrintln("Hashes match, piece download complete")
	return p, nil
}

func read(cl *client.Client) ([]byte, error) {
	if cl == nil {
		panic("Client is nil")
	}

	utils.LogAndPrintln("Reading message from client")
	msg, err := cl.Read()
	if err != nil {
		utils.LogAndPrintf("Error reading message: %v\n", err)
		return nil, err
	}

	if msg == nil {
		return nil, nil
	}

	switch msg.MessageID {
	case client.MSG_CHOKE:
		utils.LogAndPrintln("Received Choke message")
		cl.Choked = true

	case client.MSG_UNCHOKE:
		utils.LogAndPrintln("Received Unchoke message")
		cl.Choked = false

	case client.MSG_INTERESTED:
		utils.LogAndPrintln("Received Interested message")
		// Handle interested message

	case client.MSG_NOT_INTERESTED:
		utils.LogAndPrintln("Received Not Interested message")
		// Handle not interested message

	case client.MSG_HAVE:
		utils.LogAndPrintln("Received Have message")
		cl.AddPiece(msg)

	case client.MSG_BITFIELD:
		utils.LogAndPrintln("Received Bitfield message")
		// Handle bitfield message if necessary

	case client.MSG_REQUEST:
		utils.LogAndPrintln("Received Request message")
		// Handle request message if necessary

	case client.MSG_PIECE:
		utils.LogAndPrintln("Received Piece message")
		return msg.Payload[8:], nil

	case client.MSG_CANCEL:
		utils.LogAndPrintln("Received Cancel message")
		// Handle cancel message if necessary

	default:
		utils.LogAndPrintf("Received unknown message ID: %d\n", msg.MessageID)
		return nil, fmt.Errorf("received unknown message\n")
	}

	return nil, nil
}
