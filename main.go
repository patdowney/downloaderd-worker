package main

import (
	"flag"
	"io"
	"log"
	"os"

	http "github.com/patdowney/downloaderd-common/http"
	"github.com/patdowney/downloaderd-worker/api"
	"github.com/patdowney/downloaderd-worker/download"
	dh "github.com/patdowney/downloaderd-worker/http"
	"github.com/patdowney/downloaderd-worker/local"
	//"github.com/patdowney/downloaderd-common/rethinkdb"
	//"github.com/patdowney/downloaderd-worker/rethinkdb"
	//"github.com/patdowney/downloaderd-worker/s3"
)

// Config ...
type Config struct {
	ListenAddress     string
	WorkerCount       uint
	QueueLength       uint
	DownloadDirectory string
	DownloadDataFile  string
	HookDataFile      string

	AccessLogWriter io.Writer
	ErrorLogWriter  io.Writer

	RethinkDBAddress string
}

// ConfigureLogging ...
func ConfigureLogging(config *Config) {
	log.SetOutput(config.ErrorLogWriter)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

// ParseArgs ...
func ParseArgs() *Config {
	c := &Config{}
	flag.StringVar(&c.ListenAddress, "http", "localhost:8080", "address to listen on")
	flag.UintVar(&c.WorkerCount, "workers", 2, "number of workers to use")
	flag.UintVar(&c.QueueLength, "queuelength", 32, "size of download queue")
	flag.StringVar(&c.RethinkDBAddress, "rethinkdb", "localhost:28015", "address to listen on")
	flag.StringVar(&c.DownloadDirectory, "downloaddir", "./download-data", "root directory of save tree.")
	flag.StringVar(&c.DownloadDataFile, "downloaddata", "downloads.json", "download database file")
	flag.StringVar(&c.HookDataFile, "hookdata", "hooks.json", "hooks database file")
	flag.Parse()

	c.AccessLogWriter = os.Stdout
	c.ErrorLogWriter = os.Stderr

	return c
}

// CreateServer ...
func CreateServer(config *Config) {
	s := http.NewServer(&http.Config{ListenAddress: config.ListenAddress}, os.Stdout)

	downloadStore, err := local.NewDownloadStore(config.DownloadDataFile)
	/*
		c := rethinkdb.Config{Address: config.RethinkDBAddress,
			MaxIdle:  10,
			MaxOpen:  20,
			Database: "Downloaderd"}

		downloadStore, err := rethinkdb.NewDownloadStore(c)
	*/
	if err != nil {
		log.Printf("init-download-store-error: %v", err)
	}

	fileStore := local.NewFileStore(config.DownloadDirectory)
	//c3 := s3.Config{BucketName: "downloaderd", RegionName: "us-east-1"}
	//fileStore, err := s3.NewFileStore(c3)
	//if err != nil {
	//		log.Printf("s3-init-filestore-error: %v", err)
	//	}

	hookStore, err := local.NewHookStore(config.HookDataFile)
	//hookStore, err := rethinkdb.NewHookStore(c)
	if err != nil {
		log.Printf("init-hook-store-error: %v", err)
	}

	linkResolver := api.NewLinkResolver(s.Router)
	linkResolver.DefaultScheme = "http"
	linkResolver.DefaultHost = config.ListenAddress

	downloadService := download.NewDownloadService(downloadStore, fileStore, config.WorkerCount, config.QueueLength)
	downloadService.HookService = download.NewHookService(hookStore, linkResolver)

	downloadResource := dh.NewDownloadResource(downloadService, linkResolver)
	s.AddResource("/download", downloadResource)

	downloadService.Start()

	err = s.ListenAndServe()
	log.Printf("init-listen-error: %v", err)
}

func main() {
	config := ParseArgs()

	ConfigureLogging(config)

	CreateServer(config)
}
