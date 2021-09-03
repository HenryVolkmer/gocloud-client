package main

import(
	"errors"
	"os/signal"
	"os"
	"syscall"
	"io"
	"path/filepath"
	"encoding/hex"
	"crypto/sha1"
	"flag"
	"log"
	"sync"
	"time"
)

var (
	siteDir *string 
	wg sync.WaitGroup
)

type SyncFile struct {
	Name string
	Path string
	sha1 string
}

/**
 * Generate a File Struct by os.DirEntry
 */
func GetFileStruct(entry os.DirEntry,absDir string) (SyncFile,error) {

	finfo,err := entry.Info()
	if err != nil {
		panic("Could not get Fileinfo")
	}

	file, err := os.Open(absDir)
	defer file.Close()

	hash := sha1.New()

	if _, err := io.Copy(hash, file); err != nil {
		return SyncFile{}, errors.New("Could not generate sha1 hash for file " + finfo.Name())
	}	
	sf := SyncFile{}
	//Get the 20 bytes hash
	hashInBytes := hash.Sum(nil)[:20]
	sf.sha1 = hex.EncodeToString(hashInBytes)
	sf.Name = finfo.Name()
	sf.Path = absDir

	return sf,nil
}

func main() {

	siteDir = flag.String("dir", "", "the directory to sync with server")
	threads := flag.Int("threads", 5, "the number of threads to use")
	flag.Parse()
	log.Printf("sync dir %s",*siteDir)

	sigs := make(chan os.Signal, 1)
    done := make(chan bool, 1)
	queue := make(chan SyncFile,*threads)
	workLoad := make([]SyncFile,0)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
        sig := <-sigs
        log.Println(sig)
        done <- true
    }()

	for workerId := 0;workerId < *threads;workerId++ {
		wg.Add(1)
		log.Printf("go workerId %d",workerId)
		go SyncWorker(workerId,&wg,queue)
	}
	
	for {
		ReadDir(*siteDir,&workLoad)
		for _,file := range workLoad {
			select {
				case <-done:
					os.Exit(0)
				default:
					queue<-file		
			}
		}
	}
}

/**
 * parse working dir recursive and push files to []string
 */
func ReadDir(siteDir string,workLoad *[]SyncFile)  {

	dir, err := os.ReadDir(siteDir)
	if err != nil {
		log.Printf("Dir not statable: %s",dir)
		panic("Could not read dir")
	}

	for _,file := range dir {

		abs := filepath.Join(siteDir, file.Name())
		if file.IsDir() == true {
			ReadDir(abs, workLoad)
			continue
		}
		fileStruct,err := GetFileStruct(file,abs)

		if err != nil {
			continue
		}

		*workLoad = append(*workLoad,fileStruct)
	}
}

func SyncWorker(id int,wg *sync.WaitGroup,queue chan SyncFile) {
	for file := range queue {
		log.Printf("Worker ID %d proc %+v\n",id,file)
		time.Sleep(time.Second*3)
	}
}