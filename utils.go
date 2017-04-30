package main

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
