package main

import (
	"encoding/binary"
)

func getPingMsg() []byte {
	pingMsg := make([]byte, 5)
	copy(pingMsg[0:4], []byte{0, 0, 0, 1})
	copy(pingMsg[4:5], []byte{0})
	return pingMsg
}

func getPongMsg() []byte {
	pingMsg := make([]byte, 5)
	copy(pingMsg[0:4], []byte{0, 0, 0, 1})
	copy(pingMsg[4:5], []byte{1})
	return pingMsg
}

func getChatMsg(msgContent string) []byte {
	chatMsg := make([]byte, len(msgContent)+5)
	getBytesFromUint32(chatMsg[0:4], uint32(len(msgContent)+1))
	copy(chatMsg[4:5], []byte{2})
	copy(chatMsg[5:], msgContent)
	return chatMsg
}

func extractChatMsg(chatMsg []byte) []byte {
	return chatMsg[1:]
}

func getFileInfoMsg(fileLen uint64, fileName string, md5 string, uniqueID uint32) []byte {
	fileNameLen := uint8(len(fileName))
	fileMsgLen := 10 + fileNameLen + 32 + 4
	fileMsg := make([]byte, fileMsgLen+4)
	getBytesFromUint32(fileMsg[0:4], uint32(fileMsgLen))
	fileMsg[4] = 3
	fileMsg[5] = fileNameLen
	getBytesFromUint64(fileMsg[6:], fileLen)
	copy(fileMsg[14:], md5)
	getBytesFromUint32(fileMsg[46:50], uniqueID)
	copy(fileMsg[50:], fileName)
	return fileMsg
}

func getFileDataMsg(fileData []byte, uniqueID uint32) []byte {
	fileDataMsg := make([]byte, 5+len(fileData)+32)
	msgLen := len(fileData) + 32
	getBytesFromUint32(fileDataMsg[0:4], uint32(msgLen)+1)
	fileDataMsg[4] = 5
	getBytesFromUint32(fileDataMsg[5:37], uniqueID)
	copy(fileDataMsg[37:], fileData)
	return fileDataMsg
}

func extractFileDataFromMsg(fileMsg []byte) (uint32, []byte) {
	uniqueID := binary.BigEndian.Uint32(fileMsg[1:33])
	fileData := fileMsg[33:]
	return uniqueID, fileData
}

func getMsgType(msg []byte) string {
	availableMsgTypes := map[byte]string{
		0: "ping",
		1: "pong",
		2: "chat",
		3: "file_info",
		4: "file_accept",
		5: "file_data",
	}
	msgType := availableMsgTypes[msg[0]]
	return msgType
}
