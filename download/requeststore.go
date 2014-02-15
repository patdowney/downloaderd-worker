package download

type RequestStore interface {
	Add(*Request) error
	FindById(string) (*Request, error)
	FindByResourceKey(ResourceKey) ([]*Request, error)
	ListAll() ([]*Request, error)
}
