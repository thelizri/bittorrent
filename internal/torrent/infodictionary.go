package torrent

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

const MULTI = "Multi-File Torrent"
const SINGLE = "Single-File Torrent"

type torrentDictionary struct {
	Data            []byte
	Type            string
	Name            string     // Name of the file (for single-file) or root directory (for multi-file)
	FileLength      int        // Length of the file, single file torrent
	PieceLength     int        // Length of each piece
	LastPieceLength int        // Length of last piece
	NumberOfPieces  int        // Number of pieces
	PieceHashes     [][20]byte // SHA1 hashes of each piece
	Files           []fileInfo // List of files for multitorrent
}

type fileInfo struct {
	Length int
	Path   []string
}

func createInfoDictionary(infoDict map[string]interface{}) *torrentDictionary {
	infoDictionaryStruct := &torrentDictionary{}

	infoDictionaryStruct.Name = infoDict["name"].(string)
	infoDictionaryStruct.PieceLength = infoDict["piece length"].(int)
	infoDictionaryStruct.PieceHashes = splitPieceHashes(infoDict["pieces"])
	infoDictionaryStruct.NumberOfPieces = len(infoDictionaryStruct.PieceHashes)

	if length, ok := infoDict["length"].(int); ok {
		// Single-file torrent
		infoDictionaryStruct.Type = SINGLE
		infoDictionaryStruct.FileLength = length
		infoDictionaryStruct.LastPieceLength = infoDictionaryStruct.FileLength - (infoDictionaryStruct.NumberOfPieces-1)*infoDictionaryStruct.PieceLength
		infoDictionaryStruct.Data = make([]byte, infoDictionaryStruct.FileLength)
	} else {
		// Multi-file torrent
		infoDictionaryStruct.Type = MULTI
		files := infoDict["files"].([]interface{})
		totalLength := 0
		var fileStructs []fileInfo

		for _, file := range files {
			fileMap := file.(map[string]interface{})
			length := fileMap["length"].(int)
			pathInterface := fileMap["path"].([]interface{})
			path := make([]string, len(pathInterface))

			for i, p := range pathInterface {
				path[i] = p.(string)
			}

			fileStruct := fileInfo{
				Length: length,
				Path:   path,
			}

			fileStructs = append(fileStructs, fileStruct)
			totalLength += length
		}

		infoDictionaryStruct.Files = fileStructs
		infoDictionaryStruct.FileLength = totalLength
		infoDictionaryStruct.LastPieceLength = totalLength - (infoDictionaryStruct.NumberOfPieces-1)*infoDictionaryStruct.PieceLength
		infoDictionaryStruct.Data = make([]byte, totalLength)
	}

	return infoDictionaryStruct
}

func splitPieceHashes(piece_hashes interface{}) [][20]byte {
	hash, ok := piece_hashes.(string)
	if !ok {
		log.Info("Not a string")
		return nil
	}

	// Ensure that the hash length is a multiple of 20
	if len(hash)%20 != 0 {
		log.Info("String length is not a multiple of 20")
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

func (f *torrentDictionary) GetPieceLength(index int) int {
	if index == f.NumberOfPieces-1 {
		return f.LastPieceLength
	} else {
		return f.PieceLength
	}
}

func (f *torrentDictionary) addPiece(piece []byte, pieceIndex int) {
	offset := f.PieceLength * pieceIndex
	copy(f.Data[offset:offset+len(piece)], piece)
}

func (f *torrentDictionary) log() {
	log.Infof("\tFile Details:\n")
	log.Infof("\tName: %s\n", f.Name)
	log.Infof("\tType: %s\n", f.Type)
	log.Infof("\tFile Length: %d bytes\n", f.FileLength)
	log.Infof("\tPiece Length: %d bytes\n", f.PieceLength)
	log.Infof("\tLast Piece Length: %d bytes\n", f.LastPieceLength)
	log.Infof("\tNumber of Pieces: %d\n", f.NumberOfPieces)
	if f.NumberOfPieces < 10 {
		log.Info("\tPiece Hashes:\n")
		for i, hash := range f.PieceHashes {
			log.Infof("\t\tPiece %d: %x\n", i, hash)
		}
	}

	if f.Type == MULTI {
		log.Info("\tFiles:\n")
		for _, file := range f.Files {
			log.Infof("\t\tPath: %s, Length: %d bytes\n", fmt.Sprintf("%v", file.Path), file.Length)
		}
	}
}
