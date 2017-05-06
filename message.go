package main

import "encoding/binary"

func getPingMsg() []byte{
	pingMsg := make([]byte, 5)
	copy(pingMsg[0:4], []byte{0,0,0,1})
	copy(pingMsg[4:5], []byte{0})
	return pingMsg
}

func getPongMsg() []byte{
	pingMsg := make([]byte, 5)
	copy(pingMsg[0:4], []byte{0,0,0,1})
	copy(pingMsg[4:5], []byte{1})
	return pingMsg
}

func getChatMsg(msgContent string) []byte{
	chatMsg := make([]byte, len(msgContent) + 5)
	getBytesFromUint32(chatMsg[0:4], uint32(len(msgContent)))
	copy(chatMsg[4:5], []byte{2})
	copy(chatMsg[5:], msgContent)
	return chatMsg
}

func extractChatMsg(chatMsg []byte) []byte{
	msgLen := binary.BigEndian.Uint32(chatMsg[0:4])
	return chatMsg[5: msgLen]
}

func getMsgType(msg []byte) string{
	availableMsgTypes :=  map[byte]string{
		0: "ping",
		1: "pong",
		2: "chat",
	}
	msgType := availableMsgTypes[msg[0]]
	return msgType
}