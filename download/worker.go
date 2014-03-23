package download

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/patdowney/downloaderd/common"
)

const UPDATE_BYTE_DIFFERENCE = 50000

type DownloadHTTPError struct {
	URL        string
	Method     string
	StatusCode int
	Status     string
}

func (e DownloadHTTPError) Error() string {
	return fmt.Sprintf("%s failed for %s with code %d (%s)", e.Method, e.URL, e.StatusCode, e.Status)
}

type Worker struct {
	ID           uint
	Clock        common.Clock
	FileStore    FileStore
	WorkQueue    chan Download
	ErrorChannel chan DownloadError
	StatusSender StatusSender
	stop         bool
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
	e := DownloadError{DownloadID: id}
	e.Time = w.Clock.Now()
	e.OriginalError = err.Error()

	w.ErrorChannel <- e
}

func (w Worker) SaveWithStatus(download *Download) error {
	downloadHash, err := download.Hash()
	if err != nil {
		w.SendError(download.ID, err)
	}

	statusWriter := NewStatusWriter(download.ID, w.StatusSender, downloadHash, UPDATE_BYTE_DIFFERENCE)
	defer statusWriter.Close()

	outputWriter, err := w.FileStore.GetWriter(download)
	if err != nil {
		w.SendError(download.ID, err)
		return err
	}
	defer outputWriter.Close()

	statusWriter.SendStartUpdate()

	return w.Save(download, outputWriter, statusWriter)
}

func (w Worker) Save(download *Download, outputWriter io.Writer, statusWriter *StatusWriter) error {
	download.TimeStarted = time.Now()

	res, err := http.Get(download.URL)
	if err != nil {
		w.SendError(download.ID, err)
		return err
	}

	if res.StatusCode != http.StatusOK {
		err = DownloadHTTPError{
			URL:        download.URL,
			Method:     "Get",
			Status:     res.Status,
			StatusCode: res.StatusCode}

		w.SendError(download.ID, err)
		return err
	}

	fetchedBody := res.Body
	defer fetchedBody.Close()

	bufferedReader := bufio.NewReader(fetchedBody)
	err = w.WriteData(bufferedReader, outputWriter, statusWriter)
	if err != nil {
		w.SendError(download.ID, err)
	}

	return err
}

func NewWorker(id uint, workQueue chan Download, updateChannel chan StatusUpdate, errorChannel chan DownloadError, fileStore FileStore) *Worker {
	worker := &Worker{
		Clock:        &common.RealClock{},
		ID:           id,
		WorkQueue:    workQueue,
		StatusSender: &ChannelStatusSender{StatusChannel: updateChannel},
		ErrorChannel: errorChannel,
		FileStore:    fileStore}

	return worker
}
