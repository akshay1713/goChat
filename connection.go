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
		for i < 10 {
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

func listenForUDPBroadcast(ServerConn *net.UDPConn, LocalAddr net.Addr, peerManager PeerManager) {
	defer ServerConn.Close()
	appName := "goChat"
	portLen := 5
	buf := make([]byte, len(appName)+portLen)
	for {
		_, addr, err := ServerConn.ReadFromUDP(buf)

		if addr.IP.String()+":"+strconv.Itoa(addr.Port) == LocalAddr.String() {
			continue
		}
		if string(buf[0:len(appName)]) != appName {
			continue
		}
		recvdPort, err := strconv.Atoi(string(buf[len(appName):]))

		if err != nil {
			fmt.Println("Error: ", err)
		}

		if !peerManager.isConnected(addr.IP.String()) {
			newConnection, err := connectToPeer(addr.IP, recvdPort)
			if err != nil {
				fmt.Println("Err while connecting to the source of broadcase message", err)
				continue
			}
			newPeer := peerManager.addNewPeer(newConnection)
			go newPeer.setPing()
			fmt.Println("New peer joined", newConnection.RemoteAddr().String())
		}
	}
}

func connectToPeer(ip net.IP, port int) (*net.TCPConn, error) {
	fmt.Println("Connecting to ", ip, port)
	tcpAddr := net.TCPAddr{IP: ip, Port: port}
	chatConn, err := net.DialTCP("tcp", nil, &tcpAddr)
	return chatConn, err
}

func waitForTCP(peerManager PeerManager, listener net.Listener) {
	defer listener.Close()
	for {
		genericConn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error while listening in waitforTCP", err)
		}
		conn := genericConn.(*net.TCPConn)
		peerIP := strings.Split(conn.RemoteAddr().String(), ":")[0]
		if !peerManager.isConnected(peerIP) {
			if err != nil {
				fmt.Println("Error accepting: ", err.Error())
				continue
			}
			fmt.Println("Adding connection ", conn.RemoteAddr().String())
			newPeer := peerManager.addNewPeer(conn)
			newPeer.setPing()
		}
	}
}
