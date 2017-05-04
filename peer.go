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
	closeChan chan Peer
}

func (peer Peer) setPing() {
	fmt.Println("Setting Ping")
	// Do NOT forget to increase this time later
	time.AfterFunc(2*time.Second, peer.sendPing)
}

func (peer Peer) sendPing() {
	//fmt.Println("Sending Ping")
	time.AfterFunc(2*time.Second, peer.sendPing)
	pingMessage := getPingMsg()
	peer.Conn.Write(pingMessage)
	msg, err := peer.getNextMessage()
	handleErr(err, "Error while sending ping: ")
	if getMsgType(msg) != "pong" {
		fmt.Print("Response to ping not received")
		peer.disConnect()
	}
}

func (peer Peer) listenForMessages() {
	for {
		msg, err := peer.getNextMessage()
		handleErr(err, "Error while getting message from peer ")
		msgType := getMsgType(msg)
		switch msgType {
		case "ping":
			peer.pingHandler()
		}
		fmt.Println("Msg type is ", msgType)

	}
}

func (peer Peer) getNextMessage() ([]byte, error){
	msgLength := 4
	lengthMsg := make([]byte, msgLength)
	_, err := io.ReadFull(peer.Conn, lengthMsg)
	payloadLength := binary.BigEndian.Uint32(lengthMsg)
	msg := make([]byte, payloadLength)
	_, err = io.ReadFull(peer.Conn, msg)
	return msg, err
}

func (peer Peer) pingHandler() {
	peer.sendPong()
}

func (peer Peer) sendPong() {
	pongMessage := getPongMsg()
	peer.Conn.Write(pongMessage)
}

func (peer Peer) disConnect() {
	peer.Conn.Close()
	peer.closeChan <- peer
}
