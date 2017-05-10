package main

import (
	"net"
	"strings"
	"fmt"
)

type PeerManager struct{
	closeChan chan Peer
	connectedPeers map[string]Peer
}

func (peerManager PeerManager) addNewPeer(conn *net.TCPConn) Peer{
	newPeer := Peer{Conn: conn, closeChan: peerManager.closeChan}
	peerAddress := conn.RemoteAddr().String()
	peerIP := strings.Split(peerAddress, ":")[0]
	peerManager.connectedPeers[peerIP] = newPeer
	go newPeer.listenForMessages()
	return newPeer
}

func (peerManager PeerManager) init() {
	for {
		disconnectedPeer := <- peerManager.closeChan
		fmt.Println("Peer disconnected", disconnectedPeer)
	}
}

func (peerManager PeerManager) isConnected(IP string) bool{
	if _, exists := peerManager.connectedPeers[IP]; exists {
		return true
	}
	return false
}

func (peerManager PeerManager) sendMessage(message string) {
	for _, peer := range peerManager.connectedPeers {
		err := peer.sendMessage(message)
		if err != nil {
			peer.disConnect()
		}
	}
}

func (peerManager PeerManager) getAllIPs() []string{
	var peerIPs []string
	for _, peer := range peerManager.connectedPeers {
		peerIPs = append(peerIPs, peer.getIP())
	}
	return peerIPs
}
