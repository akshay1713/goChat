package main

import (
	//"bytes"
	"encoding/binary"
	"fmt"
	//"io"
	"io"
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
			time.Sleep(time.Second * 3)
		}
	}()
	return LocalAddr
}

func listenForUDPBroadcast(ServerConn *net.UDPConn, LocalAddr net.Addr, peerManager *PeerManager, port int) {
	defer ServerConn.Close()
	appName := "goChat"
	portLen := 5
	buf := make([]byte, len(appName)+portLen)
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(port))
	fmt.Println(portBytes)
	var all_ips []byte
	fmt.Println("Only port ", all_ips, portBytes)
	broadcastRecvdIPs := make(map[string]bool)
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
		if _, exists := broadcastRecvdIPs[addr.IP.String()]; !exists {
			broadcastRecvdIPs[addr.IP.String()] = true
			peerIPs := peerManager.getAllIPs()
			all_ips = []byte{}
			totalLen := 2 + len(peerIPs)*6
			msgLengthBytes := make([]byte, 4)
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
			//newPeer := peerManager.addNewPeer(newConnection)
			//go newPeer.setPing()
			sConn.Close()
			fmt.Println("New peer joined", sConn.RemoteAddr().String())
		}
	}
}

func connectToPeer(ip net.IP, port int, timestamp uint32) (*net.TCPConn, error) {
	fmt.Println("Connecting to ", ip, port)
	tcpAddr := net.TCPAddr{IP: ip, Port: port}
	chatConn, err := net.DialTCP("tcp", nil, &tcpAddr)
	timestampBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(timestampBytes, timestamp)
	chatConn.Write(timestampBytes)
	return chatConn, err
}

func waitForTCP(peerManager *PeerManager, listener net.Listener) {
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
		fmt.Println("Sender IP is ", senderIP)
		msgType := make([]byte, 1)
		fmt.Println("Recieved new connection", senderIPString, senderIP, msgType)
		_, err = io.ReadFull(conn, msgType)
		if !peerManager.isConnected(senderIPString) {
			if msgType[0] == 0 {
				//This msg is a list of IPs & ports
				msgLength := make([]byte, 4)
				_, err = io.ReadFull(conn, msgLength)
				handleErr(err, "Error while reading message ")
				peerInfoLength := binary.BigEndian.Uint32(msgLength)
				fmt.Println("Msg length ", msgLength, peerInfoLength)
				peerInfo := make([]byte, peerInfoLength)
				_, err = io.ReadFull(conn, peerInfo)
				handleErr(err, "Error while reading message ")
				senderPort := binary.BigEndian.Uint16([]byte{peerInfo[0], peerInfo[1]})
				fmt.Println("Connecting to ", senderIPString, senderPort)
				//splitIP := strings.Split(senderIPString, ".")
				currentTimestamp := uint32(time.Now().UTC().Unix())
				newConn, err := connectToPeer(senderIP, int(senderPort), currentTimestamp)
				handleErr(err, "Error while connecting to sender")
				if newConn == nil {
					fmt.Println("Nil conn")
					continue
				}
				newPeer := peerManager.addNewPeer(newConn, currentTimestamp)
				newPeer.setPing()
				handleErr(err, "Error while connecting to sender")
				for k := 2; k < len(peerInfo); k += 6 {
					peerIP := net.IPv4(peerInfo[k+2], peerInfo[k+3], peerInfo[k+4], peerInfo[k+5])
					peerPort := binary.BigEndian.Uint16([]byte{peerInfo[k], peerInfo[k+1]})
					currentTimestamp = uint32(time.Now().UTC().Unix())
					newConn, err = connectToPeer(peerIP, int(peerPort), currentTimestamp)
					if !peerManager.isConnected(peerIP.String()){
						currentTimestamp := uint32(time.Now().UTC().Unix())
						newPeer = peerManager.addNewPeer(newConn, currentTimestamp)
					}
				}
			} else {
				//This msg is a connection request
				fmt.Println("Processing connection request")
				msgLength := 4
				recvdTimestampBytes := make([]byte, msgLength)
				_, err = io.ReadFull(conn, recvdTimestampBytes)
				recvdTimestamp := binary.BigEndian.Uint32(recvdTimestampBytes)
				peerManager.addNewPeer(conn, recvdTimestamp)

			}
		} else {
			fmt.Println("Processing alternative check")
			msgLength := 4
			recvdTimestampBytes := make([]byte, msgLength)
			_, err = io.ReadFull(conn, recvdTimestampBytes)
			recvdTimestamp := binary.BigEndian.Uint32(recvdTimestampBytes)
			existingPeer := peerManager.getPeer(senderIPString)
			if recvdTimestamp < existingPeer.connectedAt {
				fmt.Println("New is old")
			}
			peerManager.updatePeer(senderIPString, conn, recvdTimestamp)
		}
	}
}
