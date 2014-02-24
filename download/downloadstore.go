package download

type DownloadStore interface {
	Add(*Download) error
	FindByID(string) (*Download, error)
	FindByResourceKey(ResourceKey) (*Download, error)
	ListAll() ([]*Download, error)
	Update(*Download) error
	Commit() error
}
