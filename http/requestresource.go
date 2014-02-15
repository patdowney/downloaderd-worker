package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/patdowney/downloaderd/api"
	"github.com/patdowney/downloaderd/common"
	"github.com/patdowney/downloaderd/download"
	"io"
	"log"
	"net/http"
	"net/url"
)

type RequestResource struct {
	Clock          common.Clock
	RequestService *download.RequestService
	router         *mux.Router
}

func NewRequestResource(requestService *download.RequestService) *RequestResource {
	return &RequestResource{
		Clock:          &common.RealClock{},
		RequestService: requestService}
}

func (r *RequestResource) RegisterRoutes(parentRouter *mux.Router) {
	parentRouter.HandleFunc("/", r.Index()).Methods("GET", "HEAD")
	parentRouter.HandleFunc("/", r.Post()).Methods("POST")
	// regexp matches ids that look like '8671301b-49fa-416c-4bc0-2869963779e5'
	parentRouter.HandleFunc("/{id:[a-f0-9-]{36}}", r.Get()).Methods("GET", "HEAD").Name("request")

	r.router = parentRouter
}

func (r *RequestResource) WrapError(err error) *api.Error {
	return api.NewError(common.NewErrorWrapper(err, r.Clock.Now()))
}

func (r *RequestResource) Index() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		requestList, err := r.RequestService.ListAll()

		encoder := json.NewEncoder(rw)
		rw.Header().Set("Content-Type", "application/json")

		if err != nil {
			log.Printf("server-error: %v", err)
			rw.WriteHeader(http.StatusInternalServerError)
			encoder.Encode(r.WrapError(err))
		} else {
			rw.WriteHeader(http.StatusOK)

			encoder.Encode(api.NewRequestList(&requestList))
		}
	}
}

func (r *RequestResource) Get() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		vars := mux.Vars(req)
		requestId := vars["id"]

		downloadRequest, err := r.RequestService.FindById(requestId)

		encoder := json.NewEncoder(rw)
		rw.Header().Set("Content-Type", "application/json")

		if err != nil {
			log.Printf("server-error: %v", err)
			rw.WriteHeader(http.StatusInternalServerError)
			encoder.Encode(r.WrapError(err))
		} else if downloadRequest != nil {
			rw.WriteHeader(http.StatusOK)
			encoder.Encode(api.NewRequest(downloadRequest))
		} else {
			errMessage := fmt.Sprintf("Unable to find request with id:%s", requestId)
			log.Printf("server-error: %v", errMessage)

			rw.WriteHeader(http.StatusNotFound)
			encoder.Encode(errors.New(errMessage))
		}
	}
}

func (r *RequestResource) ValidateIncomingRequest(inReq *api.IncomingRequest) error {
	if inReq.Url == "" {
		return errors.New("empty url")
	}

	u, err := url.Parse(inReq.Url)
	if err != nil {
		return err
	} else if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New(fmt.Sprintf("unsupported url scheme: '%s'", u.Scheme))
	}
	return nil
}

func (r *RequestResource) DecodeInputRequest(body io.Reader) (*api.IncomingRequest, error) {
	decoder := json.NewDecoder(body)
	var inReq api.IncomingRequest
	err := decoder.Decode(&inReq)
	if err != nil {
		return nil, err
	}

	return &inReq, nil
}

func (r *RequestResource) GetRequestUrl(id string) (*url.URL, error) {
	if r.router != nil {
		return r.router.Get("request").URL("id", id)
	}

	return nil, errors.New("no router set")
}

func (r *RequestResource) Post() http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		apiIncomingRequest, err := r.DecodeInputRequest(req.Body)
		if err != nil {
			log.Printf("incoming-request-decode-error: %v", err)
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		err = r.ValidateIncomingRequest(apiIncomingRequest)
		if err != nil {
			log.Printf("incoming-request-validation-error: %v", err)
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		inReq := apiIncomingRequest.ToDownloadRequest()
		downloadRequest, err := r.RequestService.ProcessNewRequest(inReq)

		if err != nil {
			log.Printf("request-processing-error: %v", err)
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusInternalServerError)
			encoder := json.NewEncoder(rw)
			encoder.Encode(r.WrapError(err))
		} else {
			newUrl, _ := r.GetRequestUrl(downloadRequest.Id)

			rw.Header().Set("Content-Type", "application/json")
			rw.Header().Set("Location", newUrl.String())
			rw.WriteHeader(http.StatusAccepted)
			encoder := json.NewEncoder(rw)
			encoder.Encode(api.NewRequest(downloadRequest))

		}
	}
}
