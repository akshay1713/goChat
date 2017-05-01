package main

import (
	"fmt"
	"net"
	"strconv"
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
	go waitForTCP(peerConnections, l)
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	ListenerAddr := l.Addr()
	LocalAddr := initUDPBroadcast(ListenerAddr, peerConnections)

	if err != nil {
		fmt.Println("Err while listening to the address", err)
	}
	defer ServerConn.Close()
	appName := "goChat"
	portLen := 5

	buf := make([]byte, len(appName)+portLen)

	for {
		_, addr, err := ServerConn.ReadFromUDP(buf)

		if addr.IP.String()+":"+strconv.Itoa(addr.Port) == LocalAddr.String() {
			continue
		}
		if string(buf[0:len(appName)]) != appName {
			continue
		}
		recvdPort, err := strconv.Atoi(string(buf[len(appName):]))

		if err != nil {
			fmt.Println("Error: ", err)
		}

		if _, exists := peerConnections[addr.IP.String()]; !exists {
			newConnection, err := connectToPeer(addr.IP, recvdPort)
			fmt.Println(peerConnections)
			if err != nil {
				fmt.Println("Err while connecting to the source of broadcase message", err)
				continue
			}
			addPeerConnection(peerConnections, newConnection)
			fmt.Println("New peer joined", newConnection.RemoteAddr().String())
		}
	}
}
