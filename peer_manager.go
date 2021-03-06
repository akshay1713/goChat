package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

type PeerManager struct {
	closeChan      chan Peer
	connectedPeers map[string]*Peer
}

func (peerManager PeerManager) addNewPeer(conn *net.TCPConn, currentTimestamp uint32, initiated bool, username string) Peer {
	if initiated {
		conn.Write([]byte{1})
		currentTimestampBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(currentTimestampBytes, currentTimestamp)
		conn.Write(currentTimestampBytes)
	}
	usernameBytes := make([]byte, len(username)+2)
	binary.BigEndian.PutUint16(usernameBytes[0:2], uint16(len(username)))
	copy(usernameBytes[2:], username)
	conn.Write(usernameBytes)
	peerUsernameLenBytes := make([]byte, 2)
	conn.Read(peerUsernameLenBytes)
	peerUsernameLen := binary.BigEndian.Uint16(peerUsernameLenBytes)
	peerUsername := make([]byte, peerUsernameLen)
	conn.Read(peerUsername)
	newPeer := Peer{Conn: conn, closeChan: peerManager.closeChan, connected: true, username: string(peerUsername)}
	fmt.Println("Connected to ", string(peerUsername))
	peerAddress := conn.RemoteAddr().String()
	peerIP := strings.Split(peerAddress, ":")[0]
	peerManager.connectedPeers[peerIP] = &newPeer
	newPeer.initPeer()
	return newPeer
}

func (peerManager PeerManager) init() {
	for {
		disconnectedPeer := <-peerManager.closeChan
		fmt.Println(disconnectedPeer.username, " disconnected")
	}
}

func (peerManager PeerManager) IsConnected(IP string) bool {
	if _, exists := peerManager.connectedPeers[IP]; exists {
		return true
	}
	return false
}

func (peerManager PeerManager) sendMessage(message string) {
	for _, peer := range peerManager.connectedPeers {
		err := peer.sendChatMessage(message)
		if err != nil {
			peer.disConnect()
		}
	}
}

func (peerManager PeerManager) GetAllIPs() []string {
	var peerIPs []string
	for _, peer := range peerManager.connectedPeers {
		peerIPs = append(peerIPs, peer.getIPWithPort())
	}
	return peerIPs

}

func (peerManager *PeerManager) compareTimestampAndUpdate(conn *net.TCPConn, newTimestamp uint32, IP string) {
	peer, exists := peerManager.connectedPeers[IP]
	if !exists {
		fmt.Println("Peer to update not found", IP)
		return
	}
	if peer.connectedAt < newTimestamp {
		fmt.Println("current timestamp is older, not updating")
		return
	}
	fmt.Println("Updating existing peer")
	peer.disConnect()
	peer.Conn = conn
	peer.connectedAt = newTimestamp
	go peer.listenForMessages()
}

func (peerManager PeerManager) getAllPeers() []Peer {
	var connectedPeers []Peer
	for _, peer := range peerManager.connectedPeers {
		connectedPeers = append(connectedPeers, *peer)
	}
	return connectedPeers
}

func (peerManager PeerManager) getAllUserNames() []string {
	var usernames []string
	for _, peer := range peerManager.connectedPeers {
		usernames = append(usernames, peer.username)
	}
	return usernames
}

func (peerManager PeerManager) printAll() {
	for k, v := range peerManager.connectedPeers {
		fmt.Println(k, ":", v)
	}
}

func (peerManager *PeerManager) sendFiles(peerIPs []string, filepath string) {
	for i := range peerIPs {
		if _, exists := peerManager.connectedPeers[peerIPs[i]]; exists {
			peerManager.connectedPeers[peerIPs[i]].sendFile(filepath)
		} else {
			fmt.Println("Does not exist", peerIPs[i])
		}
	}
}
