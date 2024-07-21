package hash

import (
	"bytes"
	"crypto/sha1"
	"fmt"
)

func CalcSha1Hash(encoding string) []byte {
	hash := sha1.Sum([]byte(encoding))
	return hash[:]
}

func CastHashTo2dByteSlice(piece_hashes interface{}) [][]byte {
	hash, ok := piece_hashes.(string)
	if !ok {
		fmt.Println("Not a string")
		return nil
	}

	length := len(hash) / 20
	result := make([][]byte, 0, length)
	for i := 0; i < len(hash); i += 20 {
		result = append(result, []byte(hash[i:i+20]))
	}
	return result
}

// Compare two hash signatures
func ValidatePieceHash(piece []byte, pieceHash []byte) error {
	hash := sha1.Sum(piece)
	if !bytes.Equal(hash[:], pieceHash) {
		return fmt.Errorf("Hashes are not matching:\n Expected: %x,\n Received: %x.\n", pieceHash, hash)
	}
	return nil
}
