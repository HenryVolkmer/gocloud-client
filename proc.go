package main

import (
	"log"
	"github.com/HenryVolkmer/libfilesync"
)

func procSyncableFile(file libfilesync.Syncable) {
	log.Printf("proc %+v\n",file)
}

