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
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		fmt.Println("Err while listening for connectionsl", err)
		return
	}
	fmt.Println(l.Addr().String())
	go waitForTCP(peerConnections, l)
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	ListenerAddr := l.Addr()
	LocalAddr := initUDPBroadcast(ListenerAddr, peerConnections)

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

func initUDPBroadcast(ListenerAddr net.Addr, peerConnections map[string]*net.TCPConn) net.Addr {
	ServerAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:7041")
	if err != nil {
		panic(err)
	}

	Conn, err := net.DialUDP("udp", nil, ServerAddr)
	LocalAddr := Conn.LocalAddr()
	if err != nil {
		panic(err)
	}
	// go waitForTCP(LocalAddr, peerConnections)

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

func waitForTCP(peerConnections map[string]*net.TCPConn, listener net.Listener) {
	// ip, _, _ := net.ParseCIDR(strings.Split(LocalAddr.String(), ":")[0])
	// port, _ := strconv.Atoi(strings.Split(LocalAddr.String(), ":")[1])
	fmt.Println("LISTEN TO ME!")
	defer listener.Close()
	fmt.Println("Listening on", listener.Addr().String)
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
			addPeerConnection(peerConnections, conn)
		}
	}
}
