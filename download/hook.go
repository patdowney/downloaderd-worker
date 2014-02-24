package download

type Hook struct {
	DownloadID string
	RequestID  string
	URL        string

	Result *HookResult
}

func NewHook(downloadID string, requestID string, url string) *Hook {
	h := Hook{
		DownloadID: downloadID,
		RequestID:  requestID,
		URL:        url}

	return &h
}
