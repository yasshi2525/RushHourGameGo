package services

import (
	"testing"
)

func TestCreateResidence(t *testing.T) {
	InitStorage()
	Config.Residence.Capacity = 10
	Config.Residence.Interval = 1

	residence := CreateResidence(1, 1)

	if got := residence.Available; got > 1 {
		t.Errorf("Available should <= 1, but = %f", got)
	}

	RemoveResidence(residence.ID)
}

func TestCreateCompany(t *testing.T) {
	InitStorage()
	Config.Company.Scale = 1

	RemoveCompany(CreateCompany(1, 1).ID)
}

func TestCreateStep(t *testing.T) {
	InitStorage()

	r := CreateResidence(1, 1)
	c := CreateCompany(2, 2)

	if got := len(r.Out); got != 1 {
		t.Errorf("Residence should be out 1, but %d", got)
	}
	if got := len(c.In); got != 1 {
		t.Errorf("Company should be in 1, but %d", got)
	}

	RemoveResidence(r.ID)
	RemoveCompany(c.ID)

	if got := len(Static.Steps); got != 0 {
		t.Errorf("Steps size should be 0, but %d", got)
	}
}
