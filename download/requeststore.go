package download

type RequestStore interface {
	Add(*Request) error
	FindByID(string) (*Request, error)
	FindByResourceKey(ResourceKey) ([]*Request, error)
	ListAll() ([]*Request, error)
}
