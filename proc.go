package main

import (
	"log"
	"github.com/HenryVolkmer/libfilesync"
)

func procSyncableFile(file libfilesync.Syncable) {
	log.Printf("type: %s",file.GetProcType())
	log.Printf("proc %+v\n",file)
}

// commits delete to the server
func deleteFile(path string) {
	log.Printf("delete %s",path)
}

