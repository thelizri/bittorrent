package network

import (
	"encoding/binary"
	"log"
	"math"
	"net"

	"karlan/torrent/internal/hash"
	"karlan/torrent/internal/types"
	"karlan/torrent/internal/utils"
)

func FetchFile(conn net.Conn, queue *types.Queue, file *types.InfoDictionary) {
	for !queue.IsEmpty() {
		pieceIndex, err := queue.Dequeue()
		if err != nil {
			log.Printf("Error: %v\n", err)
		}

		utils.LogSeparator()
		piece := FetchPiece(conn, file.GetPieceLength(pieceIndex), pieceIndex)
		if err = hash.ValidatePieceHash(piece, file.PieceHashes[pieceIndex]); err != nil {
			queue.Enqueue(pieceIndex)
			log.Printf("Error: %v\n", err)
			return
		}
		file.AddPiece(piece, pieceIndex)
	}
}

func FetchPiece(conn net.Conn, pieceLength, pieceIndex int) []byte {
	log.Printf("Downloading piece %d, length %d\n", pieceIndex, pieceLength)
	pieceData := make([]byte, pieceLength)
	blockSize := BLOCK_SIZE
	totalBlocks := int(math.Ceil(float64(pieceLength) / float64(blockSize)))
	log.Printf("Total blocks: %v\n", totalBlocks)
	lastBlockSize := pieceLength - (totalBlocks-1)*blockSize
	offset := 0

	for i := 0; i < totalBlocks-1; i++ {
		log.Printf("Fetching block %d/%d\n", i, totalBlocks)
		block := fetchBlock(conn, pieceIndex, i, blockSize, offset)
		copy(pieceData[offset:offset+blockSize], block)
		offset += blockSize
	}

	block := fetchBlock(conn, pieceIndex, totalBlocks-1, lastBlockSize, offset)
	copy(pieceData[offset:], block)

	log.Printf("Downloaded piece %d\n", pieceIndex)
	return pieceData
}

func fetchBlock(conn net.Conn, pieceIndex, blockNumber, blockSize, offset int) []byte {
	log.Printf("\tFetching piece %v, block %v, size %v\n", pieceIndex, blockNumber, blockSize)
	msg := types.Message{MessageID: types.MSG_REQUEST, Payload: createPayload(pieceIndex, blockNumber, blockSize, offset)}
	log.Printf("\tRequesting piece %v, block %v, size %v\n", pieceIndex, blockNumber, blockSize)
	SendMessageToPeer(conn, &msg)
	log.Printf("\tAwaiting response for piece %v, block %v\n", pieceIndex, blockNumber)
	msg, err := ListenForPeerMessage(conn, types.MSG_PIECE)
	if err != nil {
		log.Println("\tPeer message error:", err)
		return nil
	}
	log.Printf("\tReceived: ID %v, Length %v\n", msg.MessageID, msg.GetLength())
	pieceIndex = int(binary.BigEndian.Uint32(msg.Payload[:4]))
	blockOffset := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	log.Printf("\tIndex: %v, Offset: %v, Length: %v\n", pieceIndex, blockOffset, len(msg.Payload[8:]))
	return msg.Payload[8:]
}

func createPayload(pieceIndex, blockNumber, blockSize, offset int) []byte {
	log.Printf("\t\tCreating payload for piece %v, block %v, size %v, offset %v\n", pieceIndex, blockNumber, blockSize, offset)
	buf := make([]byte, 12)
	binary.BigEndian.PutUint32(buf[0:4], uint32(pieceIndex))
	binary.BigEndian.PutUint32(buf[4:8], uint32(offset))
	binary.BigEndian.PutUint32(buf[8:12], uint32(blockSize))
	log.Printf("\t\tPayload: %v\n", buf)
	return buf
}
