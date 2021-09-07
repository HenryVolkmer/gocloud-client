package main

import(
	"os/signal"
	"os"
	"syscall"
	"path/filepath"
	"flag"
	"log"
	"sync"
	"github.com/HenryVolkmer/libfilesync"
)

var (
	siteDir *string 
	wg sync.WaitGroup
)

func main() {

	// if no config was provided as flag, try to locate a config in users home
	siteDir := flag.String("site", "", "path to the configfile")
	threads := flag.Int("threads",5,"Number of threads")
	flag.Parse()
	log.Printf("sync dir %s",*siteDir)

	sigs := make(chan os.Signal, 1)
    done := make(chan bool, 1)
	queue := make(chan libfilesync.Syncable,*threads)
	workLoad := make([]libfilesync.Syncable,0)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
        sig := <-sigs
        log.Println(sig)
        done <- true
    }()

	for workerId := 0;workerId < *threads;workerId++ {
		wg.Add(1)
		go func (id int,wg *sync.WaitGroup,queue chan libfilesync.Syncable) {
			for file := range queue {
				procSyncableFile(file)
			}
			wg.Done()
		}(workerId,&wg,queue)
	}
	
	for {
		ReadDir(*siteDir,&workLoad)
		for _,file := range workLoad {
			select {
				case <-done:
					log.Println("Exiting, pleae wait ...")
					close(queue)
					wg.Wait()
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
func ReadDir(siteDir string,workLoad *[]libfilesync.Syncable)  {

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
		fileStruct,err := libfilesync.NewSyncableFile(file,abs)

		if err != nil {
			continue
		}

		*workLoad = append(*workLoad,fileStruct)
	}
}