package torrent

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"karlan/torrent/internal/bencode"
	"karlan/torrent/internal/hash"
	"karlan/torrent/internal/utils"
	"os"
	"sync"
)

// Torrent holds the decoded information from a .torrent file.
type Torrent struct {
	Announce string // URL of the torrent tracker
	InfoHash []byte // SHA1 hash of the 'info' section of the torrent file
	FileInfo TorrentDictionary

	Comment string // Comments about the torrent
	Creator string // Software used to create the torrent
	Date    int    // Date created

	PeerID []byte // Peer ID for this client
	Port   int    // Port number this client is listening on

	Uploaded   int // Total uploaded data in bytes
	Downloaded int // Total downloaded data in bytes
	Left       int // Number of bytes left to download

	mutex sync.RWMutex
}

// read and decode torrent file. Returns a torrent struct
func Open(filePath string) *Torrent {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		utils.LogPrintExit(fmt.Sprintf("Error reading file: %v\n", err))
	}

	decoding, _, err := bencode.Decode(string(bytes), 0)
	if err != nil {
		utils.LogPrintExit(fmt.Sprintf("Error decoding file: %v\n", err))
	}

	dict := decoding.(map[string]interface{})
	return createTorrentStruct(dict)
}

func createTorrentStruct(dict map[string]interface{}) *Torrent {

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
		utils.LogAndPrintf("Generate peer id: %s", err)
	}
}

func (t *Torrent) Print() {
	utils.LogSeparator()
	utils.LogAndPrintf("Torrent Details:\n")
	utils.LogAndPrintf("Tracker URL: %s\n", t.Announce)
	utils.LogAndPrintf("Info Hash: %s\n", hex.EncodeToString(t.InfoHash))
	if t.Comment != "" {
		utils.LogAndPrintf("Comment: %s\n", t.Comment)
	}
	if t.Creator != "" {
		utils.LogAndPrintf("Creator: %s\n", t.Creator)
	}
	if t.Date != 0 {
		utils.LogAndPrintf("Creation Date: %d\n", t.Date)
	}
	t.FileInfo.Print()
	utils.LogAndPrintf("Peer ID: %s\n", hex.EncodeToString(t.PeerID))
	utils.LogAndPrintf("Port: %d\n", t.Port)
	utils.LogAndPrintf("Uploaded: %d bytes\n", t.Uploaded)
	utils.LogAndPrintf("Downloaded: %d bytes\n", t.Downloaded)
	utils.LogAndPrintf("Left: %d bytes\n", t.Left)
	utils.LogSeparator()
}

func (t *Torrent) GetPieceLength(index int) int {
	return t.FileInfo.GetPieceLength(index)
}

func (t *Torrent) GetPieceHash(index int) [20]byte {
	return t.FileInfo.PieceHashes[index]
}

func (t *Torrent) AddPiece(data []byte, index int) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.FileInfo.AddPiece(data, index)
	t.Downloaded += len(data)
	t.Left -= len(data)
}

func (t *Torrent) FinishedDownloading() bool {
	return t.Downloaded == t.FileInfo.FileLength && t.Left == 0
}

func (t *Torrent) VerifyIntegrityOfEachPiece() bool {
	return t.FileInfo.VerifyIntegrityOfEachPiece()
}
