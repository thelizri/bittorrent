package types

import (
	"fmt"
	"karlan/torrent/internal/hash"
	"log"
	"sync"
)

const MULTI = "Multi-File Torrent"
const SINGLE = "Single-File Torrent"

type InfoDictionary struct {
	Data            []byte
	Type            string
	Name            string   // Name of the file (for single-file) or root directory (for multi-file)
	FileLength      int      // Length of the file, single file torrent
	PieceLength     int      // Length of each piece
	LastPieceLength int      // Length of last piece
	NumberOfPieces  int      // Number of pieces
	PieceHashes     [][]byte // SHA1 hashes of each piece
	Files           []File   // List of files for multitorrent
	mutex           sync.RWMutex
}

type File struct {
	Length int
	Path   []string
}

func CreateInfoDictionary(infoDict map[string]interface{}) *InfoDictionary {
	infoDictionaryStruct := &InfoDictionary{}

	infoDictionaryStruct.Name = infoDict["name"].(string)
	infoDictionaryStruct.PieceLength = infoDict["piece length"].(int)
	infoDictionaryStruct.PieceHashes = hash.CastHashTo2dByteSlice(infoDict["pieces"])
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
		var fileStructs []File

		for _, file := range files {
			fileMap := file.(map[string]interface{})
			length := fileMap["length"].(int)
			pathInterface := fileMap["path"].([]interface{})
			path := make([]string, len(pathInterface))

			for i, p := range pathInterface {
				path[i] = p.(string)
			}

			fileStruct := File{
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

func (f *InfoDictionary) GetPieceLength(index int) int {
	if index == f.NumberOfPieces-1 {
		return f.LastPieceLength
	} else {
		return f.PieceLength
	}
}

func (f *InfoDictionary) AddPiece(piece []byte, pieceIndex int) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	offset := f.PieceLength * pieceIndex
	copy(f.Data[offset:offset+len(piece)], piece)
}

func (f *InfoDictionary) Print() {
	fmt.Printf("\tFile Details:\n")
	fmt.Printf("\tName: %s\n", f.Name)
	fmt.Printf("\tType: %s\n", f.Type)
	fmt.Printf("\tFile Length: %d bytes\n", f.FileLength)
	fmt.Printf("\tPiece Length: %d bytes\n", f.PieceLength)
	fmt.Printf("\tLast Piece Length: %d bytes\n", f.LastPieceLength)
	fmt.Printf("\tNumber of Pieces: %d\n", f.NumberOfPieces)
	fmt.Printf("\tPiece Hashes:\n")
	for i, hash := range f.PieceHashes {
		fmt.Printf("\t\tPiece %d: %x\n", i, hash)
	}

	if f.Type == MULTI {
		fmt.Printf("\tFiles:\n")
		for _, file := range f.Files {
			fmt.Printf("\t\tPath: %s, Length: %d bytes\n", fmt.Sprintf("%v", file.Path), file.Length)
		}
	}
}

func (f *InfoDictionary) Log() {
	log.Printf("\tFile Details:\n")
	log.Printf("\tName: %s\n", f.Name)
	log.Printf("\tType: %s\n", f.Type)
	log.Printf("\tFile Length: %d bytes\n", f.FileLength)
	log.Printf("\tPiece Length: %d bytes\n", f.PieceLength)
	log.Printf("\tLast Piece Length: %d bytes\n", f.LastPieceLength)
	log.Printf("\tNumber of Pieces: %d\n", f.NumberOfPieces)
	log.Printf("\tPiece Hashes:\n")
	for i, hash := range f.PieceHashes {
		log.Printf("\t\tPiece %d: %x\n", i, hash)
	}

	if f.Type == MULTI {
		log.Printf("\tFiles:\n")
		for _, file := range f.Files {
			log.Printf("\t\tPath: %s, Length: %d bytes\n", fmt.Sprintf("%v", file.Path), file.Length)
		}
	}
}
