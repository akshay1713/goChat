package main

import (
	"bufio"
	"github.com/abiosoft/ishell"
	"os"
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
				peerManager.sendMessage(msg)
			}
		},
	})
	shell.Start()
}
