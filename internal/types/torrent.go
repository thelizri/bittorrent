package types

import (
	"encoding/hex"
	"fmt"
	"log"
)

// Torrent holds the decoded information from a .torrent file.
type Torrent struct {
	Announce        string   // URL of the torrent tracker
	InfoHash        []byte   // SHA1 hash of the 'info' section of the torrent file
	FileLength      int      // Length of the file to be downloaded
	PieceLength     int      // Length of each piece
	LastPieceLength int      // Length of last piece
	NumberOfPieces  int      // Number of pieces
	PieceHashes     [][]byte // SHA1 hashes of each piece
	PeerID          []byte   // Peer ID for this client
	Port            int      // Port number this client is listening on
	Uploaded        int      // Total uploaded data in bytes
	Downloaded      int      // Total downloaded data in bytes
	Left            int      // Number of bytes left to download
}

func (t *Torrent) GetPeerID() string {
	var str string
	for _, b := range t.PeerID {
		str += string(b + '0')
	}
	return str
}

func (t *Torrent) GetPieceLength(index int) int {
	if index == t.NumberOfPieces-1 {
		return t.LastPieceLength
	} else {
		return t.PieceLength
	}
}

func (t *Torrent) Print() {
	fmt.Printf("Torrent Details:\n")
	fmt.Printf("Tracker URL: %s\n", t.Announce)
	fmt.Printf("Info Hash: %s\n", hex.EncodeToString(t.InfoHash))
	fmt.Printf("File Length: %d bytes\n", t.FileLength)
	fmt.Printf("Piece Length: %d bytes\n", t.PieceLength)
	fmt.Printf("Last Piece Length: %d bytes\n", t.LastPieceLength)
	fmt.Printf("Number of Pieces: %d\n", t.NumberOfPieces)
	fmt.Printf("Piece Hashes:\n")
	for i, hash := range t.PieceHashes {
		fmt.Printf("  Piece %d: %x\n", i, hash)
	}
	fmt.Printf("Peer ID: %s\n", hex.EncodeToString(t.PeerID))
	fmt.Printf("Port: %d\n", t.Port)
	fmt.Printf("Uploaded: %d bytes\n", t.Uploaded)
	fmt.Printf("Downloaded: %d bytes\n", t.Downloaded)
	fmt.Printf("Left: %d bytes\n", t.Left)
}

func (t *Torrent) Log() {
	log.Printf("Torrent Details:\n")
	log.Printf("Tracker URL: %s\n", t.Announce)
	log.Printf("Info Hash: %s\n", hex.EncodeToString(t.InfoHash))
	log.Printf("File Length: %d bytes\n", t.FileLength)
	log.Printf("Piece Length: %d bytes\n", t.PieceLength)
	log.Printf("Last Piece Length: %d bytes\n", t.LastPieceLength)
	log.Printf("Number of Pieces: %d\n", t.NumberOfPieces)
	log.Printf("Piece Hashes:\n")
	for i, hash := range t.PieceHashes {
		log.Printf("  Piece %d: %x\n", i, hash)
	}
	log.Printf("Peer ID: %s\n", hex.EncodeToString(t.PeerID))
	log.Printf("Port: %d\n", t.Port)
	log.Printf("Uploaded: %d bytes\n", t.Uploaded)
	log.Printf("Downloaded: %d bytes\n", t.Downloaded)
	log.Printf("Left: %d bytes\n\n", t.Left)
}
