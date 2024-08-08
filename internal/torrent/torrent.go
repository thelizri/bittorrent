package torrent

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"karlan/torrent/internal/bencode"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
)

// Torrent holds the decoded information from a .torrent file.
type Torrent struct {
	Announce       string   // URL of the torrent tracker
	InfoHash       [20]byte // SHA1 hash of the 'info' section of the torrent file
	infoDictionary torrentDictionary

	Comment string // Comments about the torrent
	Creator string // Software used to create the torrent
	Date    int    // Date created

	PeerID [20]byte // Peer ID for this client
	Port   int      // Port number this client is listening on

	Uploaded   int // Total uploaded data in bytes
	Downloaded int // Total downloaded data in bytes
	Left       int // Number of bytes left to download

	mutex sync.RWMutex
}

// read and decode torrent file. Returns a torrent struct
func Open(filePath string) *Torrent {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error reading file: %v\n", err))
	}

	decoding, _, err := bencode.Decode(string(bytes), 0)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error decoding file: %v\n", err))
	}

	dict := decoding.(map[string]interface{})
	return createTorrentStruct(dict)
}

func createTorrentStruct(dict map[string]interface{}) *Torrent {

	infoDict := dict["info"].(map[string]interface{})

	torrent := &Torrent{}
	torrent.Announce = dict["announce"].(string)
	torrent.InfoHash = hashInfoDictionary(bencode.Encode(dict["info"]))

	if comment, ok := dict["comment"].(string); ok {
		torrent.Comment = comment
	}
	if creator, ok := dict["created by"].(string); ok {
		torrent.Creator = creator
	}
	if date, ok := dict["creation date"].(int); ok {
		torrent.Date = date
	}

	torrent.infoDictionary = *createInfoDictionary(infoDict)

	torrent.generatePeerID()
	torrent.Port = 6881

	torrent.Left = torrent.infoDictionary.FileLength

	return torrent
}

func hashInfoDictionary(encoding string) [20]byte {
	hash := sha1.Sum([]byte(encoding))
	return hash
}

// When I use any other peer ID than 00112233445566778899 the program crashes. I don't why
func (t *Torrent) generatePeerID() {
	peerID := make([]byte, 20)
	_, err := rand.Read(peerID)
	if err != nil {
		log.Errorf("Generate peer id: %s", err)
	}
	t.PeerID = [20]byte(peerID)
}

func (t *Torrent) Log() {
	log.Infof("Torrent Details:\n")
	log.Infof("Tracker URL: %s\n", t.Announce)
	log.Infof("Info Hash: %s\n", hex.EncodeToString(t.InfoHash[:]))
	if t.Comment != "" {
		log.Infof("Comment: %s\n", t.Comment)
	}
	if t.Creator != "" {
		log.Infof("Creator: %s\n", t.Creator)
	}
	if t.Date != 0 {
		log.Infof("Creation Date: %d\n", t.Date)
	}
	t.infoDictionary.log()
	log.Infof("Peer ID: %s\n", hex.EncodeToString(t.PeerID[:]))
	log.Infof("Port: %d\n", t.Port)
	log.Infof("Uploaded: %d bytes\n", t.Uploaded)
	log.Infof("Downloaded: %d bytes\n", t.Downloaded)
	log.Infof("Left: %d bytes\n", t.Left)
}

func (t *Torrent) GetPieceLength(index int) int {
	return t.infoDictionary.GetPieceLength(index)
}

func (t *Torrent) GetPieceHash(index int) [20]byte {
	return t.infoDictionary.PieceHashes[index]
}

func (t *Torrent) AddPiece(data []byte, index int) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.infoDictionary.addPiece(data, index)
	t.Downloaded += len(data)
	t.Left -= len(data)
}

func (t *Torrent) FinishedDownloading() bool {
	return t.Downloaded == t.infoDictionary.FileLength && t.Left == 0
}

func (t *Torrent) GetNumberOfPieces() int {
	return t.infoDictionary.NumberOfPieces
}

func (t *Torrent) GetName() string {
	return t.infoDictionary.Name
}

func (t *Torrent) GetData() []byte {
	return t.infoDictionary.Data
}
