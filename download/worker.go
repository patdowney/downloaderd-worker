package download

import (
	"bufio"
	//"encoding/hex"
	"fmt"
	//"hash"
	"io"
	//"log"
	"net/http"
	//"os"
	//"path/filepath"
	"github.com/patdowney/downloaderd/common"
	"time"
)

type DownloadHTTPError struct {
	Url        string
	Method     string
	StatusCode int
	Status     string
}

func (e DownloadHTTPError) Error() string {
	return fmt.Sprintf("%s failed for %s with code %d (%s)", e.Method, e.Url, e.StatusCode, e.Status)
}

type Worker struct {
	Id            uint
	Clock         common.Clock
	FileStore     FileStore
	WorkQueue     chan Download
	ErrorChannel  chan DownloadError
	UpdateChannel chan StatusUpdate
	stop          bool
}

func (w Worker) start() {
	go func() {
		for {
			download := <-w.WorkQueue
			w.SaveWithStatus(&download)
		}
	}()
}

func (w Worker) WriteData(dataReader io.Reader, outputWriter io.Writer, statusWriter *StatusWriter) error {
	teeReader := io.TeeReader(dataReader, statusWriter)

	_, err := io.Copy(outputWriter, teeReader)
	if err != nil {
		return err
	}

	return nil
}

func (w Worker) SendError(id string, err error) {
	e := DownloadError{DownloadId: id}
	e.Time = w.Clock.Now()
	e.OriginalError = err

	w.ErrorChannel <- e
}

func (w Worker) SaveWithStatus(download *Download) error {
	downloadHash, err := download.Hash()
	if err != nil {
		w.SendError(download.Id, err)
	}

	statusWriter := NewStatusWriter(download.Id, w.UpdateChannel, downloadHash)
	defer statusWriter.Close()

	outputWriter, err := w.FileStore.GetWriter(download)
	defer outputWriter.Close()
	if err != nil {
		w.SendError(download.Id, err)
		return err
	}

	return w.Save(download, outputWriter, statusWriter)
}

func (w Worker) Save(download *Download, outputWriter io.Writer, statusWriter *StatusWriter) error {
	download.TimeStarted = time.Now()

	res, err := http.Get(download.Url)
	if err != nil {
		w.SendError(download.Id, err)
		return err
	}

	if res.StatusCode != http.StatusOK {
		err = DownloadHTTPError{
			Url:        download.Url,
			Method:     "Get",
			Status:     res.Status,
			StatusCode: res.StatusCode}

		w.SendError(download.Id, err)
		return err
	}

	fetchedBody := res.Body
	defer fetchedBody.Close()

	bufferedReader := bufio.NewReader(fetchedBody)
	err = w.WriteData(bufferedReader, outputWriter, statusWriter)
	if err != nil {
		w.SendError(download.Id, err)
	}

	return err
}

func NewWorker(id uint, workQueue chan Download, updateChannel chan StatusUpdate, errorChannel chan DownloadError, fileStore FileStore) *Worker {
	worker := &Worker{
		Clock:         &common.RealClock{},
		Id:            id,
		WorkQueue:     workQueue,
		UpdateChannel: updateChannel,
		ErrorChannel:  errorChannel,
		FileStore:     fileStore}

	return worker
}
