package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"karlan/torrent/internal/bencode"
	"karlan/torrent/internal/fileio"
	"karlan/torrent/internal/hash"
	"karlan/torrent/internal/network"
	"karlan/torrent/internal/types"
	"karlan/torrent/internal/utils"
)

const (
	BITFIELD   byte = 5 // Contains a bitfield representing the pieces the sender has
	INTERESTED byte = 2 // Indicates the sender wants to download pieces from the recipient
	UNCHOKE    byte = 1 // Indicates the sender will now allow the receiver to request pieces
	REQUEST    byte = 6 // Requests a specific piece of data
	PIECE      byte = 7 // Contains the actual data of the piece being sent
)

func main() {
	setupLogging()

	if len(os.Args) < 2 {
		fmt.Println("Command is required")
		os.Exit(1)
	}
	command := os.Args[1]

	commands := map[string]func(){
		"decode":         decodeCommand,
		"info":           infoCommand,
		"peers":          peersCommand,
		"download":       downloadCommand,
		"download_piece": downloadPieceCommand,
		"handshake":      handshakeCommand,
	}

	if cmdFunc, exists := commands[command]; exists {
		cmdFunc()
	} else {
		fmt.Println("Unknown command:", command)
		os.Exit(1)
	}
}

func setupLogging() {
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %s", err)
	}
	defer file.Close()
	log.SetOutput(file)
	log.Printf("Program arguments: %v\n", os.Args[1:])
}

func decodeCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: ./your_bittorrent.sh decode <bencoded_string>")
		os.Exit(1)
	}
	decodeBencodedString(os.Args[2])
}

func infoCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: ./your_bittorrent.sh info <file_path>")
		os.Exit(1)
	}
	printTorrentInfo(os.Args[2])
}

func peersCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: ./your_bittorrent.sh peers <file_path>")
		os.Exit(1)
	}
	printPeers(os.Args[2])
}

func downloadCommand() {
	if len(os.Args) < 5 || os.Args[2] != "-o" {
		fmt.Println("Usage: ./your_bittorrent.sh download -o <output_path> <torrent_path>")
		os.Exit(1)
	}
	downloadFile(os.Args[4], os.Args[3])
}

func downloadPieceCommand() {
	if len(os.Args) < 6 || os.Args[2] != "-o" {
		fmt.Println("Usage: ./your_bittorrent.sh download_piece -o <output_path> <torrent_path> <piece_index>")
		os.Exit(1)
	}
	pieceIndex, err := strconv.Atoi(os.Args[5])
	if err != nil {
		fmt.Println("Invalid piece index")
		fmt.Println("Usage: ./your_bittorrent.sh download_piece -o <output_path> <torrent_path> <piece_index>")
		os.Exit(1)
	}
	downloadPiece(os.Args[4], os.Args[3], pieceIndex)
}

func handshakeCommand() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: ./your_bittorrent.sh handshake <file_path> <ip>:<port>")
		os.Exit(1)
	}
	performHandshakeWithPeer(os.Args[3], os.Args[2])
}

func performHandshakeWithPeer(address, filePath string) {
	torrent := fileio.ReadTorrentFile(filePath)
	peer, _ := utils.StringToPeerAddress(address)
	response, conn, err := network.PerformHandshake(&torrent, &peer)
	if err != nil {
		log.Fatalf("Error handshake: %v\n", err)
	}
	defer conn.Close()
	fmt.Printf("Peer ID: %s\n", hex.EncodeToString(response[len(response)-20:]))
}

func decodeBencodedString(encoding string) {
	decoded, _, err := bencode.Decode(encoding, 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	jsonOutput, _ := json.Marshal(decoded)
	fmt.Println(string(jsonOutput))
}

func printTorrentInfo(filePath string) {
	torrent := fileio.ReadTorrentFile(filePath)
	torrent.Print()
}

func printPeers(filePath string) {
	torrent := fileio.ReadTorrentFile(filePath)
	body := network.PerformTrackerRequest(&torrent)
	peers := utils.ExtractPeersFromRespone(body)
	for _, peer := range peers {
		peer.Print()
	}
}

func connectToPeer(torrent *types.Torrent, peers []types.PeerAddress) net.Conn {
	for _, peer := range peers {
		_, conn, err := network.PerformHandshake(torrent, &peer)
		if err != nil {
			log.Printf("Error connecting to peer: %v\n", err)
			continue
		} else {
			fmt.Printf("Established connection with: %v\n", peer.GetAddress())
			return conn
		}
	}
	log.Println("Failed to establish connection with any peer.")
	return nil
}

func downloadPiece(torrentPath, outputPath string, pieceIndex int) {
	torrent := fileio.ReadTorrentFile(torrentPath)
	body := network.PerformTrackerRequest(&torrent)
	peers := utils.ExtractPeersFromRespone(body)
	torrent.Log()

	conn := connectToPeer(&torrent, peers)
	if conn == nil {
		return
	}
	defer conn.Close()

	message, err := network.ListenForPeerMessage(conn, BITFIELD)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Bitfield message: %v\n", message)

	message = types.PeerMessage{MessageID: INTERESTED}
	network.SendMessageToPeer(conn, &message)

	message, err = network.ListenForPeerMessage(conn, UNCHOKE)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Unchoke message: %v\n", message)

	piece := network.FetchPiece(conn, torrent.GetPieceLength(pieceIndex), pieceIndex)
	if err = hash.ValidatePieceHash(piece, torrent.PieceHashes[pieceIndex]); err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fileio.WriteToAbsolutePath(outputPath, piece)
	log.Printf("Piece %v downloaded to %v.\n", pieceIndex, outputPath)
	fmt.Printf("Piece %v downloaded to %v.\n", pieceIndex, outputPath)
}

func downloadFile(torrentPath, outputPath string) {
	torrent := fileio.ReadTorrentFile(torrentPath)
	body := network.PerformTrackerRequest(&torrent)
	peers := utils.ExtractPeersFromRespone(body)
	torrent.Log()

	conn := connectToPeer(&torrent, peers)
	if conn == nil {
		return
	}
	defer conn.Close()

	message, err := network.ListenForPeerMessage(conn, BITFIELD)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Bitfield message: %v\n", message)

	message = types.PeerMessage{MessageID: INTERESTED}
	network.SendMessageToPeer(conn, &message)

	message, err = network.ListenForPeerMessage(conn, UNCHOKE)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Unchoke message: %v\n", message)

	fileSlice := make([]byte, torrent.FileLength)
	offset := 0
	for pieceIndex := 0; pieceIndex < torrent.NumberOfPieces; pieceIndex++ {
		piece := network.FetchPiece(conn, torrent.GetPieceLength(pieceIndex), pieceIndex)
		if err = hash.ValidatePieceHash(piece, torrent.PieceHashes[pieceIndex]); err != nil {
			log.Printf("Error: %v\n", err)
			return
		}
		copy(fileSlice[offset:offset+len(piece)], piece)
		offset += len(piece)
	}

	fileio.WriteToAbsolutePath(outputPath, fileSlice)
	log.Printf("Downloaded %v to %v.\n", torrentPath, outputPath)
	fmt.Printf("Downloaded %v to %v.\n", torrentPath, outputPath)
}
