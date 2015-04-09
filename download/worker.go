package download

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/patdowney/downloaderd/common"
)

// UpdateByteDifference ...
const UpdateByteDifference = 50000

// HTTPError ...
type HTTPError struct {
	URL        string
	Method     string
	StatusCode int
	Status     string
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("%s failed for %s with code %d (%s)", e.Method, e.URL, e.StatusCode, e.Status)
}

// Worker ...
type Worker struct {
	ID           uint
	Clock        common.Clock
	FileStore    FileStore
	WorkQueue    chan Download
	ErrorChannel chan Error
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

// WriteData ...
func (w Worker) WriteData(dataReader io.Reader, outputWriter io.Writer, statusWriter *StatusWriter) error {
	teeReader := io.TeeReader(dataReader, statusWriter)

	_, err := io.Copy(outputWriter, teeReader)
	if err != nil {
		return err
	}

	return nil
}

// SendError ...
func (w Worker) SendError(id string, err error) {
	e := Error{DownloadID: id}
	e.Time = w.Clock.Now()
	e.OriginalError = err.Error()

	w.ErrorChannel <- e
}

// SaveWithStatus ...
func (w Worker) SaveWithStatus(download *Download) error {
	downloadHash, err := download.Hash()
	if err != nil {
		w.SendError(download.ID, err)
	}

	statusWriter := NewStatusWriter(download.ID, w.StatusSender, downloadHash, UpdateByteDifference)
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

// Save ...
func (w Worker) Save(download *Download, outputWriter io.Writer, statusWriter *StatusWriter) error {
	download.TimeStarted = time.Now()

	res, err := http.Get(download.URL)
	if err != nil {
		w.SendError(download.ID, err)
		return err
	}

	if res.StatusCode != http.StatusOK {
		err = HTTPError{
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

// NewWorker ...
func NewWorker(id uint, workQueue chan Download, updateChannel chan StatusUpdate, errorChannel chan Error, fileStore FileStore) *Worker {
	worker := &Worker{
		Clock:        &common.RealClock{},
		ID:           id,
		WorkQueue:    workQueue,
		StatusSender: &ChannelStatusSender{StatusChannel: updateChannel},
		ErrorChannel: errorChannel,
		FileStore:    fileStore}

	return worker
}
