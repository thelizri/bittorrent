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
		log.Fatalf("Error decoding bencode: %v", err)
	}
	jsonOutput, _ := json.Marshal(decoded)
	log.Debugf("Decoded JSON: %s", string(jsonOutput))
	fmt.Println("Decoded bencode string to JSON format.")
}

func printTorrentInfo(filePath string) {
	log.Infof("Opening torrent file: %s", filePath)
	torrent := torrent.Open(filePath)
	log.Infof("Printing torrent info")
	torrent.Print()
	fmt.Printf("Torrent info for %s has been printed.\n", filePath)
}

func printPeers(filePath string) {
	log.Infof("Opening torrent file: %s", filePath)
	torrent := torrent.Open(filePath)
	log.Infof("Fetching peers from tracker")
	interval, clients := tracker.GET(torrent)
	log.Debugf("Tracker interval: %v", interval)
	fmt.Printf("Found %d peers.\n", len(clients))
	for _, client := range clients {
		fmt.Printf("Peer address: %s\n", client.Address())
	}
}

func performHandshakeWithPeer(address, filePath string) {
	log.Infof("Opening torrent file: %s", filePath)
	torrent := torrent.Open(filePath)
	log.Debugf("Converting address to client: %s", address)
	client, err := client.StringToClient(address)
	if err != nil {
		log.Errorf("Error parsing address: %v", err)
		log.Fatal(fmt.Sprintf("Error parsing address: %v", err))
	}
	log.Infof("Initiating handshake with peer")
	err = client.Init(torrent.InfoHash, torrent.PeerID)
	if err != nil {
		log.Errorf("Error connecting with client: %v", err)
		log.Fatal(fmt.Sprintf("Error connecting with client: %v", err))
	}
	fmt.Printf("Handshake successful with peer at address %s. Peer ID: %s\n", address, hex.EncodeToString(client.PeerID[:]))
}

func downloadPiece(torrentPath, outputPath string, pieceIndex int) {
	log.Infof("Opening torrent file: %s", torrentPath)
	torrent := torrent.Open(torrentPath)
	log.Debugf("Printing torrent info")
	torrent.Log()
	log.Infof("Fetching peers from tracker")
	interval, clients := tracker.GET(torrent)
	log.Debugf("Tracker interval: %v", interval)

	for i, c := range clients {
		log.Debugf("Index of clients: %d", i)
		log.Infof("Initiating connection with client %s", c.Address())
		err := (&c).Init(torrent.InfoHash, torrent.PeerID)
		if err != nil {
			log.Warnf("Error initializing client: %s, Error: %v", c.Address(), err)
			continue
		}

		log.Infof("Downloading piece %d from client %s", pieceIndex, c.Address())
		pieceProgress, err := download.DownloadPiece(&c, pieceIndex, torrent.GetPieceLength(pieceIndex), torrent.GetPieceHash(pieceIndex))
		if err != nil {
			log.Warnf("Error downloading piece: %v", err)
			continue
		}

		c.Conn.Close()
		log.Infof("Writing downloaded piece to file: %s", outputPath)
		fileio.WriteToAbsolutePath(outputPath, pieceProgress.Data)
		fmt.Printf("Downloaded piece %d to %s\n", pieceIndex, outputPath)
		break
	}
}

func downloadFile(torrentPath, outputPath string) {
	log.Infof("Opening torrent file: %s", torrentPath)
	t := torrent.Open(torrentPath)
	log.Debugf("Printing torrent info")
	t.Log()
	log.Infof("Fetching peers from tracker")
	interval, clients := tracker.GET(t)
	log.Debugf("Tracker interval: %v", interval)

	q := queue.NewQueue(t.GetNumberOfPieces())
	var wg sync.WaitGroup
	for i, c := range clients {
		log.Debugf("Index of clients: %d", i)
		log.Infof("Initiating connection with client %s", c.Address())
		err := (&c).Init(t.InfoHash, t.PeerID)
		if err != nil {
			log.Warnf("Error initializing client: %s, Error: %v", c.Address(), err)
			continue
		}

		log.Infof("Downloading file %v from client %s", t.GetName(), c.Address())
		wg.Add(1)
		go download.DownloadFile(&c, t, q, &wg)
	}

	wg.Wait()
	t.Log()

	// Verify the integrity of each piece and check that everything is downloaded
	if !t.FinishedDownloading() {
		log.Errorf("Torrent finished downloading with missing pieces")
		os.Exit(1)
	}

	fileio.WriteToAbsolutePath(outputPath, t.GetData())
	log.Infof("Writing torrent to file %s", outputPath)
	fmt.Printf("Downloaded and wrote torrent to %s\n", outputPath)
}
