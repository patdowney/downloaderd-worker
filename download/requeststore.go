package download

type RequestStore interface {
	Add(*Request) error
	FindByID(string) (*Request, error)
	FindByResourceKey(ResourceKey, uint, uint) ([]*Request, error)
	FindAll(uint, uint) ([]*Request, error)
}
