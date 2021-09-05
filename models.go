package main

import (
	"os"
	"io"
	"errors"
	"encoding/hex"
	"crypto/sha1"
)

type SyncableFile struct {
	Name string
	Path string
	sha1 string
	modTime string

}

/**
 * Generate a File Struct by os.DirEntry
 */
func GetSyncableFile(entry os.DirEntry,absDir string) (SyncableFile,error) {

	finfo,err := entry.Info()
	if err != nil {
		panic("Could not get Fileinfo")
	}

	file, err := os.Open(absDir)
	defer file.Close()

	hash := sha1.New()

	if _, err := io.Copy(hash, file); err != nil {
		return SyncableFile{}, errors.New("Could not generate sha1 hash for file " + finfo.Name())
	}	
	sf := SyncableFile{}
	//Get the 20 bytes hash
	hashInBytes := hash.Sum(nil)[:20]
	sf.sha1 = hex.EncodeToString(hashInBytes)
	sf.Name = finfo.Name()
	sf.Path = absDir

	return sf,nil
}