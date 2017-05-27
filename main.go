package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"flag"
	"context"
	"golang.org/x/oauth2"
	"github.com/google/go-github/github"
	"github.com/pelletier/go-toml"
)

func main() {
	chatType := getChatType()
	username := getUserName(chatType)
	if username == "" {
		fmt.Println("Please specify a username or chat as your github avatar")
		return
	}
	fmt.Println("Joining as ", username)
	ServerAddr, err := net.ResolveUDPAddr("udp", ":7041")
	if err != nil {
		fmt.Println("Err while resolving IP address", err)
	}
	peerConnections := make(map[string]Peer)
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		fmt.Println("Err while listening for connections", err)
		return
	}
	closeChan := make(chan Peer)
	peerManager := PeerManager{closeChan: closeChan, connectedPeers: peerConnections}
	go peerManager.init()
	go waitForTCP(peerManager, l, username)
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	ListenerAddr := l.Addr()
	LocalAddr := initUDPBroadcast(ListenerAddr, peerConnections, byte(0))

	if err != nil {
		fmt.Println("Err while listening to the address", err)
	}
	tcpListenerAddr := strings.Split(l.Addr().String(), ":")
	port, _ := strconv.Atoi(tcpListenerAddr[len(tcpListenerAddr)-1])
	go listenForUDPBroadcast(ServerConn, LocalAddr, peerManager, port, byte(0))
	startCli(peerManager)
}

func getChatType() byte {
	gitPtr := flag.Bool("g", false, "Use your github username & connect " +
		"only to the collaborators of the current repository")
	if *gitPtr {
		return 1
	}
	return 0
}

func getUserName(chatType byte) string {
	var usernamePtr string
	if chatType == 0 {
		fmt.Print("Enter the username you wish to use for this chat session: ")
		fmt.Scanln(&usernamePtr)
		return  usernamePtr
	}
	config, err := toml.LoadFile("gochat.toml")
	if err != nil {
		fmt.Println("Error while trying to load config file: ",err)
		return ""
	}
	auth_token := config.Get("auth_token").(string)
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: auth_token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	user, _, _ := client.Users.Get("")
	usernamePtr = *user.Login
	return usernamePtr
}
