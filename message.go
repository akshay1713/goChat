package main

func getPingMsg() []byte{
	pingMsg := make([]byte, 5)
	copy(pingMsg[0:4], []byte{0,0,0,1})
	copy(pingMsg[4:5], []byte{0})
	return pingMsg
}

func getMsgType(msg []byte) string{
	availableMsgTypes :=  map[byte]string{
		0: "ping",
		1: "pong",
	}
	msgType := availableMsgTypes[msg[0]]
	return msgType
}