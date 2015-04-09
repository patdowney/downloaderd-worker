package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/patdowney/downloaderd-common/common"
	"github.com/patdowney/downloaderd-worker/api"
	"github.com/patdowney/downloaderd-worker/download"
)

// DownloadResource ...
type DownloadResource struct {
	Clock           common.Clock
	DownloadService *download.Service
	router          *mux.Router
	linkResolver    *api.LinkResolver
}

// NewDownloadResource ...
func NewDownloadResource(downloadService *download.Service, linkResolver *api.LinkResolver) *DownloadResource {
	return &DownloadResource{
		Clock:           &common.RealClock{},
		DownloadService: downloadService,
		linkResolver:    linkResolver}
}

func (r *DownloadResource) populateListLinks(req *http.Request, downloadList *[]*api.Download) {
	for _, l := range *downloadList {
		r.populateLinks(req, l)
	}
}

func (r *DownloadResource) populateLinks(req *http.Request, download *api.Download) {
	download.ResolveLinks(r.linkResolver, req)
}

// RegisterRoutes ...
func (r *DownloadResource) RegisterRoutes(parentRouter *mux.Router) {
	parentRouter.HandleFunc("/", r.Post()).Methods("POST")
	parentRouter.HandleFunc("/", r.Index(r.AllIndex())).Methods("GET", "HEAD")

	// regexp matches ids that look like '8671301b-49fa-416c-4bc0-2869963779e5'
	parentRouter.HandleFunc("/{id:[a-f0-9-]{36}}", r.Get()).Methods("GET", "HEAD").Name("download")
	parentRouter.HandleFunc("/{id:[a-f0-9-]{36}}", r.Delete()).Methods("DELETE").Name("download-delete")
	parentRouter.HandleFunc("/{id:[a-f0-9-]{36}}/data", r.GetData()).Methods("GET", "HEAD").Name("download-data")
	parentRouter.HandleFunc("/{id:[a-f0-9-]{36}}/verify", r.VerifyData()).Methods("GET", "HEAD").Name("download-verify")

	// predefined searches
	parentRouter.HandleFunc("/all", r.Index(r.AllIndex())).Methods("GET", "HEAD")
	parentRouter.HandleFunc("/all/stats", r.Stats(r.AllIndex())).Methods("GET", "HEAD")

	parentRouter.HandleFunc("/finished", r.Index(r.FinishedIndex())).Methods("GET", "HEAD")
	parentRouter.HandleFunc("/finished/stats", r.Stats(r.FinishedIndex())).Methods("GET", "HEAD")

	parentRouter.HandleFunc("/notfinished", r.Index(r.NotFinishedIndex())).Methods("GET", "HEAD")
	parentRouter.HandleFunc("/notfinished/stats", r.Stats(r.NotFinishedIndex())).Methods("GET", "HEAD")

	parentRouter.HandleFunc("/inprogress", r.Index(r.InProgressIndex())).Methods("GET", "HEAD")
	parentRouter.HandleFunc("/inprogress/stats", r.Stats(r.InProgressIndex())).Methods("GET", "HEAD")

	parentRouter.HandleFunc("/waiting", r.Index(r.WaitingIndex())).Methods("GET", "HEAD")
	parentRouter.HandleFunc("/waiting/stats", r.Stats(r.WaitingIndex())).Methods("GET", "HEAD")

	r.router = parentRouter
	r.linkResolver = api.NewLinkResolver(parentRouter)
}

// WrapError ...
func (r *DownloadResource) WrapError(err error) *api.Error {
	return download.ToAPIError(common.NewTimestampedError(err, r.Clock.Now()))
}

// IndexFunc ...
type IndexFunc func() ([]*download.Download, error)

// FinishedIndex ...
func (r *DownloadResource) FinishedIndex() IndexFunc {
	return r.DownloadService.ListFinished

}

// NotFinishedIndex ...
func (r *DownloadResource) NotFinishedIndex() IndexFunc {
	return r.DownloadService.ListNotFinished
}

// InProgressIndex ...
func (r *DownloadResource) InProgressIndex() IndexFunc {
	return r.DownloadService.ListInProgress
}

// WaitingIndex ...
func (r *DownloadResource) WaitingIndex() IndexFunc {
	return r.DownloadService.ListWaiting
}

// AllIndex ...
func (r *DownloadResource) AllIndex() IndexFunc {
	return r.DownloadService.ListAll
}

// Index ...
func (r *DownloadResource) Index(indexFunc IndexFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		downloadList, err := indexFunc()

		encoder := json.NewEncoder(rw)
		rw.Header().Set("Content-Type", "application/json")

		if err != nil {
			log.Printf("server-error: %v", err)
			rw.WriteHeader(http.StatusInternalServerError)
			encErr := encoder.Encode(r.WrapError(err))
			if encErr != nil {
				log.Printf("encoder-error: %v", encErr)
			}
		} else {
			rw.WriteHeader(http.StatusOK)
			dl := download.ToAPIDownloadList(&downloadList)
			r.populateListLinks(req, dl)
			encErr := encoder.Encode(dl)
			if encErr != nil {
				log.Printf("encoder-error: %v", encErr)
				log.Printf("encoder-error-struct: %v", dl)
			}
		}
	}
}

// Stats ...
func (r *DownloadResource) Stats(indexFunc IndexFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		downloadList, err := indexFunc()

		encoder := json.NewEncoder(rw)
		rw.Header().Set("Content-Type", "application/json")

		if err != nil {
			log.Printf("server-error: %v", err)
			rw.WriteHeader(http.StatusInternalServerError)
			encErr := encoder.Encode(r.WrapError(err))
			if encErr != nil {
				log.Printf("encoder-error: %v", encErr)
			}
		} else {
			rw.Header().Set("Access-Control-Allow-Origin", "*")

			stats := download.Stats{Clock: r.Clock}
			stats.AddList(downloadList)
			rw.WriteHeader(http.StatusOK)
			ds := download.ToAPIDownloadStats(&stats)
			//r.populateListLinks(req, dl)
			encErr := encoder.Encode(ds)
			if encErr != nil {
				log.Printf("encoder-error: %v", encErr)
			}
		}
	}
}

// VerifyData ...
func (r *DownloadResource) VerifyData() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		downloadID := vars["id"]

		download, err := r.DownloadService.FindByID(downloadID)
		encoder := json.NewEncoder(rw)

		if err != nil {
			log.Printf("server-error-verify-data(%s): %v", downloadID, err)
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusInternalServerError)
			encErr := encoder.Encode(r.WrapError(err))
			if encErr != nil {
				log.Printf("encoder-error-verify-data(%s): %v", downloadID, encErr)
			}
		} else if download != nil {
			if download.Finished {
				ok, err := r.DownloadService.Verify(download)
				if err != nil {
					log.Printf("server-error-verify-data(%s): %v", downloadID, err)
					rw.Header().Set("Content-Type", "application/json")
					rw.WriteHeader(http.StatusInternalServerError)
					encErr := encoder.Encode(r.WrapError(err))
					if encErr != nil {
						log.Printf("encoder-error-verify-data(%s): %v", downloadID, encErr)
					}
					return
				}

				if ok {
					rw.WriteHeader(http.StatusOK)
				} else {
					rw.WriteHeader(http.StatusConflict)
				}

				encErr := encoder.Encode(ok)
				if encErr != nil {
					log.Printf("encoder-error-verify-data(%s): %v", downloadID, encErr)
				}
			} else {
				rw.WriteHeader(http.StatusPartialContent)
			}
		} else {
			rw.WriteHeader(http.StatusNotFound)
		}

	}
}

// GetData ...
func (r *DownloadResource) GetData() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		downloadID := vars["id"]

		download, err := r.DownloadService.FindByID(downloadID)
		encoder := json.NewEncoder(rw)

		if err != nil {
			log.Printf("server-error-get-data(%s): %v", downloadID, err)
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusInternalServerError)
			encErr := encoder.Encode(r.WrapError(err))
			if encErr != nil {
				log.Printf("encoder-error-get-data(%s): %v", downloadID, encErr)
			}
		} else if download != nil {
			if download.Finished {
				bufferedReader, err := r.DownloadService.GetReader(download)
				if err != nil {
					log.Printf("server-error-get-data(%s): %v", downloadID, err)
					rw.Header().Set("Content-Type", "application/json")
					rw.WriteHeader(http.StatusInternalServerError)
					encErr := encoder.Encode(r.WrapError(err))
					if encErr != nil {
						log.Printf("encoder-error-get-data(%s): %v", downloadID, encErr)
					}
					return
				}

				meta := download.Metadata

				if meta.MimeType != "" {
					rw.Header().Set("Content-Type", meta.MimeType)
				}
				if meta.Size != 0 {
					rw.Header().Set("Content-Length", fmt.Sprintf("%d", meta.Size))
				} else {
					rw.Header().Set("Content-Length", fmt.Sprintf("%d", download.Status.BytesRead))
				}

				u, _ := url.Parse(download.URL)
				filename := filepath.Base(u.Path)
				rw.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

				rw.WriteHeader(http.StatusOK)

				io.Copy(rw, bufferedReader)
			} else {
				rw.WriteHeader(http.StatusNoContent)
			}
		} else {
			rw.WriteHeader(http.StatusNotFound)
		}

	}
}

// Delete ...
func (r *DownloadResource) Delete() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		downloadID := vars["id"]

		deleted, err := r.DownloadService.DeleteByID(downloadID)

		encoder := json.NewEncoder(rw)
		rw.Header().Set("Content-Type", "application/json")

		if err != nil {
			log.Printf("server-error-get(%s): %v", downloadID, err)
			rw.WriteHeader(http.StatusInternalServerError)
			encErr := encoder.Encode(r.WrapError(err))
			if encErr != nil {
				log.Printf("encoder-error-get(%s): %v", downloadID, encErr)
			}
		} else if deleted {
			log.Printf("deleted-download-with-id: %v", downloadID)

			rw.WriteHeader(http.StatusOK)
		} else {
			log.Printf("deleted-download-not-found: %v", downloadID)

			rw.WriteHeader(http.StatusNotFound)
		}
	}
}

// Get ...
func (r *DownloadResource) Get() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		downloadID := vars["id"]

		foundDownload, err := r.DownloadService.FindByID(downloadID)

		encoder := json.NewEncoder(rw)
		rw.Header().Set("Content-Type", "application/json")

		if err != nil {
			log.Printf("server-error-get(%s): %v", downloadID, err)
			rw.WriteHeader(http.StatusInternalServerError)
			encErr := encoder.Encode(r.WrapError(err))
			if encErr != nil {
				log.Printf("encoder-error-get(%s): %v", downloadID, encErr)
			}
		} else if foundDownload != nil {
			rw.WriteHeader(http.StatusOK)
			d := download.ToAPIDownload(foundDownload)
			r.populateLinks(req, d)
			encErr := encoder.Encode(d)
			if encErr != nil {
				log.Printf("encoder-error-get(%s): %v", downloadID, encErr)
			}
		} else {
			errMessage := fmt.Sprintf("unable to find download with id:%s", downloadID)
			log.Printf("server-error-get(%s): %v", downloadID, errMessage)

			rw.WriteHeader(http.StatusNotFound)
			encErr := encoder.Encode(errors.New(errMessage))
			if encErr != nil {
				log.Printf("encoder-error-get(%s): %v", downloadID, encErr)
			}
		}
	}
}

// GetDownloadURL ...
func (r *DownloadResource) GetDownloadURL(id string) (*url.URL, error) {
	if r.router != nil {
		return r.router.Get("download").URL("id", id)
	}

	return nil, errors.New("no router set")
}

// DecodeIncomingDownload ...
func (r *DownloadResource) DecodeIncomingDownload(body io.Reader) (*api.IncomingDownload, error) {
	decoder := json.NewDecoder(body)
	var inDown api.IncomingDownload
	err := decoder.Decode(&inDown)
	if err != nil {
		return nil, err
	}

	return &inDown, nil
}

// ValidateIncomingDownload ...
func (r *DownloadResource) ValidateIncomingDownload(inDown *api.IncomingDownload) error {
	if inDown.URL == "" {
		return errors.New("empty url")
	}

	u, err := url.Parse(inDown.URL)
	if err != nil {
		return err
	} else if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("unsupported url scheme: '%s'", u.Scheme)
	}
	return nil
}

// IncomingDownload ...
type IncomingDownload struct {
	URL string
}

// Post ...
func (r *DownloadResource) Post() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		incomingDownload, err := r.DecodeIncomingDownload(req.Body)
		if err != nil {
			log.Printf("incoming-request-decode-error: %v", err)
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		err = r.ValidateIncomingDownload(incomingDownload)
		if err != nil {
			log.Printf("incoming-request-validation-error: %v", err)
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		downloadReq := download.FromAPIIncomingDownload(incomingDownload)
		d, err := r.DownloadService.ProcessRequest(downloadReq)

		var encErr error
		encoder := json.NewEncoder(rw)
		rw.Header().Set("Content-Type", "application/json")

		if err != nil {
			log.Printf("server-error-post(%s): %v", d.ID, err)
			rw.WriteHeader(http.StatusInternalServerError)
			encErr = encoder.Encode(r.WrapError(err))
		} else {
			newURL, _ := r.GetDownloadURL(d.ID)
			rw.Header().Set("Location", newURL.String())
			rw.WriteHeader(http.StatusAccepted)
			da := download.ToAPIDownload(d)
			//			r.populateLinks(req, da)
			encErr = encoder.Encode(da)
		}
		if encErr != nil {
			log.Printf("encoder-error-post(%s): %v", d.ID, encErr)
		}
	}
}
