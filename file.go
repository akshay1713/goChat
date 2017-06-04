package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type File struct {
	filePath           string
	fileSize           uint64
	transferredSize    uint64
	handshake_complete bool
	md5                string
	filePtr            *os.File
	uniqueID           uint32
}

func (file File) getFileName() string {
	return filepath.Base(file.filePath)
}

func (file *File) getNextBytes() []byte {
	remainingSize := int(file.fileSize - file.transferredSize)
	if remainingSize == 0 {
		return []byte{}
	}
	bytesToTransfer := 4096
	if remainingSize < 4096 {
		bytesToTransfer = remainingSize
		defer file.filePtr.Close()
		fmt.Println("Finished sending file")
	}
	nextBytes := make([]byte, int(bytesToTransfer))
	file.transferredSize += uint64(bytesToTransfer)
	file.filePtr.Read(nextBytes)
	return nextBytes
}

func (file *File) writeBytes(nextBytes []byte) {
	file.filePtr.Write(nextBytes)
	file.transferredSize += uint64(len(nextBytes))
	if file.transferredSize == file.fileSize {
		file.filePtr.Close()
		fmt.Println("Finished receiving file")
	}
}

type MultipleFiles []File

func (files MultipleFiles) add(newFile File) MultipleFiles {
	files = append(files, newFile)
	fmt.Println("Added to slice ", files)
	return files
}

func (files MultipleFiles) updateAfterHandshake(md5 string) MultipleFiles {
	fmt.Println("Files received for update", files)
	for i := range files {
		fmt.Println("Checking file", files[i])
		if files[i].md5 == md5 {
			files[i].handshake_complete = true
			files[i].filePtr, _ = os.Open(files[i].filePath)
			break
		}
	}
	return files
}

func (files MultipleFiles) get(uniqueID uint32) File {
	for i := range files {
		if files[i].uniqueID == uniqueID {
			return files[i]
		}
	}
	fmt.Println("File with uniqueID ", uniqueID, "not found")
	return File{}
}

func (files MultipleFiles) update(file File) MultipleFiles {
	for i := range files {
		if files[i].md5 == file.md5 {
			files[i] = file
			return files
		}
	}
	fmt.Println("No matching file found, unchanged")
	return files
}

func (files MultipleFiles) openForReading(md5 string) MultipleFiles {
	fmt.Println("Opening for reading ")
	for i := range files {
		if files[i].md5 == md5 {
			files[i].filePtr, _ = os.Open(files[i].filePath)
		}
	}
	return files
}

func (files MultipleFiles) openForWriting(md5 string) MultipleFiles {
	fmt.Println("Opening for writing")
	for i := range files {
		if files[i].md5 == md5 {
			files[i].filePtr, _ = os.OpenFile(files[i].filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		}
	}
	return files
}

func (files MultipleFiles) getOpenPointer(md5 string) (*os.File, error) {
	var reqFile File
	for i := range files {
		if files[i].md5 == md5 {
			reqFile = files[i]
			break
		}
	}
	file, err := os.Open(reqFile.filePath)
	return file, err
}

func newFile(filePath string) (File, error) {
	md5, err := getMD5Hash(filePath)
	file := File{
		filePath:           filePath,
		fileSize:           getFileSize(filePath),
		transferredSize:    0,
		handshake_complete: false,
		md5:                md5,
		uniqueID:           uint32(time.Now().UTC().Unix()),
	}
	return file, err
}

func getFileSize(filePath string) uint64 {
	filePtr, _ := os.Open(filePath)
	defer filePtr.Close()
	fileStats, _ := filePtr.Stat()
	return uint64(fileStats.Size())
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
