package main

import ()

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

func getFileInfoMsg(fileLen uint64, fileName string) []byte {
	fileNameLen := uint8(len(fileName))
	fileMsg := make([]byte, 10+fileNameLen)
	fileMsg[0] = 3
	fileMsg[1] = fileNameLen
	getBytesFromUint64(fileMsg[2:], fileLen)
	fileMsg = append(fileMsg, fileName...)
	return fileMsg
}

func getMsgType(msg []byte) string {
	availableMsgTypes := map[byte]string{
		0: "ping",
		1: "pong",
		2: "chat",
		3: "file_info",
	}
	msgType := availableMsgTypes[msg[0]]
	return msgType
}
