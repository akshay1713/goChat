package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

//Peer contains the following data associated with a connected peer-
//Conn - The TCP connection with that peer
type Peer struct {
	Conn *net.TCPConn
}

func (peer *Peer) setPing() {
	fmt.Println("Setting Ping")
	time.AfterFunc(2*time.Second, peer.sendPing)
}

func (peer *Peer) sendPing() {
	fmt.Println("Sending Ping")
	time.AfterFunc(2*time.Second, peer.sendPing)
	pingMessage := getPingMsg()
	peer.Conn.Write(pingMessage)
}

func (peer *Peer) listenForMessages() {
	for {
		msgLength := 4
		lengthMsg := make([]byte, msgLength)
		_, err := io.ReadFull(peer.Conn, lengthMsg)
		handleErr(err, "Error while listening")
		payloadLength := binary.BigEndian.Uint32(lengthMsg)
		msg := make([]byte, payloadLength)
		_, err = io.ReadFull(peer.Conn, msg)
		fmt.Println("Msg received from peer is ", msg)

	}
}