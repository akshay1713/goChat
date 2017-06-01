package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

type File struct {
	filePath   string
	percentage int
	handshake_complete bool
}
//Peer contains the following data associated with a connected peer-
//Conn - The TCP connection with that peer
type Peer struct {
	Conn        *net.TCPConn
	closeChan   chan Peer
	connectedAt uint32
	connected   bool
	username    string
	msgChan     chan []byte
	stopMsgChan chan bool
	sendingFiles []File
	receivingFiles []File
}

func(peer *Peer) initPeer() {
	peer.createMsgChan()
	go peer.listenForMessages()
	peer.setPing()
}

func (peer Peer) setPing() {
	fmt.Println("Setting Ping")
	// Do NOT forget to increase this time later
	time.AfterFunc(2*time.Second, peer.sendPing)
}

func (peer Peer) sendPing() {
	if !peer.connected {
		fmt.Println("Stopping ping")
		return
	}
	time.AfterFunc(2*time.Second, peer.sendPing)
	pingMessage := getPingMsg()
	peer.Conn.Write(pingMessage)
}

func (peer Peer) listenForMessages() {
	for {
		msg := <- peer.msgChan
		if len(msg) == 0 {
			return
		}
		msgType := getMsgType(msg)
		switch msgType {
		case "ping":
			peer.pingHandler()
		case "chat":
			msgContent := extractChatMsg(msg)
			peer.chatHandler(msgContent)
		case "file_info":
			peer.fileInfoHandler(msg)
		case "file_accept":
			peer.fileAcceptHandler(msg)
		}

	}
}

func (peer *Peer) fileInfoHandler(fileInfoMsg []byte) {
	fmt.Println("File info message received", fileInfoMsg)
	fmt.Println("Name length: ", int(fileInfoMsg[1]))
	fmt.Println("File length: ", binary.BigEndian.Uint64(fileInfoMsg[2:10]))
	fmt.Println("File name: ", string(fileInfoMsg[10:]))
	peer.Conn.Write(fileInfoMsg)
}

func (peer *Peer) fileAcceptHandler(fileInfoMsg []byte) {
	fmt.Println("File info message received", fileInfoMsg)
	fmt.Println("Name length: ", int(fileInfoMsg[1]))
	fmt.Println("File length: ", binary.BigEndian.Uint64(fileInfoMsg[2:10]))
	fmt.Println("File name: ", string(fileInfoMsg[10:]))
}

func (peer *Peer) createMsgChan() {
	msgChan := make(chan []byte)
	peer.stopMsgChan = make(chan bool)
	fmt.Println("Chan created")
	go func(){
		for {
			select {
			case <- peer.stopMsgChan:
				fmt.Println("Stopping poll func")
				return
			default:
				msg, err := peer.getNextMessage()
				if len(msg) == 0 || err != nil {
					peer.disConnect()
					peer.stopMsgChan <- true
					return
				}
				msgChan <- msg
			}
		}
	}()
	peer.msgChan = msgChan
}

func (peer Peer) stopMsgLoop() {
	peer.stopMsgChan <- true
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
	fmt.Println(peer.username,": ", string(msgContent))
}

func (peer Peer) pingHandler() {
	//fmt.Println("Ping received")
}

func (peer Peer) sendPong() {
	pongMessage := getPongMsg()
	peer.Conn.Write(pongMessage)
}

func (peer *Peer) disConnect() {
	fmt.Println(peer.username, " disconnected")
	peer.Conn.Close()
	peer.connected = false
	peer.closeChan <- *peer
	close(peer.msgChan)
}

func (peer Peer) getIP() string {
	return peer.Conn.RemoteAddr().String()
}

func (peer *Peer) sendFile(filePath string) {
	fileMsg := getFileInfoMsg(6000, filePath)
	fmt.Println("Sending file", fileMsg)
	peer.Conn.Write(fileMsg)
}
