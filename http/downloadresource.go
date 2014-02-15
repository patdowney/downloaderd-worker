package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/patdowney/downloaderd/api"
	"github.com/patdowney/downloaderd/download"
	"log"
	"net/http"
)

type DownloadResource struct {
	DownloadService *download.DownloadService
	router          *mux.Router
}

func NewDownloadResource(downloadService *download.DownloadService) *DownloadResource {
	return &DownloadResource{DownloadService: downloadService}
}

func (r *DownloadResource) RegisterRoutes(parentRouter *mux.Router) {
	parentRouter.HandleFunc("/", r.Index()).Methods("GET", "HEAD")
	// regexp matches ids that look like '8671301b-49fa-416c-4bc0-2869963779e5'
	parentRouter.HandleFunc("/{id:[a-f0-9-]{36}}", r.Get()).Methods("GET", "HEAD").Name("request")

	r.router = parentRouter
}

func (r *DownloadResource) Index() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		downloadList, err := r.DownloadService.ListAll()

		encoder := json.NewEncoder(rw)
		rw.Header().Set("Content-Type", "application/json")

		if err != nil {
			log.Printf("server-error: %v", err)
			rw.WriteHeader(http.StatusInternalServerError)
			encoder.Encode(api.NewError(err))
		} else {
			rw.WriteHeader(http.StatusOK)

			encoder.Encode(api.NewDownloadList(&downloadList))
		}
	}
}

func (r *DownloadResource) Get() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		downloadId := vars["id"]

		download, err := r.DownloadService.FindById(downloadId)

		encoder := json.NewEncoder(rw)
		rw.Header().Set("Content-Type", "application/json")

		if err != nil {
			log.Printf("server-error: %v", err)
			rw.WriteHeader(http.StatusInternalServerError)
			encoder.Encode(api.NewError(err))
		} else if download != nil {
			rw.WriteHeader(http.StatusOK)
			encoder.Encode(api.NewDownload(download))
		} else {
			errMessage := fmt.Sprintf("Unable to find order with id:%s", downloadId)
			log.Printf("server-error: %v", errMessage)

			rw.WriteHeader(http.StatusNotFound)
			encoder.Encode(errors.New(errMessage))
		}
	}
}
