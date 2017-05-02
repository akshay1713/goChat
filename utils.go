package main

import "fmt"

func padLeft(str, pad string, length int) string {
	if len(str) >= length {
		return str
	}
	for {
		str = pad + str
		if len(str) > length {
			return str[0:length]
		}
	}
}

func handleErr(err error, prefix string) {
	if err != nil{
		fmt.Println(prefix,": ", err)
	}
}