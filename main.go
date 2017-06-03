package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"flag"
)

func main() {
	gitPtr := flag.Bool("g", false, "Use your github username & connect " +
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
	go waitForTCP(peerManager, l, *usernamePtr)
	candidatePorts := []string{"7041", "7042"}
	var ServerConn *net.UDPConn
	ListenerAddr := l.Addr()
	possibleLocalAddrs := make([]string, len(candidatePorts))
	for i := range candidatePorts {
		ServerAddr, err := net.ResolveUDPAddr("udp", ":"+candidatePorts[i])
		if ServerConn == nil {
			ServerConn, err = net.ListenUDP("udp", ServerAddr)
		}
		tempLocalAddr := initUDPBroadcast(ListenerAddr, byte(0), candidatePorts[i])
		if err != nil {
			fmt.Println("Err while listening to the address", err)
		} else {
			possibleLocalAddrs = append(possibleLocalAddrs, tempLocalAddr.String())
		}
	}
	tcpListenerAddr := strings.Split(l.Addr().String(), ":")
	port, _ := strconv.Atoi(tcpListenerAddr[len(tcpListenerAddr)-1])
	go listenForUDPBroadcast(ServerConn, possibleLocalAddrs, peerManager, port, byte(0))
	startCli(peerManager)
}
