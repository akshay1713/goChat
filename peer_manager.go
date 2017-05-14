package main

import (
	"net"
	"strings"
	"fmt"
	"encoding/binary"
)

type PeerManager struct{
	closeChan chan Peer
	connectedPeers map[string]Peer
}

func (peerManager *PeerManager) addNewPeer(conn *net.TCPConn, currentTimestamp uint32) Peer{
	newPeer := Peer{
		Conn: conn,
		closeChan: peerManager.closeChan,
		connectedAt: currentTimestamp,
	}
	peerAddress := conn.RemoteAddr().String()
	if existingPeer, exists := peerManager.connectedPeers[peerAddress]; exists {
		fmt.Println("Exists, disconnecting")
		existingPeer.disConnect()
	}
	peerIP := strings.Split(peerAddress, ":")[0]
	peerManager.connectedPeers[peerIP] = newPeer
	timestampBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(timestampBytes, uint32(currentTimestamp))
	conn.Write(timestampBytes)
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
	fmt.Print("sending to peers ", message)
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

func (peerManager PeerManager) getPeer(IP string) Peer{
	return peerManager.connectedPeers[IP]
}

func (peerManager *PeerManager) updatePeer(IP string, conn *net.TCPConn, connectedAt uint32) {
	peerToUpdate := peerManager.connectedPeers[IP]
	peerToUpdate.disConnect()
	peerToUpdate.Conn = conn
	peerToUpdate.connectedAt = connectedAt
	fmt.Println("Starting loop again")
	go peerToUpdate.listenForMessages()
}
