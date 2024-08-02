package types

import (
	"crypto/rand"
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
	FileInfo InfoDictionary

	Comment string // Comments about the torrent
	Creator string // Software used to create the torrent
	Date    int    // Date created

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
	if date, ok := dict["creation date"].(int); ok {
		torrent.Date = date
	}

	torrent.FileInfo = *CreateInfoDictionary(infoDict)

	torrent.generatePeerID()
	torrent.Port = 6881

	torrent.Left = torrent.FileInfo.FileLength
	return torrent
}

// When I use any other peer ID than 00112233445566778899 the program crashes. I don't why
func (t *Torrent) generatePeerID() {
	t.PeerID = make([]byte, 20)
	_, err := rand.Read(t.PeerID)
	if err != nil {
		fmt.Println(err)
	}
}

func (t *Torrent) Print() {
	fmt.Printf("Torrent Details:\n")
	fmt.Printf("Tracker URL: %s\n", t.Announce)
	fmt.Printf("Info Hash: %s\n", hex.EncodeToString(t.InfoHash))
	if t.Comment != "" {
		fmt.Printf("Comment: %s\n", t.Comment)
	}
	if t.Creator != "" {
		fmt.Printf("Creator: %s\n", t.Creator)
	}
	if t.Date != 0 {
		fmt.Printf("Creation Date: %d\n", t.Date)
	}
	t.FileInfo.Print()
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
	if t.Comment != "" {
		log.Printf("Comment: %s\n", t.Comment)
	}
	if t.Creator != "" {
		log.Printf("Creator: %s\n", t.Creator)
	}
	if t.Date != 0 {
		log.Printf("Creation Date: %d\n", t.Date)
	}
	t.FileInfo.Log()
	log.Printf("Peer ID: %s\n", hex.EncodeToString(t.PeerID))
	log.Printf("Port: %d\n", t.Port)
	log.Printf("Uploaded: %d bytes\n", t.Uploaded)
	log.Printf("Downloaded: %d bytes\n", t.Downloaded)
	log.Printf("Left: %d bytes\n", t.Left)
}
