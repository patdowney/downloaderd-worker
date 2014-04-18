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
	"github.com/patdowney/downloaderd/api"
	"github.com/patdowney/downloaderd/common"
	"github.com/patdowney/downloaderd/download"
)

type DownloadResource struct {
	Clock           common.Clock
	DownloadService *download.DownloadService
	router          *mux.Router
	linkResolver    *api.LinkResolver
}

func NewDownloadResource(downloadService *download.DownloadService, linkResolver *api.LinkResolver) *DownloadResource {
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

func (r *DownloadResource) RegisterRoutes(parentRouter *mux.Router) {
	parentRouter.HandleFunc("/", r.Index(r.AllIndex())).Methods("GET", "HEAD")

	parentRouter.HandleFunc("/finished", r.Index(r.FinishedIndex())).Methods("GET", "HEAD")
	parentRouter.HandleFunc("/finished/stats", r.Stats(r.FinishedIndex())).Methods("GET", "HEAD")

	parentRouter.HandleFunc("/notfinished", r.Index(r.NotFinishedIndex())).Methods("GET", "HEAD")
	parentRouter.HandleFunc("/notfinished/stats", r.Stats(r.NotFinishedIndex())).Methods("GET", "HEAD")

	parentRouter.HandleFunc("/inprogress", r.Index(r.InProgressIndex())).Methods("GET", "HEAD")
	parentRouter.HandleFunc("/inprogress/stats", r.Stats(r.InProgressIndex())).Methods("GET", "HEAD")

	parentRouter.HandleFunc("/waiting", r.Index(r.WaitingIndex())).Methods("GET", "HEAD")
	parentRouter.HandleFunc("/waiting/stats", r.Stats(r.WaitingIndex())).Methods("GET", "HEAD")

	parentRouter.HandleFunc("/all", r.Index(r.AllIndex())).Methods("GET", "HEAD")
	parentRouter.HandleFunc("/all/stats", r.Stats(r.AllIndex())).Methods("GET", "HEAD")

	// regexp matches ids that look like '8671301b-49fa-416c-4bc0-2869963779e5'
	parentRouter.HandleFunc("/{id:[a-f0-9-]{36}}", r.Get()).Methods("GET", "HEAD").Name("download")
	parentRouter.HandleFunc("/{id:[a-f0-9-]{36}}", r.Delete()).Methods("DELETE").Name("download-delete")

	parentRouter.HandleFunc("/{id:[a-f0-9-]{36}}/data", r.GetData()).Methods("GET", "HEAD").Name("download-data")

	parentRouter.HandleFunc("/{id:[a-f0-9-]{36}}/verify", r.VerifyData()).Methods("GET", "HEAD").Name("download-verify")

	r.router = parentRouter
	r.linkResolver = api.NewLinkResolver(parentRouter)
}

func (r *DownloadResource) WrapError(err error) *api.Error {
	return download.ToAPIError(common.NewErrorWrapper(err, r.Clock.Now()))
}

type IndexFunc func() ([]*download.Download, error)

func (r *DownloadResource) FinishedIndex() IndexFunc {
	return r.DownloadService.ListFinished
}
func (r *DownloadResource) NotFinishedIndex() IndexFunc {
	return r.DownloadService.ListNotFinished
}
func (r *DownloadResource) InProgressIndex() IndexFunc {
	return r.DownloadService.ListInProgress
}
func (r *DownloadResource) WaitingIndex() IndexFunc {
	return r.DownloadService.ListWaiting
}
func (r *DownloadResource) AllIndex() IndexFunc {
	return r.DownloadService.ListAll
}

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
			}
		}
	}
}

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

			stats := download.DownloadStats{Clock: r.Clock}
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
