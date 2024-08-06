package torrent

import (
	"fmt"
	"karlan/torrent/internal/hash"
	"karlan/torrent/internal/utils"
	"log"
)

const MULTI = "Multi-File Torrent"
const SINGLE = "Single-File Torrent"

type TorrentDictionary struct {
	Data            []byte
	Type            string
	Name            string     // Name of the file (for single-file) or root directory (for multi-file)
	FileLength      int        // Length of the file, single file torrent
	PieceLength     int        // Length of each piece
	LastPieceLength int        // Length of last piece
	NumberOfPieces  int        // Number of pieces
	PieceHashes     [][20]byte // SHA1 hashes of each piece
	Files           []File     // List of files for multitorrent
}

type File struct {
	Length int
	Path   []string
}

func CreateInfoDictionary(infoDict map[string]interface{}) *TorrentDictionary {
	infoDictionaryStruct := &TorrentDictionary{}

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

func (f *TorrentDictionary) GetPieceLength(index int) int {
	if index == f.NumberOfPieces-1 {
		return f.LastPieceLength
	} else {
		return f.PieceLength
	}
}

func (f *TorrentDictionary) AddPiece(piece []byte, pieceIndex int) {
	offset := f.PieceLength * pieceIndex
	copy(f.Data[offset:offset+len(piece)], piece)
}

func (f *TorrentDictionary) Print() {
	utils.LogAndPrintf("\tFile Details:\n")
	utils.LogAndPrintf("\tName: %s\n", f.Name)
	utils.LogAndPrintf("\tType: %s\n", f.Type)
	utils.LogAndPrintf("\tFile Length: %d bytes\n", f.FileLength)
	utils.LogAndPrintf("\tPiece Length: %d bytes\n", f.PieceLength)
	utils.LogAndPrintf("\tLast Piece Length: %d bytes\n", f.LastPieceLength)
	utils.LogAndPrintf("\tNumber of Pieces: %d\n", f.NumberOfPieces)
	if f.NumberOfPieces < 10 {
		utils.LogAndPrintf("\tPiece Hashes:\n")
		for i, hash := range f.PieceHashes {
			utils.LogAndPrintf("\t\tPiece %d: %x\n", i, hash)
		}
	}

	if f.Type == MULTI {
		utils.LogAndPrintf("\tFiles:\n")
		for _, file := range f.Files {
			utils.LogAndPrintf("\t\tPath: %s, Length: %d bytes\n", fmt.Sprintf("%v", file.Path), file.Length)
		}
	}
}

func (f *TorrentDictionary) Log() {
	log.Printf("\tFile Details:\n")
	log.Printf("\tName: %s\n", f.Name)
	log.Printf("\tType: %s\n", f.Type)
	log.Printf("\tFile Length: %d bytes\n", f.FileLength)
	log.Printf("\tPiece Length: %d bytes\n", f.PieceLength)
	log.Printf("\tLast Piece Length: %d bytes\n", f.LastPieceLength)
	log.Printf("\tNumber of Pieces: %d\n", f.NumberOfPieces)
	if f.NumberOfPieces < 10 {
		log.Printf("\tPiece Hashes:\n")
		for i, hash := range f.PieceHashes {
			log.Printf("\t\tPiece %d: %x\n", i, hash)
		}
	}

	if f.Type == MULTI {
		log.Printf("\tFiles:\n")
		for _, file := range f.Files {
			log.Printf("\t\tPath: %s, Length: %d bytes\n", fmt.Sprintf("%v", file.Path), file.Length)
		}
	}
}

func (f *TorrentDictionary) VerifyIntegrityOfEachPiece() bool {
	for i := 0; i < f.NumberOfPieces-1; i++ {
		start := f.PieceLength * i
		end := start + f.PieceLength
		piece := f.Data[start:end]
		err := hash.ValidatePieceHash(piece, f.PieceHashes[i])
		if err != nil {
			utils.LogAndPrintln(err.Error())
			return false
		}
	}

	start := f.PieceLength * (f.NumberOfPieces - 1)
	piece := f.Data[start:]
	err := hash.ValidatePieceHash(piece, f.PieceHashes[(f.NumberOfPieces-1)])
	if err != nil {
		utils.LogAndPrintln(err.Error())
		return false
	}

	return true
}
