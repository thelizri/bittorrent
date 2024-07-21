package types

import (
	"encoding/hex"
	"fmt"
	"karlan/torrent/internal/bencode"
	"karlan/torrent/internal/hash"
	"log"
)

// Torrent holds the decoded information from a .torrent file.
type Torrent struct {
	Announce string // URL of the torrent tracker
	InfoHash []byte // SHA1 hash of the 'info' section of the torrent file
	File     InfoDictionary

	Comment string // Comments about the torrent
	Creator string // Software used to create the torrent

	PeerID []byte // Peer ID for this client
	Port   int    // Port number this client is listening on

	Uploaded   int // Total uploaded data in bytes
	Downloaded int // Total downloaded data in bytes
	Left       int // Number of bytes left to download
}

func CreateTorrentStruct(dict map[string]interface{}) *Torrent {

	infoDict := dict["info"].(map[string]interface{})

	torrent := &Torrent{}
	torrent.Announce = dict["announce"].(string)
	torrent.InfoHash = hash.CalcSha1Hash(bencode.Encode(dict["info"]))

	if comment, ok := dict["comment"].(string); ok {
		torrent.Comment = comment
	}
	if creator, ok := dict["created by"].(string); ok {
		torrent.Creator = creator
	}

	torrent.File = *CreateInfoDictionary(infoDict)

	torrent.PeerID = []byte{0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9}
	torrent.Port = 6881

	torrent.Left = torrent.File.FileLength
	return torrent
}

func (t *Torrent) GetPeerID() string {
	var str string
	for _, b := range t.PeerID {
		str += string(b + '0')
	}
	return str
}

func (t *Torrent) Print() {
	fmt.Printf("Torrent Details:\n")
	fmt.Printf("Tracker URL: %s\n", t.Announce)
	fmt.Printf("Info Hash: %s\n", hex.EncodeToString(t.InfoHash))
	fmt.Printf("Comment: %s\n", t.Comment)
	fmt.Printf("Creator: %s\n", t.Creator)
	t.File.Print()
	fmt.Printf("Peer ID: %s\n", t.GetPeerID())
	fmt.Printf("Port: %d\n", t.Port)
	fmt.Printf("Uploaded: %d bytes\n", t.Uploaded)
	fmt.Printf("Downloaded: %d bytes\n", t.Downloaded)
	fmt.Printf("Left: %d bytes\n", t.Left)
}

func (t *Torrent) Log() {
	log.Printf("Torrent Details:\n")
	log.Printf("Tracker URL: %s\n", t.Announce)
	log.Printf("Info Hash: %s\n", hex.EncodeToString(t.InfoHash))
	log.Printf("Comment: %s\n", t.Comment)
	log.Printf("Creator: %s\n", t.Creator)
	t.File.Log()
	log.Printf("Peer ID: %s\n", t.GetPeerID())
	log.Printf("Port: %d\n", t.Port)
	log.Printf("Uploaded: %d bytes\n", t.Uploaded)
	log.Printf("Downloaded: %d bytes\n", t.Downloaded)
	log.Printf("Left: %d bytes\n", t.Left)
}
