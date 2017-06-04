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

//getNextBytes gets the next bytes to transfer to a peer, for a file which is currently being transferred to a peer
//Returns data in chunks of 4096 bytes.
//Required an open pointer to the file to exist
func (file *File) getNextBytes() []byte {
	remainingSize := int(file.fileSize - file.transferredSize)
	if remainingSize == 0 {
		return []byte{}
	}
	//Transfer in chunks of 4096 bytes
	bytesToTransfer := 4096
	if remainingSize < 4096 {
		bytesToTransfer = remainingSize
		defer file.filePtr.Close()
		fmt.Println("Finished sending file", file.getFileName())
	}
	nextBytes := make([]byte, int(bytesToTransfer))
	file.transferredSize += uint64(bytesToTransfer)
	file.filePtr.Read(nextBytes)
	return nextBytes
}

//writeBytes writes the given bytes to a file which is already open, and is being transferred by a peer.
//Requires an open pointer to the file to exist
func (file *File) writeBytes(nextBytes []byte) {
	file.filePtr.Write(nextBytes)
	file.transferredSize += uint64(len(nextBytes))
	if file.transferredSize == file.fileSize {
		file.filePtr.Close()
		fmt.Println("Finished receiving file", file.getFileName())
	}
}

type MultipleFiles []File

func (files MultipleFiles) add(newFile File) MultipleFiles {
	files = append(files, newFile)
	return files
}

//updateAfterHandshake finds an updates the specific file object after the potential file receiver sends a file
//acceptance message. Also opens a file pointer to the file, to make it ready for transfer.
func (files MultipleFiles) updateAfterHandshake(md5 string) MultipleFiles {
	for i := range files {
		if files[i].md5 == md5 {
			files[i].handshake_complete = true
			files[i].filePtr, _ = os.Open(files[i].filePath)
			break
		}
	}
	return files
}

//get accepts a unique id associated with a file, with a peer, and returns that file object if found.
//If not found, an empty File object is returned
func (files MultipleFiles) get(uniqueID uint32) File {
	for i := range files {
		if files[i].uniqueID == uniqueID {
			return files[i]
		}
	}
	fmt.Println("File with uniqueID ", uniqueID, "not found")
	return File{}
}

//update updates the structure containing multiple files, by replacing an existing file object with the given File
//object, and uses the md5 to find the File object which is supposed to be replaced
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
	for i := range files {
		if files[i].md5 == md5 {
			files[i].filePtr, _ = os.Open(files[i].filePath)
		}
	}
	return files
}

func (files MultipleFiles) openForWriting(md5 string) MultipleFiles {
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
