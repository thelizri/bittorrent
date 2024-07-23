package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"

	"karlan/torrent/internal/bencode"
	"karlan/torrent/internal/fileio"
	"karlan/torrent/internal/hash"
	"karlan/torrent/internal/network"
	"karlan/torrent/internal/types"
	"karlan/torrent/internal/utils"
)

func main() {
	file := setupLogging()
	defer file.Close()

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
	response, conn, err := network.PerformHandshake(torrent, &peer)
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
	body := network.PerformTrackerRequest(torrent)
	peers := utils.ExtractPeersFromResponse(body)
	for _, peer := range peers {
		peer.Print()
	}
}

func downloadPiece(torrentPath, outputPath string, pieceIndex int) {

	torrent := fileio.ReadTorrentFile(torrentPath)
	body := network.PerformTrackerRequest(torrent)
	peers := utils.ExtractPeersFromResponse(body)
	torrent.Log()
	utils.LogSeparator()

	conn := network.ConnectToPeer(torrent, peers)
	if conn == nil {
		return
	}
	defer conn.Close()

	message, err := network.ListenForPeerMessage(conn, types.MSG_BITFIELD)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Bitfield message: %v\n", message)

	message = types.Message{MessageID: types.MSG_INTERESTED}
	network.SendMessageToPeer(conn, &message)

	message, err = network.ListenForPeerMessage(conn, types.MSG_UNCHOKE)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Unchoke message: %v\n", message)

	piece := network.FetchPiece(conn, torrent.File.GetPieceLength(pieceIndex), pieceIndex)
	if err = hash.ValidatePieceHash(piece, torrent.File.PieceHashes[pieceIndex]); err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	fileio.WriteToAbsolutePath(outputPath, piece)
	utils.LogSeparator()
	utils.LogAndPrint(fmt.Sprintf("Piece %v downloaded to %v.\n", pieceIndex, outputPath))

}

func downloadFile(torrentPath, outputPath string) {
	torrent := fileio.ReadTorrentFile(torrentPath)
	body := network.PerformTrackerRequest(torrent)
	peers := utils.ExtractPeersFromResponse(body)
	torrent.Log()
	utils.LogSeparator()

	file := &torrent.File
	queue := types.NewQueue()
	for i := 0; i < file.NumberOfPieces; i++ {
		queue.Enqueue(i)
	}

	var wg sync.WaitGroup
	connections := network.ConnectToPeers(torrent, peers)
	for _, conn := range connections {
		wg.Add(1)
		go worker(conn, queue, file, &wg)
	}
	wg.Wait()

	fileio.WriteToAbsolutePath(outputPath, file.Data)
	utils.LogSeparator()
	utils.LogAndPrint(fmt.Sprintf("Downloaded %v to %v.\n", torrentPath, outputPath))
}

func worker(conn net.Conn, queue *types.Queue, file *types.InfoDictionary, wg *sync.WaitGroup) {
	defer wg.Done()
	defer conn.Close()

	message, err := network.ListenForPeerMessage(conn, types.MSG_BITFIELD)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Bitfield message: %v\n", message)

	message = types.Message{MessageID: types.MSG_INTERESTED}
	network.SendMessageToPeer(conn, &message)

	message, err = network.ListenForPeerMessage(conn, types.MSG_UNCHOKE)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Unchoke message: %v\n", message)

	network.FetchFile(conn, queue, file)
}
