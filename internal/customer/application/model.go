package application

import "github.com/jnikolaeva/eshop-common/uuid"

type CustomerID uuid.UUID

func (u CustomerID) String() string {
	return uuid.UUID(u).String()
}

type Customer struct {
	ID        CustomerID
	FirstName string
	LastName  string
	Email     string
	Phone     string
}

type Repository interface {
	Add(user Customer) error
	FindByID(id CustomerID) (*Customer, error)
	Update(user Customer) error
}
