## Commands for Testing Functionality

### Decode Commands

```sh
./bittorrent.sh decode 5:hello
```
"hello"


```sh
./bittorrent.sh decode i52e
```
52


```sh
./bittorrent.sh decode l5:helloi52ee
```
["hello",52]


```sh
./bittorrent.sh decode d3:foo3:bar5:helloi52ee
```
{"foo":"bar","hello":52}


```sh
./bittorrent.sh info sample.torrent
```
Tracker URL: http://bittorrent-test-tracker.codecrafters.io/announce
Length: 92063
Info Hash: d69f91e6b2ae4c542468d1073a71d4ea13879a7f
Piece Length: 32768
Piece Hashes:
e876f67a2a8886e8f36b136726c30fa29703022d
6e2275e604a0766656736e81ff10b55204ad8d35
f00d937a0213df1982bc8d097227ad9e909acc17


```sh
./bittorrent.sh peers sample.torrent
```
178.62.82.89:51470
165.232.33.77:51467
178.62.85.20:51489


```sh
./bittorrent.sh handshake sample.torrent <peer_ip>:<peer_port>
```
Peer ID: 0102030405060708090a0b0c0d0e0f1011121314


```sh
./bittorrent.sh download_piece -o /Users/william/Documents/bittorrent/tmp/test-piece-0 sample.torrent 0
```
Piece 0 downloaded to /tmp/test-piece-0.


```sh
./bittorrent.sh download -o /Users/william/Documents/bittorrent/tmp/sample.txt sample.torrent
```
Downloaded sample.torrent to /Users/william/Documents/bittorrent/tmp/sample.txt.


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