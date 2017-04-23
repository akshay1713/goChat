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
	newPeerChan := make(chan string)
	if err != nil {
		fmt.Println("Err while resolving IP address", err)
	}
	peerConnections := make(map[string]*net.TCPConn)
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	LocalAddr := initUDPBroadcast(newPeerChan)

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
			fmt.Println("New peer joined")
			peerConnections[addr.IP.String()] = newConnection
		}
	}
}

func addNewPeerConnection(peerAddr string, peerConn *net.TCPConn) {}

func connectToPeer(udpAddr *net.UDPAddr) (*net.TCPConn, error) {
	tcpAddr := net.TCPAddr{IP: udpAddr.IP, Port: udpAddr.Port}
	chatConn, err := net.DialTCP("tcp", nil, &tcpAddr)
	return chatConn, err
}

func initUDPBroadcast(newPeerChan chan string) net.Addr {
	ServerAddr, err := net.ResolveUDPAddr("udp", "192.168.1.255:7041")
	if err != nil {
		panic(err)
	}

	Conn, err := net.DialUDP("udp", nil, ServerAddr)
	LocalAddr := Conn.LocalAddr()
	if err != nil {
		panic(err)
	}
	go waitForTCP(LocalAddr, newPeerChan)

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

func waitForTCP(LocalAddr net.Addr, newPeerChan chan string) {
	fmt.Println(LocalAddr)
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
		conn, err := l.Accept()
		remoteAddr := conn.RemoteAddr()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			continue
		}
		newPeerChan <- remoteAddr.String()
	}
}
