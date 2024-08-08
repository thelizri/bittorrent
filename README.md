# BitTorrent Client

This BitTorrent client is a simple tool designed to download files using `.torrent` files. It has limited functionality, with specific constraints and features outlined below.

## Features

- **Supports `.torrent` files**: This client works exclusively with `.torrent` files and does not support magnet links.
- **Supports HTTP trackers**: Only HTTP trackers are supported. There is no support for UDP trackers.
- **Single-file torrents only**: Multi-file torrents are not supported.
- **Leech-only mode**: This client downloads files but does not upload pieces back to the network.
- **No DHT support**: The client does not support the Distributed Hash Table (DHT) protocol.

## Usage

Below are the example commands and their descriptions:

### Decode

Decode bencoded data using the `decode` command:

```sh
./bittorrent.sh decode <bencoded_data>
```

**Examples:**

```sh
./bittorrent.sh decode 5:hello
```
Output:
```
"hello"
```

```sh
./bittorrent.sh decode i52e
```
Output:
```
52
```

```sh
./bittorrent.sh decode l5:helloi52ee
```
Output:
```
["hello", 52]
```

```sh
./bittorrent.sh decode d3:foo3:bar5:helloi52ee
```
Output:
```
{"foo": "bar", "hello": 52}
```

### Torrent Info

Get information about a torrent file using the `info` command:

```sh
./bittorrent.sh info <file.torrent>
```

**Example:**

```sh
./bittorrent.sh info torrents/sample.torrent
```
Output:
```
Tracker URL: http://bittorrent-test-tracker.codecrafters.io/announce
Length: 92063
Info Hash: d69f91e6b2ae4c542468d1073a71d4ea13879a7f
Piece Length: 32768
Piece Hashes:
e876f67a2a8886e8f36b136726c30fa29703022d
6e2275e604a0766656736e81ff10b55204ad8d35
f00d937a0213df1982bc8d097227ad9e909acc17
```

### List Peers

List peers available for a torrent using the `peers` command:

```sh
./bittorrent.sh peers <file.torrent>
```

**Example:**

```sh
./bittorrent.sh peers torrents/sample.torrent
```
Output:
```
178.62.82.89:51470
165.232.33.77:51467
178.62.85.20:51489
```

### Handshake

Initiate a handshake with a peer using the `handshake` command:

```sh
./bittorrent.sh handshake <file.torrent> <peer_ip>:<peer_port>
```

**Example:**

```sh
./bittorrent.sh handshake torrents/sample.torrent 178.62.82.89:51470
```
Output:
```
Peer ID: 0102030405060708090a0b0c0d0e0f1011121314
```

### Download Piece

Download a specific piece of the file using the `download_piece` command:

```sh
./bittorrent.sh download_piece -o <output_path> <file.torrent> <piece_index>
```

**Example:**

```sh
./bittorrent.sh download_piece -o /Users/william/Documents/bittorrent/tmp/test-piece-0 torrents/sample.torrent 0
```
Output:
```
Piece 0 downloaded to /tmp/test-piece-0.
```

### Download File

Download the entire file using the `download` command:

```sh
./bittorrent.sh download -o <output_path> <file.torrent>
```

**Examples:**

```sh
./bittorrent.sh download -o /Users/william/Documents/bittorrent/tmp/sample.txt torrents/sample.torrent
```

```sh
./bittorrent.sh download -o /Users/william/Documents/bittorrent/tmp/codercat.gif torrents/codercat.gif.torrent
```

```sh
./bittorrent.sh download -o /Users/william/Documents/bittorrent/tmp/congratulations.gif torrents/congratulations.gif.torrent
```

```sh
./bittorrent.sh download -o /Users/william/Documents/bittorrent/tmp/debian.iso torrents/debian.torrent
```

```sh
./bittorrent.sh download -o /Users/william/Documents/bittorrent/tmp/itsworking.gif torrents/itsworking.gif.torrent
```

## Limitations

- **No support for magnet links**: Ensure you have a valid `.torrent` file.
- **HTTP trackers only**: Make sure the tracker URL is an HTTP link; UDP trackers are not supported.
- **Single-file torrents**: Multi-file torrents are not supported, so use only single-file torrents.
- **Leech-only client**: This client does not upload pieces, so it will not contribute to the sharing process.
- **No DHT support**: The Distributed Hash Table (DHT) protocol is not implemented.

## Planned Features

We are planning to implement the following features in future updates:

- **Seeding**: Ability to upload pieces to contribute to the network.
- **Multi-torrent support**: Download multiple torrents simultaneously.
- **DHT protocol**: Support for the Distributed Hash Table protocol to enhance peer discovery.

## Installation

Ensure you have the necessary permissions to run the script and place it in a directory included in your system's PATH.

```sh
chmod +x bittorrent.sh
```