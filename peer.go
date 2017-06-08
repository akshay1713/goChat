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

//sendMessage is the route through which all messages are sent to a peer.
//Uses a mutex(not strictly necessary)
func (peer *Peer) sendMessage(msg []byte) error {
	peer.sendMutex.Lock()
	_, err := peer.Conn.Write(msg)
	peer.sendMutex.Unlock()
	return err
}

func (peer Peer) setPing() {
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

//fileInfoHandler handles the message which informs the user that a peer wants to send a file.
//Sends a file acceptance message to that peer and creates a File object with the necessary parameters
func (peer *Peer) fileInfoHandler(fileInfoMsg []byte) {
	fileName := string(fileInfoMsg[46:])
	md5 := string(fileInfoMsg[10:42])
	uniqueID := binary.BigEndian.Uint32(fileInfoMsg[42:46])
	fileLength := binary.BigEndian.Uint64(fileInfoMsg[2:10])
	fmt.Println("Receiving file ", fileName, "from ", peer.username)
	//Get user approval before sending acceptance message here
	fmt.Println("Sending acceptance message")
	file := File{
		filePath:           fileName,
		fileSize:           fileLength,
		transferredSize:    0,
		handshake_complete: true,
		md5:                md5,
		uniqueID:           uniqueID,
	}
	peer.receivingFiles = peer.receivingFiles.add(file)
	peer.receivingFiles = peer.receivingFiles.openForWriting(md5)
	fileAcceptMsg := make([]byte, len(fileInfoMsg)+4)
	getBytesFromUint32(fileAcceptMsg[0:4], uint32(len(fileInfoMsg)))
	fileAcceptMsg[4] = 4
	copy(fileAcceptMsg[5:], fileInfoMsg[1:])
	peer.sendMessage(fileAcceptMsg)
}

//fileAcceptanceHandler handles a file acceptance message from a peer to whom the user had earlier sent a file info
// message, letting it know that it wants to initiate a file transfer. Updates the relevant File object, and starts
//sending the file
func (peer *Peer) fileAcceptHandler(fileInfoMsg []byte) {
	md5 := string(fileInfoMsg[10:42])
	uniqueID := binary.BigEndian.Uint32(fileInfoMsg[42:46])
	peer.sendingFiles = peer.sendingFiles.updateAfterHandshake(md5)
	go peer.sendFileData(uniqueID)
}

//createMsgChan creates a chan into which all the messages sent by a peer will be sent
func (peer *Peer) createMsgChan() {
	msgChan := make(chan []byte)
	peer.stopMsgChan = make(chan bool)
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

//getNextMessage gets the next message from a connected peer. Each message is preceded by 4 bytes containing the length
//of the actual message. The first byte of the actual message identifies the type of the message
func (peer Peer) getNextMessage() ([]byte, error) {
	msgLength := 4
	lengthMsg := make([]byte, msgLength)
	_, err := io.ReadFull(peer.Conn, lengthMsg)
	payloadLength := binary.BigEndian.Uint32(lengthMsg)
	msg := make([]byte, payloadLength)
	_, err = io.ReadFull(peer.Conn, msg)
	return msg, err
}

//sendChatMessage sends a chat message to the peer
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

//sendFile sends a file info message to a peer, containing information regarding the file which needs to be sent
//The actual file isn't sent until a file acceptance message is received from the peer
func (peer *Peer) sendFile(filePath string) {
	file, err := newFile(strings.TrimSpace(filePath))
	if err != nil {
		fmt.Println("Err while sending file: ", err, "\nAre you sure the file exists?")
		return
	}
	fileMsg := getFileInfoMsg(file.fileSize, file.getFileName(), file.md5, file.uniqueID)
	peer.sendingFiles = peer.sendingFiles.add(file)
	peer.sendMessage(fileMsg)
}

//sendFileData sends the actual file to a peer, who has sent a file acceptance message for a file info message sent
//to it earlier
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

//fileDataHandler receives the file data for a file being transferred to the user by a peer. The data is written to a
//file
func (peer *Peer) fileDataHandler(fileDataMsg []byte) {
	uniqueID, fileData := extractFileDataFromMsg(fileDataMsg)
	fileToWrite := peer.receivingFiles.get(uniqueID)
	fileToWrite.writeBytes(fileData)
	peer.receivingFiles = peer.receivingFiles.update(fileToWrite)
}
