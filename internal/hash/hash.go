package hash

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"karlan/torrent/internal/utils"
)

func CastHashTo2dByteSlice(piece_hashes interface{}) [][20]byte {
	hash, ok := piece_hashes.(string)
	if !ok {
		utils.LogAndPrintln("Not a string")
		return nil
	}

	// Ensure that the hash length is a multiple of 20
	if len(hash)%20 != 0 {
		utils.LogAndPrintln("String length is not a multiple of 20")
		return nil
	}

	length := len(hash) / 20
	result := make([][20]byte, 0, length)

	for i := 0; i < len(hash); i += 20 {
		var temp [20]byte
		copy(temp[:], hash[i:i+20])
		result = append(result, temp)
	}

	return result
}

// Compare two hash signatures
func ValidatePieceHash(piece []byte, pieceHash [20]byte) error {
	hash := sha1.Sum(piece)
	if !bytes.Equal(hash[:], pieceHash[:]) {
		return fmt.Errorf("Hashes are not matching:\n Expected: %x,\n Received: %x.\n", pieceHash, hash)
	}
	return nil
}
