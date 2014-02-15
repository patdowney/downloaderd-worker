package download

import (
	"github.com/nu7hatch/gouuid"
)

type IDGenerator interface {
	GenerateID() (string, error)
}

type UUIDGenerator struct{}

func (g *UUIDGenerator) GenerateID() (string, error) {
	uuid, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	return uuid.String(), nil
}

type FakeIDGenerator struct {
	FakeID string
}

func (g *FakeIDGenerator) GenerateID() (string, error) {
	return g.FakeID, nil
}
