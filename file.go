package main

import (
	"os"
	"crypto/md5"
	"io"
	"encoding/hex"
	"fmt"
)

type File struct {
	filePath   string
	fileSize   uint32
	transferredSize uint32
	handshake_complete bool
	md5 string
}

type MultipleFiles []File

func (files MultipleFiles) add(newFile File) MultipleFiles{
	files = append(files, newFile)
	fmt.Println("Added to slice ", files)
	return files
}

func (files MultipleFiles) updateAfterHandshake(md5 string) MultipleFiles{
	fmt.Println("Files received for update", files)
	for i:= range files {
		fmt.Println("Checking file", files[i])
		if files[i].md5 == md5 {
			files[i].handshake_complete = true
			break
		}
	}
	return files
}


func getMD5Hash(filePath string) (string, error) {
	var returnMD5String string
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}
	//Get the 16 bytes hash
	hashInBytes := hash.Sum(nil)[:16]
	//Convert the bytes to a string
	returnMD5String = hex.EncodeToString(hashInBytes)
	return returnMD5String, nil

}
