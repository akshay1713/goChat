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

func getBytesFromUint32(source []byte, num uint32) {
	source[0] = byte(num >> 24)
	source[1] = byte(num >> 16)
	source[2] = byte(num >> 8)
	source[3] = byte(num)
}

