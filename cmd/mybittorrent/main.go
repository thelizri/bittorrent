package main

import (
	"karlan/torrent/internal/utils"
	"log"
	"os"
	"strconv"
)

func main() {
	file := setupLogging()
	defer file.Close()

	if len(os.Args) < 2 {
		utils.LogAndPrintln("Command is required")
		os.Exit(1)
	}
	command := os.Args[1]

	commands := map[string]func(){
		"decode":         decodeCommand,
		"info":           infoCommand,
		"peers":          peersCommand,
		"handshake":      handshakeCommand,
		"download_piece": downloadPieceCommand,
		"download":       downloadFileCommand,
	}

	if cmdFunc, exists := commands[command]; exists {
		cmdFunc()
	} else {
		utils.LogAndPrintf("Unknown command: %v\n", command)
		os.Exit(1)
	}
}

func setupLogging() *os.File {
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %s", err)
	}
	log.SetOutput(file)
	log.Printf("Program arguments: %v\n", os.Args[1:])
	return file
}

func decodeCommand() {
	if len(os.Args) < 3 {
		utils.LogAndPrintln("Usage: ./your_bittorrent.sh decode <bencoded_string>")
		os.Exit(1)
	}
	decodeBencodedString(os.Args[2])
}

func infoCommand() {
	if len(os.Args) < 3 {
		utils.LogAndPrintln("Usage: ./your_bittorrent.sh info <file_path>")
		os.Exit(1)
	}
	printTorrentInfo(os.Args[2])
}

func peersCommand() {
	if len(os.Args) < 3 {
		utils.LogAndPrintln("Usage: ./your_bittorrent.sh peers <file_path>")
		os.Exit(1)
	}
	printPeers(os.Args[2])
}

func handshakeCommand() {
	if len(os.Args) < 4 {
		utils.LogAndPrintln("Usage: ./your_bittorrent.sh handshake <file_path> <ip>:<port>")
		os.Exit(1)
	}
	performHandshakeWithPeer(os.Args[3], os.Args[2])
}

func downloadPieceCommand() {
	if len(os.Args) < 6 || os.Args[2] != "-o" {
		utils.LogAndPrintln("Usage: ./your_bittorrent.sh download_piece -o <output_path> <torrent_path> <piece_index>")
		os.Exit(1)
	}
	pieceIndex, err := strconv.Atoi(os.Args[5])
	if err != nil {
		utils.LogAndPrintln("Invalid piece index")
		utils.LogAndPrintln("Usage: ./your_bittorrent.sh download_piece -o <output_path> <torrent_path> <piece_index>")
		os.Exit(1)
	}
	downloadPiece(os.Args[4], os.Args[3], pieceIndex)
}

func downloadFileCommand() {
	if len(os.Args) < 5 || os.Args[2] != "-o" {
		utils.LogAndPrintln("Usage: ./your_bittorrent.sh download -o <output_path> <torrent_path>")
		os.Exit(1)
	}
	downloadFile(os.Args[4], os.Args[3])
}
