package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

//Peer contains the following data associated with a connected peer-
//Conn - The TCP connection with that peer
type Peer struct {
	Conn           *net.TCPConn
	closeChan      chan Peer
	connectedAt    uint32
	connected      bool
	username       string
	msgChan        chan []byte
	stopMsgChan    chan bool
	sendingFiles   MultipleFiles
	receivingFiles MultipleFiles
	sendMutex      sync.Mutex
}

func (peer *Peer) initPeer() {
	peer.createMsgChan()
	go peer.listenForMessages()
	peer.setPing()
	//peer.sendingFiles = []File{}
	//peer.receivingFiles = []File{}
}

func (peer *Peer) sendMessage(msg []byte) error {
	peer.sendMutex.Lock()
	_, err := peer.Conn.Write(msg)
	peer.sendMutex.Unlock()
	return err
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
	peer.sendMessage(pingMessage)
}

func (peer *Peer) listenForMessages() {
	for {
		msg := <-peer.msgChan
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
		case "file_data":
			peer.fileDataHandler(msg)
		}

	}
}

func (peer *Peer) fileInfoHandler(fileInfoMsg []byte) {
	fileName := string(fileInfoMsg[46:])
	md5 := string(fileInfoMsg[10:42])
	uniqueID := binary.BigEndian.Uint32(fileInfoMsg[42:46])
	fileLength := binary.BigEndian.Uint64(fileInfoMsg[2:10])
	fmt.Println("File name: ", fileName)
	//Get user approval before sending acceptance message here
	fmt.Println("Sending acceptance message")
	file := File{
		filePath:           fileName,
		fileSize:           fileLength,
		transferredSize:    0,
		handshake_complete: true,
		md5:                md5,
		uniqueID: 	    uniqueID,
	}
	peer.receivingFiles = peer.receivingFiles.add(file)
	peer.receivingFiles = peer.receivingFiles.openForWriting(md5)
	fileAcceptMsg := make([]byte, len(fileInfoMsg)+4)
	getBytesFromUint32(fileAcceptMsg[0:4], uint32(len(fileInfoMsg)))
	fileAcceptMsg[4] = 4
	copy(fileAcceptMsg[5:], fileInfoMsg[1:])
	peer.sendMessage(fileAcceptMsg)
}

func (peer *Peer) fileAcceptHandler(fileInfoMsg []byte) {
	md5 := string(fileInfoMsg[10:42])
	uniqueID := binary.BigEndian.Uint32(fileInfoMsg[42:46])
	peer.sendingFiles = peer.sendingFiles.updateAfterHandshake(md5)
	go peer.sendFileData(uniqueID)
}

func (peer *Peer) transferFile() {}

func (peer *Peer) createMsgChan() {
	msgChan := make(chan []byte)
	peer.stopMsgChan = make(chan bool)
	fmt.Println("Chan created")
	go func() {
		for {
			select {
			case <-peer.stopMsgChan:
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

func (peer Peer) sendChatMessage(msgContent string) error {
	chatMsg := getChatMsg(msgContent)
	return peer.sendMessage(chatMsg)
}

func (peer Peer) chatHandler(msgContent []byte) {
	fmt.Println(peer.username+": ", string(msgContent))
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

func (peer Peer) getIPWithPort() string {
	return peer.Conn.RemoteAddr().String()
}

func (peer Peer) getIPWithoutPort() string {
	return strings.Split(peer.Conn.RemoteAddr().String(), ":")[0]
}

func (peer *Peer) sendFile(filePath string) {
	file, _ := newFile(strings.TrimSpace(filePath))
	fileMsg := getFileInfoMsg(file.fileSize, file.getFileName(), file.md5, file.uniqueID)
	peer.sendingFiles = peer.sendingFiles.add(file)
	peer.sendMessage(fileMsg)
}

func (peer *Peer) sendFileData(uniqueID uint32) {
	fileToSend := peer.sendingFiles.get(uniqueID)
	nextBytes := fileToSend.getNextBytes()
	for len(nextBytes) > 0 {
		fileDataMsg := getFileDataMsg(nextBytes, fileToSend.uniqueID)
		peer.sendMessage(fileDataMsg)
		peer.sendingFiles = peer.sendingFiles.update(fileToSend)
		nextBytes = fileToSend.getNextBytes()
	}
}

func (peer *Peer) fileDataHandler(fileDataMsg []byte) {
	uniqueID, fileData := extractFileDataFromMsg(fileDataMsg)
	fileToWrite := peer.receivingFiles.get(uniqueID)
	fileToWrite.writeBytes(fileData)
	peer.receivingFiles = peer.receivingFiles.update(fileToWrite)
}
