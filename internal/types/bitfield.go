package types

// Represents the pieces the peer has
type Bitfield []byte

func (b Bitfield) HasPiece(index int) bool {
	byteIndex := index / 8
	offset := index % 8
	return b[byteIndex]>>(7-offset)&1 != 0
}
