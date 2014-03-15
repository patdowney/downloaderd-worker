package download

type RequestStore interface {
	Add(*Request) error
	FindByID(string) (*Request, error)
	FindByResourceKey(ResourceKey) ([]*Request, error)
	FindAll() ([]*Request, error)
}
