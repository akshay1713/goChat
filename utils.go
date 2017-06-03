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
	if err != nil {
		fmt.Println(prefix, ": ", err)
	}
}

func getBytesFromUint32(source []byte, num uint32) {
	source[0] = byte(num >> 24)
	source[1] = byte(num >> 16)
	source[2] = byte(num >> 8)
	source[3] = byte(num)
}

func getBytesFromUint64(source []byte, num uint64) {
	source[0] = byte(num >> 56)
	source[1] = byte(num >> 48)
	source[2] = byte(num >> 40)
	source[3] = byte(num >> 32)
	source[4] = byte(num >> 24)
	source[5] = byte(num >> 16)
	source[6] = byte(num >> 8)
	source[7] = byte(num)
}

func pos(slice []string, value string) int {
	for p, v := range slice {
		if v == value {
			return p
		}
	}
	return -1
}
