package main

import (
	"fmt"
	"net"
	"strings"
	"strconv"
)

func main() {
	ServerAddr, err := net.ResolveUDPAddr("udp", ":7041")
	if err != nil {
		fmt.Println("Err while resolving IP address", err)
	}
	peerConnections := make(map[string]Peer)
	expectingConnections := make(map[string]bool)
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		fmt.Println("Err while listening for connectionsl", err)
		return
	}
	closeChan := make(chan Peer)
	peerManager := PeerManager{
		closeChan: closeChan,
		connectedPeers: peerConnections,
		expectingConnections: expectingConnections,
	}
	go peerManager.init()
	go waitForTCP(&peerManager, l)
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	ListenerAddr := l.Addr()
	fmt.Println(ListenerAddr)
	LocalAddr := initUDPBroadcast(ListenerAddr, peerConnections)

	if err != nil {
		fmt.Println("Err while listening to the address", err)
	}
	tcpListenerAddr := strings.Split(l.Addr().String(), ":")
	port, _ := strconv.Atoi(tcpListenerAddr[len(tcpListenerAddr) - 1])
	go listenForUDPBroadcast(ServerConn, LocalAddr, &peerManager, port)
	startCli(peerManager)
}
