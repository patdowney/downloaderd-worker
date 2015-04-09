package download

// Request ...
type Request struct {
	ID            string
	URL           string
	Checksum      string
	ChecksumType  string
	Callback      string
	ETag          string
	ContentLength uint64
}

// ResourceKey ...
func (r *Request) ResourceKey() ResourceKey {
	rk := ResourceKey{URL: r.URL}
	if r.ETag != "" {
		rk.ETag = r.ETag
	}
	return rk
}
