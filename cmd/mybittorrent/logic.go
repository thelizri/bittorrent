package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"karlan/torrent/internal/bencode"
	"karlan/torrent/internal/client"
	"karlan/torrent/internal/download"
	"karlan/torrent/internal/fileio"
	"karlan/torrent/internal/queue"
	"karlan/torrent/internal/torrent"
	"karlan/torrent/internal/tracker"
	"karlan/torrent/internal/utils"
	"os"
	"sync"
)

func decodeBencodedString(encoding string) {
	decoded, _, err := bencode.Decode(encoding, 0)
	if err != nil {
		utils.LogAndPrintf("Error decoding bencode: %v\n", err)
		utils.LogPrintExit(fmt.Sprintf("Error decoding bencode: %v\n", err))
	}
	jsonOutput, _ := json.Marshal(decoded)
	utils.LogAndPrintf("%s\n", string(jsonOutput))
}

func printTorrentInfo(filePath string) {
	utils.LogAndPrintf("Opening torrent file: %s\n", filePath)
	torrent := torrent.Open(filePath)
	utils.LogAndPrintln("Printing torrent info")
	torrent.Print()
}

func printPeers(filePath string) {
	utils.LogAndPrintf("Opening torrent file: %s\n", filePath)
	torrent := torrent.Open(filePath)
	utils.LogAndPrintln("Fetching peers from tracker")
	interval, clients := tracker.GET(torrent)
	utils.LogAndPrintf("Interval: %v\n", interval)
	for _, client := range clients {
		utils.LogAndPrintf("Peer address: %s\n", client.Address())
	}
}

func performHandshakeWithPeer(address, filePath string) {
	utils.LogAndPrintf("Opening torrent file: %s\n", filePath)
	torrent := torrent.Open(filePath)
	utils.LogAndPrintf("Converting address to client: %s\n", address)
	client, err := client.StringToClient(address)
	if err != nil {
		utils.LogAndPrintf("Error parsing address: %v\n", err)
		utils.LogPrintExit(fmt.Sprintf("Error parsing address: %v\n", err))
	}
	utils.LogAndPrintln("Initiating handshake with peer")
	err = client.Init(torrent.InfoHash, torrent.PeerID)
	if err != nil {
		utils.LogAndPrintf("Error connecting with client: %v\n", err)
		utils.LogAndPrintln(fmt.Sprintf("Error connecting with client: %v\n", err))
	}
	utils.LogAndPrintf("Handshake successful. Peer ID: %s\n", hex.EncodeToString(client.PeerID[:]))
}

func downloadPiece(torrentPath, outputPath string, pieceIndex int) {
	utils.LogAndPrintf("Opening torrent file: %s\n", torrentPath)
	torrent := torrent.Open(torrentPath)
	utils.LogAndPrintln("Printing torrent info")
	torrent.Print()
	utils.LogAndPrintln("Fetching peers from tracker")
	interval, clients := tracker.GET(torrent)
	utils.LogAndPrintf("Interval: %v\n", interval)

	for i, c := range clients {
		utils.LogAndPrintf("Index of clients: %d\n", i)
		utils.LogAndPrintf("Initiating connection with client %s\n", c.Address())
		err := (&c).Init(torrent.InfoHash, torrent.PeerID)
		if err != nil {
			utils.LogAndPrintf("Error initializing client: %s, Error: %v\n", c.Address(), err)
			continue
		}

		utils.LogAndPrintf("Downloading piece %d from client %s\n", pieceIndex, c.Address())
		pieceProgress, err := download.DownloadPiece(&c, pieceIndex, torrent.GetPieceLength(pieceIndex), torrent.GetPieceHash(pieceIndex))
		if err != nil {
			utils.LogAndPrintf("Error downloading piece: %v\n", err)
			continue
		}

		c.Conn.Close()
		utils.LogAndPrintf("Writing downloaded piece to file: %s\n", outputPath)
		fileio.WriteToAbsolutePath(outputPath, pieceProgress.Data)
		break
	}
}

func downloadFile(torrentPath, outputPath string) {
	utils.LogAndPrintf("Opening torrent file: %s\n", torrentPath)
	t := torrent.Open(torrentPath)
	utils.LogAndPrintln("Printing torrent info")
	t.Print()
	utils.LogAndPrintln("Fetching peers from tracker")
	interval, clients := tracker.GET(t)
	utils.LogAndPrintf("Interval: %v\n", interval)

	q := queue.NewQueue(t.GetNumberOfPieces())
	var wg sync.WaitGroup
	for i, c := range clients {
		utils.LogAndPrintf("Index of clients: %d\n", i)
		utils.LogAndPrintf("Initiating connection with client %s\n", c.Address())
		err := (&c).Init(t.InfoHash, t.PeerID)
		if err != nil {
			utils.LogAndPrintf("Error initializing client: %s, Error: %v\n", c.Address(), err)
			continue
		}

		utils.LogAndPrintf("Downloading File %v from client %s\n", t.GetName(), c.Address())
		wg.Add(1)
		go download.DownloadFile(&c, t, q, &wg)
	}

	wg.Wait()
	t.Print()

	// Verify the integrity of each piece and check that everything is downloaded
	if !t.FinishedDownloading() {
		utils.LogSeparator()
		utils.LogAndPrintln("Torrent finished downloading with missing pieces")
		os.Exit(1)
	}

	fileio.WriteToAbsolutePath(outputPath, t.GetData())
	utils.LogAndPrintf("Writing torrent to file %s", outputPath)
}
