package main

import (
	"flag"
	"github.com/patdowney/downloaderd/api"
	"github.com/patdowney/downloaderd/download"
	dh "github.com/patdowney/downloaderd/http"
	"github.com/patdowney/downloaderd/local"
	"io"
	"log"
	"os"
)

type Config struct {
	ListenAddress     string
	WorkerCount       uint
	QueueLength       uint
	DownloadDirectory string
	DownloadDataFile  string
	RequestDataFile   string
	HookDataFile      string

	AccessLogWriter io.Writer
	ErrorLogWriter  io.Writer
}

func ConfigureLogging(config *Config) {
	log.SetOutput(config.ErrorLogWriter)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func ParseArgs() *Config {
	c := &Config{}
	flag.StringVar(&c.ListenAddress, "http", "localhost:8080", "address to listen on")
	flag.UintVar(&c.WorkerCount, "workers", 2, "number of workers to use")
	flag.UintVar(&c.QueueLength, "queuelength", 32, "size of download queue")
	flag.StringVar(&c.DownloadDirectory, "downloaddir", "./download-data", "root directory of save tree.")
	flag.StringVar(&c.DownloadDataFile, "downloaddata", "downloads.json", "download database file")
	flag.StringVar(&c.RequestDataFile, "requestdata", "requests.json", "request database file")
	flag.StringVar(&c.HookDataFile, "hookdata", "hooks.json", "hooks database file")
	flag.Parse()

	c.AccessLogWriter = os.Stdout
	c.ErrorLogWriter = os.Stderr

	return c
}

func CreateServer(config *Config) {
	s := dh.NewServer(&dh.HTTPConfig{ListenAddress: config.ListenAddress})

	downloadStore, err := local.NewDownloadStore(config.DownloadDataFile)
	if err != nil {
		log.Printf("init-download-store-error: %v", err)
	}

	fileStore := local.NewFileStore(config.DownloadDirectory)

	requestStore, err := local.NewRequestStore(config.RequestDataFile)
	if err != nil {
		log.Printf("init-request-store-error: %v", err)
	}

	hookStore, err := local.NewHookStore(config.HookDataFile)
	if err != nil {
		log.Printf("init-hook-store-error: %v", err)
	}

	linkResolver := api.NewLinkResolver(s.Router)
	linkResolver.DefaultScheme = "http"
	linkResolver.DefaultHost = config.ListenAddress

	downloadService := download.NewDownloadService(downloadStore, fileStore, config.WorkerCount, config.QueueLength)
	downloadService.HookService = download.NewHookService(hookStore, linkResolver)

	requestService := download.NewRequestService(requestStore, downloadService)

	downloadResource := dh.NewDownloadResource(downloadService, linkResolver)
	s.AddResource("/download", downloadResource)

	requestResource := dh.NewRequestResource(requestService, linkResolver)
	s.AddResource("/request", requestResource)

	downloadService.Start()

	err = s.ListenAndServe()
	log.Printf("init-listen-error: %v", err)
}

func main() {
	config := ParseArgs()

	ConfigureLogging(config)

	CreateServer(config)
}
