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
	Conn      *net.TCPConn
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
}

func (peer Peer) listenForMessages() {
	for {
		msg, err := peer.getNextMessage()
		handleErr(err, "Error while getting message from peer ")
		msgType := getMsgType(msg)
		switch msgType {
		case "ping":
			peer.pingHandler()
		case "chat":
			msgContent := extractChatMsg(msg)
			peer.chatHandler(msgContent)
		}

	}
}

func (peer Peer) getNextMessage() ([]byte, error) {
	msgLength := 4
	lengthMsg := make([]byte, msgLength)
	_, err := io.ReadFull(peer.Conn, lengthMsg)
	payloadLength := binary.BigEndian.Uint32(lengthMsg)
	msg := make([]byte, payloadLength)
	_, err = io.ReadFull(peer.Conn, msg)
	return msg, err
}

func (peer Peer) sendMessage(msgContent string) error {
	chatMsg := getChatMsg(msgContent)
	_, err := peer.Conn.Write(chatMsg)
	return err
}

func (peer Peer) chatHandler(msgContent []byte) {
	fmt.Println("Msg from peer: ", string(msgContent))
}

func (peer Peer) pingHandler() {
	fmt.Println("Ping received")
}

func (peer Peer) sendPong() {
	pongMessage := getPongMsg()
	peer.Conn.Write(pongMessage)
}

func (peer Peer) disConnect() {
	peer.Conn.Close()
	peer.closeChan <- peer
}

func (peer Peer) getIP() string {
	return peer.Conn.RemoteAddr().String()
}
