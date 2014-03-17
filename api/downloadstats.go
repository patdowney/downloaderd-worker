package api

type Stat struct {
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Mean  float64 `json:"mean"`
	Sum   float64 `json:"sum"`
	Count int     `json:"count"`
}

type DownloadStats struct {
	WaitTime     Stat `json:"wait_time_ms"`
	DownloadTime Stat `json:"download_time_ms"`
	BytesRead    Stat `json:"bytes_read"`
}
