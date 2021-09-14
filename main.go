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
	"github.com/fsnotify/fsnotify"
)

var (
	// siteDir is the local directory to keep in sync with
	// the server
	siteDir *string 
	wg sync.WaitGroup
)

func main() {

	siteDir := flag.String("site", "", "path to the dir to keep in sync with the server")
	threads := flag.Int("threads",5,"Number of threads")
	flag.Parse()

	sigs := make(chan os.Signal, 1)
    done := make(chan bool, 1)
	queue := make(chan libfilesync.Syncable,*threads)
	workLoad := make([]libfilesync.Syncable,0)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	// main loop
	// track file changes in siteDir and its subfolders
	go func() {
		for {
			select {

			case <-done:
				return

			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				abs := filepath.Join(*siteDir, event.Name)
		
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					log.Println("removed file:", event.Name)
					deleteFile(abs)
					return
				}

				// stat file and check if dir or file
				file, err := os.Open(abs)
				if (err != nil) {
					panic("could not open file for watching")
				}
				finfo,err := file.Stat()

				
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
				}
				if event.Op&fsnotify.Create == fsnotify.Create {

					log.Println("Created file:", event.Name)	
					if (finfo.IsDir()) {
						log.Println("added for watchcing:", event.Name)	
						watcher.Add(abs)
					} else {
						fileStruct,err := libfilesync.NewSyncableFile(abs,libfilesync.CHECK)
						if err != nil {
							panic("could not create fileStruct by FileInfo")
						}
						queue<-fileStruct
					}
				}
				if event.Op&fsnotify.Rename == fsnotify.Rename {
					log.Println("Renamed file:", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	// listener for signals
	go func() {
        sig := <-sigs
        log.Println(sig)
        done <- true
    }()

    // workers syncing Files with Server
	for workerId := 0;workerId < *threads;workerId++ {
		wg.Add(1)
		go func () {
			for file := range queue {
				procSyncableFile(file)
			}
			wg.Done()
		}()
	}
	
	// scan siteDir on startup
	// this subscribes all Directorys recursive
	// to fsnotify and put all Files to workLoad
	ReadDir(*siteDir,&workLoad,watcher)

	// enqueue all Files
	// the workers will check for modifications
	// an performs the sync with the Server
	for _,file := range workLoad {
		select {
			case <-done:
				log.Println("Exiting, waiting for workers to finish ...")
				close(queue)
				wg.Wait()
				//os.Exit(0)
				break
			default:
				queue<-file		
		}
	}

	// run till done
	<-done
	os.Exit(0)
}

/**
 * parse working dir recursive and push files to []string
 * Directorys will be subscribed to fsnotify
 */
func ReadDir(siteDir string,workLoad *[]libfilesync.Syncable,watcher *fsnotify.Watcher)  {

	dir, err := os.ReadDir(siteDir)	
	if err != nil {
		log.Printf("Dir not statable: %s",dir)
		panic("Could not read dir")
	}

	watcher.Add(siteDir)

	for _,file := range dir {

		abs := filepath.Join(siteDir, file.Name())
		if file.IsDir() == true {
			ReadDir(abs, workLoad, watcher)
			continue
		}
		fileStruct,err := libfilesync.NewSyncableFile(abs,libfilesync.CHECK)

		if err != nil {
			continue
		}

		*workLoad = append(*workLoad,fileStruct)
	}
}