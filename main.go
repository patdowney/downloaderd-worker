package main

import (
	"flag"
	"github.com/patdowney/downloaderd/download"
	dh "github.com/patdowney/downloaderd/http"
	"github.com/patdowney/downloaderd/local"
	"io"
	"log"
	"os"
)

type Config struct {
	ListenAddress     string
	WorkerCount       int
	QueueLength       int
	DownloadDirectory string
	DownloadDataFile  string
	RequestDataFile   string

	AccessLogWriter io.Writer
	ErrorLogWriter  io.Writer
}

func ConfigureLogging(config *Config) {
	log.SetOutput(config.ErrorLogWriter)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func ParseArgs() *Config {
	c := &Config{}
	flag.StringVar(&c.ListenAddress, "http", ":8080", "address to listen on")
	flag.IntVar(&c.WorkerCount, "workers", 2, "number of workers to use")
	flag.IntVar(&c.QueueLength, "queuelength", 32, "size of download queue")
	flag.StringVar(&c.DownloadDirectory, "downloaddir", ".", "root directory of save tree.")
	flag.StringVar(&c.DownloadDataFile, "downloaddata", "downloads.json", "download database file")
	flag.StringVar(&c.RequestDataFile, "requestdata", "requests.json", "request database file")
	flag.Parse()

	c.AccessLogWriter = os.Stdout
	c.ErrorLogWriter = os.Stderr

	return c
}

func CreateServer(config *Config) {
	s := dh.NewServer(&dh.HTTPConfig{ListenAddress: config.ListenAddress})

	downloadStore, err := local.NewDownloadStore(config.DownloadDataFile)
	downloadService := download.NewDownloadService(downloadStore)
	downloadResource := dh.NewDownloadResource(downloadService)
	s.AddResource("/download", downloadResource)

	requestStore, err := local.NewRequestStore(config.RequestDataFile)
	requestService := download.NewRequestService(requestStore, downloadService)
	requestResource := dh.NewRequestResource(requestService)

	s.AddResource("/request", requestResource)

	err = s.ListenAndServe()
	log.Print(err)
}

func main() {
	config := ParseArgs()

	ConfigureLogging(config)

	CreateServer(config)
}
