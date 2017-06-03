package main

import (
	"bufio"
	"github.com/abiosoft/ishell"
	"os"
	"fmt"
)

func startCli(peerManager PeerManager) {
	shell := ishell.New()
	shell.Println("Started goChat")
	shell.AddCmd(&ishell.Cmd{
		Name: "chat",
		Help: "Start chatting",
		Func: func(c *ishell.Context) {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				msg := scanner.Text()
				if msg == "stop_chat" {
					break
				}
				peerManager.sendMessage(msg)
			}
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "send_file",
		Help: "Send a file to a peer",
		Func: func (c*ishell.Context) {
			var filePath string
			fmt.Print("Enter the file path you wish to send: ")
			fmt.Scanln(&filePath)
			fmt.Println("sending file ", filePath)
			connectedPeers := peerManager.getAllPeers()
			fmt.Println("Connected peers are: ")
			for i := range connectedPeers {
				fmt.Println(i,". ",connectedPeers[i].username)
			}
			var peerIndex int
			fmt.Print("Enter the index of the peer to whom you wish to send the file: ")
			fmt.Scan(&peerIndex)
			if peerIndex > len(connectedPeers) || peerIndex < 0 {
				fmt.Println("Enter a valid index")
				return
			}
			//Add support for multiple target peers later
			targeUsernames := []string{connectedPeers[peerIndex].getIPWithoutPort()}
			peerManager.sendFiles(targeUsernames, filePath)
		},
	})
	shell.Start()
}
