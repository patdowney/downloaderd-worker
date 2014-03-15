package download

type DownloadStore interface {
	Add(*Download) error
	Update(*Download) error
	FindByID(string) (*Download, error)
	FindByResourceKey(ResourceKey) (*Download, error)
	FindAll() ([]*Download, error)
	FindFinished() ([]*Download, error)
	FindNotFinished() ([]*Download, error)
	FindInProgress() ([]*Download, error)
	FindWaiting() ([]*Download, error)
}
