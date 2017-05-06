package main

import (
	"fmt"
	"net"
)

func main() {
	ServerAddr, err := net.ResolveUDPAddr("udp", ":7041")
	if err != nil {
		fmt.Println("Err while resolving IP address", err)
	}
	peerConnections := make(map[string]Peer)
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		fmt.Println("Err while listening for connectionsl", err)
		return
	}
	closeChan := make(chan Peer)
	peerManager := PeerManager{closeChan: closeChan, connectedPeers: peerConnections}
	go peerManager.init()
	go waitForTCP(peerManager, l)
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	ListenerAddr := l.Addr()
	LocalAddr := initUDPBroadcast(ListenerAddr, peerConnections)

	if err != nil {
		fmt.Println("Err while listening to the address", err)
	}
	go listenForUDPBroadcast(ServerConn, LocalAddr, peerManager)
	startCli(peerManager)
}
