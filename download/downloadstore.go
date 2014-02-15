package download

type DownloadStore interface {
	Add(*Download) error
	FindById(string) (*Download, error)
	FindByResourceKey(ResourceKey) (*Download, error)
	ListAll() ([]*Download, error)
	Commit() error
}
