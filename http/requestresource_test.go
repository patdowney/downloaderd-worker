package http_test

import (
	"github.com/gorilla/mux"
	"github.com/patdowney/downloaderd/api"
	dh "github.com/patdowney/downloaderd/http"
	"net/url"
	"strings"
	"testing"
)

func TestUrlResolving(t *testing.T) {
	requestId := "this-is-some-id"

	expectedUrl, _ := url.Parse("/request/this-is-some-id")
	res := dh.RequestResource{}

	router := mux.NewRouter()

	res.RegisterRoutes(router.PathPrefix("/request").Subrouter())

	u, err := res.GetRequestUrl(requestId)

	if err != nil {
		t.Error(err)
	}

	if *u != *expectedUrl {
		t.Errorf(`GetRequestUrl('%s') = %q want %q`, requestId, u, expectedUrl)
	}
}

func TestGetRequestUrl(t *testing.T) {
	requestId := "this-is-some-id"

	expectedUrl, _ := url.Parse("/this-is-some-id")
	res := dh.RequestResource{}
	router := mux.NewRouter()

	res.RegisterRoutes(router)

	u, err := res.GetRequestUrl(requestId)

	if err != nil {
		t.Error(err)
	}

	if *u != *expectedUrl {
		t.Errorf(`GetRequestUrl('%s') = %q want %q`, requestId, u, expectedUrl)
	}
}

func TestParseCompleteRequest(t *testing.T) {
	res := dh.RequestResource{}

	jsonString := `{ "url": "http://example.com/some/resource", "checksum": "abcde", "checksum_type": "abc", "callback": "http://example.com/callback" }`
	incomingJson := strings.NewReader(jsonString)

	expectedIncoming := api.IncomingRequest{
		Url:          "http://example.com/some/resource",
		Checksum:     "abcde",
		ChecksumType: "abc",
		Callback:     "http://example.com/callback"}

	r, _ := res.DecodeInputRequest(incomingJson)
	if *r != expectedIncoming {
		t.Errorf(`DecodeInputRequest('%s') = %q want %q`, jsonString, r, expectedIncoming)
	}
}

func TestParseWithPartialRequest(t *testing.T) {
	res := dh.RequestResource{}

	jsonString := `{ "url": "http://example.com/some/resource" }`
	incomingJson := strings.NewReader(jsonString)

	expectedIncoming := api.IncomingRequest{Url: "http://example.com/some/resource"}

	r, _ := res.DecodeInputRequest(incomingJson)
	if *r != expectedIncoming {
		t.Errorf(`DecodeInputRequest('%s') = %q want %q`, jsonString, r, expectedIncoming)
	}
}

func TestRequestResourceGetIndex(t *testing.T)            {}
func TestRequestResourceGetRequest(t *testing.T)          {}
func TestRequestResourcePostIncomingRequest(t *testing.T) {}

//
//func (r *RequestResource) Index() HandlerFunc {
//	return func(rw http.ResponseWriter, req *http.Request) {
//		rw.Header().Set("Content-Type", "text/json")
//		rw.WriteHeader(http.StatusOK)
//		encoder := json.NewEncoder(rw)
//		encoder.Encode(r.RequestStore.ListAll())
//	}
//}
//
//func (r *RequestResource) Get() HandlerFunc {
//	return func(rw http.ResponseWriter, req *http.Request) {
//		vars := mux.Vars(req)
//		requestId := vars["requestId"]
//
//		downloadRequest := r.RequestStore.FindById(requestId)
//
//		if downloadRequest != nil {
//			rw.Header().Set("Content-Type", "text/json")
//			rw.WriteHeader(http.StatusOK)
//			encoder := json.NewEncoder(rw)
//			encoder.Encode(downloadRequest)
//		} else {
//			rw.WriteHeader(http.StatusNotFound)
//			log.Printf("Couldn't find request with id:%s", downloadRequestId)
//		}
//	}
//}
//
//func (r *RequestResource) decodeIncomingRequest(body io.Reader) (*ApiIncomingRequest, error) {
//	decoder := json.NewDecoder(body)
//	var incomingRequest ApiIncomingRequest
//	err := decoder.Decode(&incomingRequest)
//	if err != nil {
//		return nil, err
//	}
//
//	_, err = incomingRequest.IsValid()
//	if err != nil {
//		return nil, err
//	}
//
//	return &incomingRequest
//}
//
//func (r *RequestResource) Post() HandlerFunc {
//	return func(http.ResponseWriter, *http.Request) {
//		incomingRequest, err := r.decodeInputRequest(req.Body)
//		if err != nil {
//			http.Error(rw, err.Error(), http.StatusBadRequest)
//			return
//		}
//
//		d, err := downloader.ProcessIncomingRequest(&incomingRequest)
//
//		if err != nil {
//			rw.WriteHeader(http.StatusInternalServerError)
//			log.Printf("server-error: %v", err)
//		} else {
//			newUrl := GetUrlForDownloadRequestId(req, d.Id)
//
//			rw.Header().Set("Content-Type", "text/json")
//			rw.Header().Set("Location", newUrl)
//			rw.WriteHeader(http.StatusAccepted)
//			encoder := json.NewEncoder(rw)
//			encoder.Encode(d)
//
//		}
//	}
//}
