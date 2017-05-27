package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"flag"
)

func main() {
	usernamePtr := flag.String("u", "", "Desired username")
	flag.Parse()
	ServerAddr, err := net.ResolveUDPAddr("udp", ":7041")
	if err != nil {
		fmt.Println("Err while resolving IP address", err)
	}
	peerConnections := make(map[string]Peer)
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		fmt.Println("Err while listening for connections", err)
		return
	}
	closeChan := make(chan Peer)
	peerManager := PeerManager{closeChan: closeChan, connectedPeers: peerConnections}
	go peerManager.init()
	go waitForTCP(peerManager, l, *usernamePtr)
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	ListenerAddr := l.Addr()
	LocalAddr := initUDPBroadcast(ListenerAddr, peerConnections, byte(0))

	if err != nil {
		fmt.Println("Err while listening to the address", err)
	}
	tcpListenerAddr := strings.Split(l.Addr().String(), ":")
	port, _ := strconv.Atoi(tcpListenerAddr[len(tcpListenerAddr)-1])
	go listenForUDPBroadcast(ServerConn, LocalAddr, peerManager, port, byte(0))
	startCli(peerManager)
}
