package download

type HookStore interface {
	Add(*Hook) error
	Update(*Hook) error
	FindByRequestID(requetID string) ([]*Hook, error)
	FindByDownloadID(downloadID string) ([]*Hook, error)
	ListAll() ([]*Hook, error)
}
