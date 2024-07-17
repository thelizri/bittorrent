package fileio

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"karlan/torrent/internal/bencode"
	"karlan/torrent/internal/hash"
	"karlan/torrent/internal/types"
)

func WriteToRelativePath(filePath string, data []byte) {
	// Write the byte slice to a file
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Combine the current directory with the file path
	filePath = filepath.Clean(filePath)
	fullPath := filepath.Join(dir, filePath)

	// Extract the directory from the full path
	dirPath := filepath.Dir(fullPath)

	// Create the directory if it doesn't exist
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		log.Fatal(err)
	}

	// Write the file
	err = os.WriteFile(fullPath, data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func WriteToAbsolutePath(filePath string, data []byte) {
	// Write the byte slice to a file

	// Combine the current directory with the file path
	filePath = filepath.Clean(filePath)

	// Extract the directory from the full path
	dirPath := filepath.Dir(filePath)

	// Create the directory if it doesn't exist
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		log.Fatal(err)
	}

	// Write the file
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

// read and decode torrent file. Returns a torrent struct
func ReadTorrentFile(filePath string) types.Torrent {

	bytes, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return types.Torrent{}
	}

	decoding, _, err := bencode.Decode(string(bytes), 0)
	if err != nil {
		fmt.Println(err)
		return types.Torrent{}
	}

	dict := decoding.(map[string]interface{})
	info_dict := dict["info"].(map[string]interface{})

	torrent := types.Torrent{}
	torrent.Announce = dict["announce"].(string)
	torrent.InfoHash = hash.CalcSha1Hash(bencode.Encode(dict["info"]))
	torrent.PeerID = []byte{0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9}
	torrent.Port = 6881
	torrent.FileLength = info_dict["length"].(int)
	torrent.PieceLength = info_dict["piece length"].(int)
	torrent.Left = torrent.FileLength
	torrent.PieceHashes = hash.CastHashTo2dByteSlice(info_dict["pieces"])
	torrent.NumberOfPieces = len(torrent.PieceHashes)
	torrent.LastPieceLength = torrent.FileLength - (torrent.NumberOfPieces-1)*torrent.PieceLength
	return torrent
}
