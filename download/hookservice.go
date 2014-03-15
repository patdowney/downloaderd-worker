package download

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/patdowney/downloaderd/api"
	"github.com/patdowney/downloaderd/common"
)

type HookService struct {
	Clock        common.Clock
	hookStore    HookStore
	linkResolver *api.LinkResolver
}

func NewHookService(hookStore HookStore, linkResolver *api.LinkResolver) *HookService {
	s := HookService{
		Clock:        &common.RealClock{},
		hookStore:    hookStore,
		linkResolver: linkResolver}

	return &s
}

func (s *HookService) Register(downloadID string, requestID string, hookURL string) {
	h := NewHook(downloadID, requestID, hookURL)
	s.hookStore.Add(h)
}

func (s *HookService) Notify(download *Download) error {
	downloadHooks, err := s.hookStore.FindByDownloadID(download.ID)
	if err != nil {
		return err
	}
	go s.notifyHooks(downloadHooks, download)

	return nil
}

func (s *HookService) notifyHooks(hooks []*Hook, download *Download) {
	for _, h := range hooks {
		if h.Result == nil {
			hr, err := s.notifyHook(h, download)
			if err != nil {
				log.Printf("notify-hook: downloadID: %s, url: %s, %v", download.ID, h.URL, err)
			}
			h.Result = hr
			s.hookStore.Update(h)
		}
	}
}

func (s *HookService) notifyHook(hook *Hook, download *Download) (*HookResult, error) {
	hr := NewHookResult()
	hr.Time = s.Clock.Now()

	apiDownload := ToAPIDownload(download)
	apiDownload.ResolveLinks(s.linkResolver, nil)
	jsonBytes, err := json.Marshal(apiDownload)
	if err != nil {
		return nil, err
	}

	byteReader := bytes.NewReader(jsonBytes)
	res, err := http.Post(hook.URL, "application/json", byteReader)
	if err != nil {
		return nil, err
	}

	hr.StatusCode = res.StatusCode

	if res.StatusCode != http.StatusOK {
		e := common.HTTPError{
			URL:        hook.URL,
			Method:     "Post",
			StatusCode: res.StatusCode,
			Status:     res.Status}

		hr.AddError(e)
	}
	return hr, nil
}

func (s *HookService) FindByDownloadID(id string) ([]*Hook, error) {
	return s.hookStore.FindByDownloadID(id)
}

func (s *HookService) FindByRequestID(id string) ([]*Hook, error) {
	return s.hookStore.FindByRequestID(id)
}
