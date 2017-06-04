package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

func initUDPBroadcast(ListenerAddr net.Addr, chatType byte, port string) net.Addr {
	ServerAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:"+port)
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
		msg = append(msg, chatType)
		for i < 5 {
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

func listenForUDPBroadcast(ServerConn *net.UDPConn, possibleLocalAddrs []string, peerManager PeerManager, port int, chatType byte) {
	defer ServerConn.Close()
	appName := "goChat"
	portLen := 5
	typeLen := 1
	buf := make([]byte, len(appName)+portLen+typeLen)
	broadcastRecvdIPs := make(map[string]bool)
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(port))
	fmt.Println(portBytes)
	for {
		_, addr, err := ServerConn.ReadFromUDP(buf)

		if pos(possibleLocalAddrs, addr.IP.String()+":"+strconv.Itoa(addr.Port)) != -1 {
			continue
		}

		if string(buf[0:len(appName)]) != appName {
			continue
		}
		recvdPort, err := strconv.Atoi(string(buf[len(appName) : len(buf)-1]))

		if err != nil {
			fmt.Println("Error: ", err)
		}

		if buf[len(buf)-1] != chatType {
			continue
		}

		//if !peerManager.isConnected(addr.IP.String()) {
		if _, exists := broadcastRecvdIPs[addr.IP.String()]; !exists {
			broadcastRecvdIPs[addr.IP.String()] = true
			peerIPs := peerManager.getAllIPs()
			all_ips := []byte{}
			msgLengthBytes := make([]byte, 4)
			totalLen := 2 + len(peerIPs)*6
			binary.BigEndian.PutUint32(msgLengthBytes, uint32(totalLen))
			all_ips = append(all_ips, 0)
			all_ips = append(all_ips, msgLengthBytes...)
			all_ips = append(all_ips, portBytes...)
			fmt.Println(peerIPs)
			for i := range peerIPs {
				splitAddress := strings.Split(peerIPs[i], ":")
				peer_portBytes := make([]byte, 2)
				peer_port, _ := strconv.Atoi(splitAddress[1])
				binary.BigEndian.PutUint16(peer_portBytes, uint16(peer_port))
				splitIP := strings.Split(splitAddress[0], ".")
				for j := 0; j < 4; j++ {
					partIP, _ := strconv.Atoi(splitIP[j])
					all_ips = append(all_ips, byte(partIP))
				}
				all_ips = append(all_ips, peer_portBytes...)
			}
			tcpAddr := net.TCPAddr{IP: addr.IP, Port: recvdPort}
			sConn, err := net.DialTCP("tcp", nil, &tcpAddr)
			if err != nil {
				fmt.Println("Err while connecting to the source of broadcase message", err)
				continue
			}
			sConn.Write(all_ips)
		}
	}
}

func connectToPeer(ip net.IP, port int) (*net.TCPConn, error) {
	tcpAddr := net.TCPAddr{IP: ip, Port: port}
	chatConn, err := net.DialTCP("tcp", nil, &tcpAddr)
	return chatConn, err
}

func waitForTCP(peerManager PeerManager, listener net.Listener, username string) {
	defer listener.Close()
	for {
		genericConn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error while listening in waitforTCP", err)
		}
		conn := genericConn.(*net.TCPConn)
		senderIPString := strings.Split(conn.RemoteAddr().String(), ":")[0]
		senderIPOctets := strings.Split(senderIPString, ".")
		fmt.Println(senderIPString)
		handleErr(err, "Parsing IP")
		var senderIP []byte
		for i := 0; i < len(senderIPOctets); i++ {
			octetInt, _ := strconv.Atoi(senderIPOctets[i])
			senderIP = append(senderIP, byte(octetInt))
		}
		msgType := make([]byte, 1)
		_, err = io.ReadFull(conn, msgType)
		if !peerManager.isConnected(senderIPString) {
			if msgType[0] == 0 {
				//This msg is a list of IPs & ports
				msgLength := make([]byte, 4)
				_, err = io.ReadFull(conn, msgLength)
				handleErr(err, "Error while reading message ")
				peerInfoLength := binary.BigEndian.Uint32(msgLength)
				peerInfo := make([]byte, peerInfoLength)
				_, err = io.ReadFull(conn, peerInfo)
				handleErr(err, "Error while reading message ")
				senderPort := binary.BigEndian.Uint16([]byte{peerInfo[0], peerInfo[1]})
				//splitIP := strings.Split(senderIPString, ".")
				newConn, err := connectToPeer(senderIP, int(senderPort))
				handleErr(err, "Error while connecting to sender")
				if newConn == nil {
					fmt.Println("Nil conn")
					continue
				}
				currentTimestamp := uint32(time.Now().UTC().Unix())
				peerManager.addNewPeer(newConn, currentTimestamp, true, username)
				handleErr(err, "Error while connecting to sender")
				for k := 2; k < len(peerInfo); k += 6 {
					peerIP := net.IPv4(peerInfo[k+2], peerInfo[k+3], peerInfo[k+4], peerInfo[k+5])
					peerPort := binary.BigEndian.Uint16([]byte{peerInfo[k], peerInfo[k+1]})
					newConn, err = connectToPeer(peerIP, int(peerPort))
					if !peerManager.isConnected(peerIP.String()) {
						currentTimestamp = uint32(time.Now().UTC().Unix())
						peerManager.addNewPeer(newConn, currentTimestamp, true, username)
					}
				}
			} else {
				//This msg is a connection request
				recvdTimestampBytes := make([]byte, 4)
				_, err := io.ReadFull(conn, recvdTimestampBytes)
				handleErr(err, "While getting timestamp")
				recvdTimestamp := binary.BigEndian.Uint32(recvdTimestampBytes)
				peerManager.addNewPeer(conn, recvdTimestamp, false, username)
			}
		} else if msgType[0] == 1 {
			fmt.Println("Checking existing peer")
			recvdTimestampBytes := make([]byte, 4)
			_, err := io.ReadFull(conn, recvdTimestampBytes)
			handleErr(err, "While getting timestamp")
			recvdTimestamp := binary.BigEndian.Uint32(recvdTimestampBytes)
			peerManager.compareTimestampAndUpdate(conn, recvdTimestamp, senderIPString)
		}
	}
}
