package main

import (
	"fmt"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

var logFile *os.File

func init() {
	// Open a file for logging
	var err error
	logFile, err = os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// Output to the log file instead of the default stderr
	log.SetOutput(logFile)

	// Only log the warning severity or above.
	log.SetLevel(log.TraceLevel)
}

func main() {
	defer logFile.Close()

	if len(os.Args) < 2 {
		fmt.Println("Command is required")
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
		log.Infof("Unknown command: %v\n", command)
		os.Exit(1)
	}
}

func decodeCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: ./bittorrent.sh decode <bencoded_string>")
		os.Exit(1)
	}
	decodeBencodedString(os.Args[2])
}

func infoCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: ./bittorrent.sh info <file_path>")
		os.Exit(1)
	}
	printTorrentInfo(os.Args[2])
}

func peersCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: ./bittorrent.sh peers <file_path>")
		os.Exit(1)
	}
	printPeers(os.Args[2])
}

func handshakeCommand() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: ./bittorrent.sh handshake <file_path> <ip>:<port>")
		os.Exit(1)
	}
	performHandshakeWithPeer(os.Args[3], os.Args[2])
}

func downloadPieceCommand() {
	if len(os.Args) < 6 || os.Args[2] != "-o" {
		fmt.Println("Usage: ./bittorrent.sh download_piece -o <output_path> <torrent_path> <piece_index>")
		os.Exit(1)
	}
	pieceIndex, err := strconv.Atoi(os.Args[5])
	if err != nil {
		fmt.Println("Invalid piece index")
		fmt.Println("Usage: ./bittorrent.sh download_piece -o <output_path> <torrent_path> <piece_index>")
		os.Exit(1)
	}
	downloadPiece(os.Args[4], os.Args[3], pieceIndex)
}

func downloadFileCommand() {
	if len(os.Args) < 5 || os.Args[2] != "-o" {
		fmt.Println("Usage: ./bittorrent.sh download -o <output_path> <torrent_path>")
		os.Exit(1)
	}
	downloadFile(os.Args[4], os.Args[3])
}
