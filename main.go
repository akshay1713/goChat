package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"time"
	"io"
	"encoding/binary"
	"strings"
)

func main() {
	gitPtr := flag.Bool("g", false, "Use your github username & connect "+
		"only to the collaborators of the current repository")
	var usernamePtr *string
	if !*gitPtr {
		usernamePtr = flag.String("u", "", "Desired username")
	}
	flag.Parse()
	if *usernamePtr == "" {
		fmt.Println("Please specify a username or chat as your github avatar")
		return
	}
	peerConnections := make(map[string]*Peer)
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		fmt.Println("Err while listening for connections", err)
		return
	}
	closeChan := make(chan Peer)
	peerManager := PeerManager{closeChan: closeChan, connectedPeers: peerConnections}
	go peerManager.init()
	initDiscovery(peerManager, usernamePtr, l)
	startCli(&peerManager)
}

func initDiscovery(peerManager PeerManager, usernamePtr *string, l net.Listener){
	candidatePorts := []string{"7041", "7042"}
	connChan := make(chan ConnAndType)
	go func(){
		for connAndType := range connChan{
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
	go waitForTCP(peerManager, l, connChan)
	toSendPort := strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
	portInt, _ := strconv.Atoi(toSendPort)
	udpBroadcaster := UDPBroadcaster{ports: candidatePorts, appName: "goChat"}
	broadcastListenerConn := udpBroadcaster.init(peerManager)
	possibleLocalPorts := udpBroadcaster.broadcastOnAllPorts(toSendPort)
	udpListener := UDPListener{
		listenerPort:       portInt,
		listenerConn:       broadcastListenerConn,
		possibleLocalAddrs: possibleLocalPorts,
		appName: 	    "goChat",
	}
	go udpListener.listenForUDPBroadcast(peerManager)
}
