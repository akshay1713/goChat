package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func initUDPBroadcast(ListenerAddr net.Addr, peerConnections map[string]Peer) net.Addr {
	ServerAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:7041")
	if err != nil {
		panic(err)
	}

	Conn, err := net.DialUDP("udp", nil, ServerAddr)
	LocalAddr := Conn.LocalAddr()
	if err != nil {
		panic(err)
	}
	i := 0
	go func() {
		defer Conn.Close()
		var msg []byte
		appName := "goChat"
		msg = append(msg, appName...)
		port := strconv.Itoa(ListenerAddr.(*net.TCPAddr).Port)

		port = padLeft(port, "0", 5)
		msg = append(msg, port...)
		fmt.Println("Port found is ", port)
		for {
			i++
			buf := []byte(msg)
			_, err := Conn.Write(buf)
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(time.Second * 1)
		}
	}()
	return LocalAddr
}

func addPeerConnection(peerConnections map[string]Peer, conn *net.TCPConn) Peer{
	peerAddress := conn.RemoteAddr().String()
	peerIP := strings.Split(peerAddress, ":")[0]
	newPeer := Peer{Conn: conn}
	peerConnections[peerIP] = newPeer
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(15 * time.Second)
	go newPeer.listenForMessages()
	return newPeer
}

func connectToPeer(ip net.IP, port int) (*net.TCPConn, error) {
	fmt.Println("Connecting to ", ip, port)
	tcpAddr := net.TCPAddr{IP: ip, Port: port}
	chatConn, err := net.DialTCP("tcp", nil, &tcpAddr)
	return chatConn, err
}

func waitForTCP(peerConnections map[string]Peer, listener net.Listener) {
	defer listener.Close()
	for {
		genericConn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error while listening in waitforTCP", err)
		}
		conn := genericConn.(*net.TCPConn)
		peerIP := strings.Split(conn.RemoteAddr().String(), ":")[0]
		if _, exists := peerConnections[peerIP]; !exists {
			if err != nil {
				fmt.Println("Error accepting: ", err.Error())
				continue
			}
			fmt.Println("Adding connection ", conn.RemoteAddr().String())
			newPeer := addPeerConnection(peerConnections, conn)
			newPeer.setPing()
		}
	}
}
