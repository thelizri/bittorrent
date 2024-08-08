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

	"os"
	"sync"

	log "github.com/sirupsen/logrus"
)

func decodeBencodedString(encoding string) {
	decoded, _, err := bencode.Decode(encoding, 0)
	if err != nil {
		log.Info("Error decoding bencode: %v\n", err)
		log.Fatal(fmt.Sprintf("Error decoding bencode: %v\n", err))
	}
	jsonOutput, _ := json.Marshal(decoded)
	log.Info("%s\n", string(jsonOutput))
}

func printTorrentInfo(filePath string) {
	log.Info("Opening torrent file: %s\n", filePath)
	torrent := torrent.Open(filePath)
	log.Info("Printing torrent info")
	torrent.Log()
}

func printPeers(filePath string) {
	log.Info("Opening torrent file: %s\n", filePath)
	torrent := torrent.Open(filePath)
	log.Info("Fetching peers from tracker")
	interval, clients := tracker.GET(torrent)
	log.Info("Interval: %v\n", interval)
	for _, client := range clients {
		log.Info("Peer address: %s\n", client.Address())
	}
}

func performHandshakeWithPeer(address, filePath string) {
	log.Info("Opening torrent file: %s\n", filePath)
	torrent := torrent.Open(filePath)
	log.Info("Converting address to client: %s\n", address)
	client, err := client.StringToClient(address)
	if err != nil {
		log.Info("Error parsing address: %v\n", err)
		log.Fatal(fmt.Sprintf("Error parsing address: %v\n", err))
	}
	log.Info("Initiating handshake with peer")
	err = client.Init(torrent.InfoHash, torrent.PeerID)
	if err != nil {
		log.Info("Error connecting with client: %v\n", err)
		log.Info(fmt.Sprintf("Error connecting with client: %v\n", err))
	}
	log.Info("Handshake successful. Peer ID: %s\n", hex.EncodeToString(client.PeerID[:]))
}

func downloadPiece(torrentPath, outputPath string, pieceIndex int) {
	log.Info("Opening torrent file: %s\n", torrentPath)
	torrent := torrent.Open(torrentPath)
	log.Info("Printing torrent info")
	torrent.Log()
	log.Info("Fetching peers from tracker")
	interval, clients := tracker.GET(torrent)
	log.Info("Interval: %v\n", interval)

	for i, c := range clients {
		log.Info("Index of clients: %d\n", i)
		log.Info("Initiating connection with client %s\n", c.Address())
		err := (&c).Init(torrent.InfoHash, torrent.PeerID)
		if err != nil {
			log.Info("Error initializing client: %s, Error: %v\n", c.Address(), err)
			continue
		}

		log.Info("Downloading piece %d from client %s\n", pieceIndex, c.Address())
		pieceProgress, err := download.DownloadPiece(&c, pieceIndex, torrent.GetPieceLength(pieceIndex), torrent.GetPieceHash(pieceIndex))
		if err != nil {
			log.Info("Error downloading piece: %v\n", err)
			continue
		}

		c.Conn.Close()
		log.Info("Writing downloaded piece to file: %s\n", outputPath)
		fileio.WriteToAbsolutePath(outputPath, pieceProgress.Data)
		break
	}
}

func downloadFile(torrentPath, outputPath string) {
	log.Info("Opening torrent file: %s\n", torrentPath)
	t := torrent.Open(torrentPath)
	log.Info("Printing torrent info")
	t.Log()
	log.Info("Fetching peers from tracker")
	interval, clients := tracker.GET(t)
	log.Info("Interval: %v\n", interval)

	q := queue.NewQueue(t.GetNumberOfPieces())
	var wg sync.WaitGroup
	for i, c := range clients {
		log.Info("Index of clients: %d\n", i)
		log.Info("Initiating connection with client %s\n", c.Address())
		err := (&c).Init(t.InfoHash, t.PeerID)
		if err != nil {
			log.Info("Error initializing client: %s, Error: %v\n", c.Address(), err)
			continue
		}

		log.Info("Downloading File %v from client %s\n", t.GetName(), c.Address())
		wg.Add(1)
		go download.DownloadFile(&c, t, q, &wg)
	}

	wg.Wait()
	t.Log()

	// Verify the integrity of each piece and check that everything is downloaded
	if !t.FinishedDownloading() {

		log.Info("Torrent finished downloading with missing pieces")
		os.Exit(1)
	}

	fileio.WriteToAbsolutePath(outputPath, t.GetData())
	log.Info("Writing torrent to file %s", outputPath)
}
