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

func getFileInfoMsg(fileLen uint64, fileName string, md5 string) []byte {
	fileNameLen := uint8(len(fileName))
	fileMsgLen := 10+fileNameLen+32
	fileMsg := make([]byte, fileMsgLen+4)
	getBytesFromUint32(fileMsg[0:4], uint32(fileMsgLen))
	fileMsg[4] = 3
	fileMsg[5] = fileNameLen
	getBytesFromUint64(fileMsg[6:], fileLen)
	copy(fileMsg[14:], md5)
	copy(fileMsg[46:], fileName)
	return fileMsg
}

func getMsgType(msg []byte) string {
	availableMsgTypes := map[byte]string{
		0: "ping",
		1: "pong",
		2: "chat",
		3: "file_info",
		4: "file_accept",
	}
	msgType := availableMsgTypes[msg[0]]
	return msgType
}
