package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/akshay1713/LANPeerDiscovery"
	"io"
	"strings"
	"time"
)

func main() {
	var usernamePtr *string
	usernamePtr = flag.String("u", "", "Desired username")
	flag.Parse()
	if *usernamePtr == "" {
		fmt.Println("Please specify a username")
		return
	}
	peerConnections := make(map[string]*Peer)
	closeChan := make(chan Peer)
	peerManager := PeerManager{closeChan: closeChan, connectedPeers: peerConnections}
	go peerManager.init()
	initDiscovery(peerManager, usernamePtr)
	startCli(&peerManager)
}

func initDiscovery(peerManager PeerManager, usernamePtr *string) {
	candidatePorts := []string{"7041", "7042"}
	connChan := LANPeerDiscovery.GetConnectionsChan(candidatePorts, peerManager, "goChat")
	go func() {
		for connAndType := range connChan {
			switch connAndType.Type {
			case "sender":
				currentTimestamp := uint32(time.Now().UTC().Unix())
				peerManager.addNewPeer(connAndType.Connection, currentTimestamp, true, *usernamePtr)
			case "receiver":
				recvdTimestampBytes := make([]byte, 4)
				_, err := io.ReadFull(connAndType.Connection, recvdTimestampBytes)
				handleErr(err, "While getting timestamp")
				recvdTimestamp := binary.BigEndian.Uint32(recvdTimestampBytes)
				peerManager.addNewPeer(connAndType.Connection, recvdTimestamp, false, *usernamePtr)
			case "duplicate_receiver":
				recvdTimestampBytes := make([]byte, 4)
				_, err := io.ReadFull(connAndType.Connection, recvdTimestampBytes)
				handleErr(err, "While getting timestamp")
				recvdTimestamp := binary.BigEndian.Uint32(recvdTimestampBytes)
				senderIPString := strings.Split(connAndType.Connection.RemoteAddr().String(), ":")[0]
				peerManager.compareTimestampAndUpdate(connAndType.Connection, recvdTimestamp, senderIPString)
			}
		}
	}()
}
