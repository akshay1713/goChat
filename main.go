package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func main() {
	ServerAddr, err := net.ResolveUDPAddr("udp", ":7041")
	if err != nil {
		fmt.Println("Err while resolving IP address", err)
	}
	peerConnections := make(map[string]*net.TCPConn)
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	LocalAddr := initUDPBroadcast(peerConnections)

	if err != nil {
		fmt.Println("Err while listening to the address", err)
	}
	defer ServerConn.Close()

	buf := make([]byte, 1024)

	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		if addr.IP.String()+":"+strconv.Itoa(addr.Port) == LocalAddr.String() {
			continue
		}
		if string(buf[0:n]) != "goChat" {
			continue
		}

		if err != nil {
			fmt.Println("Error: ", err)
		}

		if _, exists := peerConnections[addr.IP.String()]; !exists {
			newConnection, err := connectToPeer(addr)
			if err != nil {
				fmt.Println("Err while connecting to the source of broadcase message", err)
				continue
			}

			fmt.Println("New peer joined", newConnection.RemoteAddr().String())
			addPeerConnection(peerConnections, newConnection)
		}
	}
}

func addPeerConnection(peerConnections map[string]*net.TCPConn, conn *net.TCPConn) {
	peerAddress := conn.RemoteAddr().String()
	peerIP := strings.Split(peerAddress, ":")[0]
	peerConnections[peerIP] = conn
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(15 * time.Second)
}

func connectToPeer(udpAddr *net.UDPAddr) (*net.TCPConn, error) {
	tcpAddr := net.TCPAddr{IP: udpAddr.IP, Port: udpAddr.Port}
	chatConn, err := net.DialTCP("tcp", nil, &tcpAddr)
	return chatConn, err
}

func initUDPBroadcast(peerConnections map[string]*net.TCPConn) net.Addr {
	ServerAddr, err := net.ResolveUDPAddr("udp", "192.168.1.255:7041")
	if err != nil {
		panic(err)
	}

	Conn, err := net.DialUDP("udp", nil, ServerAddr)
	LocalAddr := Conn.LocalAddr()
	if err != nil {
		panic(err)
	}
	go waitForTCP(LocalAddr, peerConnections)

	i := 0
	go func() {
		defer Conn.Close()
		for {
			msg := "goChat"
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

func waitForTCP(LocalAddr net.Addr, peerConnections map[string]*net.TCPConn) {
	ip, _, _ := net.ParseCIDR(strings.Split(LocalAddr.String(), ":")[0])
	port, _ := strconv.Atoi(strings.Split(LocalAddr.String(), ":")[1])
	l, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   ip,
		Port: port,
	})
	if err != nil {
		fmt.Println("Err while listening for connectionsl", err)
		return
	}
	for {
		conn, err := l.AcceptTCP()
		peerIP := strings.Split(conn.RemoteAddr().String(), ":")[0]
		if _, exists := peerConnections[peerIP]; !exists {
			if err != nil {
				fmt.Println("Error accepting: ", err.Error())
				continue
			}
			fmt.Println("Adding connection ", conn.RemoteAddr().String())
			addPeerConnection(peerConnections, conn)
		}
	}
}
