package main

import (
	"fmt"
	"net"
)

//Peer contains the following data associated with a connected peer-
//Conn - The TCP connection with that peer
type Peer struct {
	Conn *net.TCPConn
	IP   net.IP
	Port int
}

func (peer *Peer) listenForMessage() {
	fmt.Println("Listening for Message from peer ", peer.IP, peer.Port)
}
