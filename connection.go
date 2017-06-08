package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

type IPeerManager interface {
	getAllIPs() []string
	isConnected(IP string) bool
}

type UDPBroadcaster struct {
	ports   []string
	appName string
}

func (udpBroadcaster UDPBroadcaster) init(manager IPeerManager) *net.UDPConn  {
	var serverConn *net.UDPConn
	for i := range udpBroadcaster.ports {
		serverAddr, err := net.ResolveUDPAddr("udp", ":"+udpBroadcaster.ports[i])
		if err != nil {
			log.Println("Error while resolving address ", err)
		}
		if serverConn == nil {
			serverConn, err = net.ListenUDP("udp", serverAddr)
			if err != nil {
				log.Println("Error while listening to address", err)
			}
		}
	}
	if serverConn == nil {
		panic("Unable to listen for UDP on any of the ports")
	}
	return serverConn
}

func (udpBroadcaster UDPBroadcaster) broadcastOnAllPorts(tcpListenerPort string) []string {
	ports := udpBroadcaster.ports
	var localAddrs []string
	for i := range ports {
		serverAddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:"+ports[i])
		if err != nil {
			panic(err)
		}
		udpConn, err := net.DialUDP("udp", nil, serverAddr)
		if err != nil {
			panic(err)
		}
		go udpBroadcaster.broadcastOnSinglePort(udpConn, tcpListenerPort)
		localAddrs = append(localAddrs, udpConn.LocalAddr().String())
	}
	return localAddrs
}

func (udpBroadcaster UDPBroadcaster) broadcastOnSinglePort(conn *net.UDPConn, port string) {
	defer conn.Close()
	var msg []byte
	msg = append(msg, udpBroadcaster.appName...)
	port = padLeft(port, "0", 5)
	msg = append(msg, port...)
	buf := []byte(msg)
	for i := 0; i < 5; i++ {
		_, err := conn.Write(buf)
		if err != nil {
			log.Println("Error while broadcasting:", err)
			time.Sleep(time.Second * 1)
		}
	}
}

type UDPListener struct {
	listenerPort int
	listenerConn *net.UDPConn
	possibleLocalAddrs []string
	appName string
}

func (udpListener UDPListener) isMessageValid(addr *net.UDPAddr, msg []byte) bool{
	possibleLocalAddrs := udpListener.possibleLocalAddrs
	appName := udpListener.appName
	if pos(possibleLocalAddrs, addr.IP.String()+":"+strconv.Itoa(addr.Port)) != -1 {
		return false
	}

	if string(msg[0:len(appName)]) != appName {
		return false
	}
	return true
}

func (udpListener UDPListener) listenForUDPBroadcast(peerManager IPeerManager) {

	ServerConn := udpListener.listenerConn
	port := udpListener.listenerPort
	defer ServerConn.Close()
	appName := udpListener.appName
	portLen := 5
	typeLen := 0
	buf := make([]byte, len(appName)+portLen+typeLen)
	broadcastRecvdIPs := make(map[string]bool)
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(port))
	for {
		_, addr, err := ServerConn.ReadFromUDP(buf)
		if !udpListener.isMessageValid(addr, buf) {
			continue
		}
		recvdPort, err := strconv.Atoi(string(buf[len(appName):]))
		if err != nil {
			fmt.Println("Error: ", err)
			continue
		}
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
			//fmt.Println(peerIPs)
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

type ConnAndType struct{
	Connection *net.TCPConn
	Type       string
}

func waitForTCP(peerManager PeerManager, listener net.Listener, initiatorConn chan ConnAndType) {
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
				connAndType := ConnAndType{Connection:newConn, Type:"sender"}
				initiatorConn <- connAndType
				for k := 2; k < len(peerInfo); k += 6 {
					peerIP := net.IPv4(peerInfo[k+2], peerInfo[k+3], peerInfo[k+4], peerInfo[k+5])
					peerPort := binary.BigEndian.Uint16([]byte{peerInfo[k], peerInfo[k+1]})
					if !peerManager.isConnected(peerIP.String()) {
						newConn, err = connectToPeer(peerIP, int(peerPort))
						handleErr(err, "Error while connecting to peer")
						initiatorConn <- ConnAndType{Connection:newConn, Type:"sender"}
					}
				}
			} else {
				initiatorConn <- ConnAndType{Connection:conn, Type:"receiver"}
			}
		} else if msgType[0] == 1 {
			fmt.Println("Checking existing peer")
			initiatorConn <- ConnAndType{Connection:conn, Type:"duplicate_receiver"}
		}
	}
}
