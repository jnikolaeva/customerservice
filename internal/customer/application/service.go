package application

import (
	"context"
	"errors"

	"github.com/jnikolaeva/eshop-common/uuid"
)

var (
	ErrCustomerNotFound = errors.New("user not found")
	// TODO: email is not really unique value
	ErrDuplicateUser = errors.New("user with such email already exists")
)

type Service interface {
	Create(ctx context.Context, id uuid.UUID, firstName, lastName, email, phone string) (CustomerID, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Customer, error)
	Update(ctx context.Context, id uuid.UUID, firstName, lastName, email, phone string) (*Customer, error)
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

type service struct {
	repo Repository
}

func (s service) Create(ctx context.Context, id uuid.UUID, firstName, lastName, email, phone string) (CustomerID, error) {
	user := Customer{
		ID:        CustomerID(id),
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Phone:     phone,
	}

	if err := s.repo.Add(user); err != nil {
		return user.ID, err
	}

	return user.ID, nil
}

func (s service) FindByID(ctx context.Context, id uuid.UUID) (*Customer, error) {
	user, err := s.repo.FindByID(CustomerID(id))
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s service) Update(ctx context.Context, id uuid.UUID, firstName, lastName, email, phone string) (*Customer, error) {
	user, err := s.repo.FindByID(CustomerID(id))
	if err != nil {
		return nil, err
	}

	user.FirstName = firstName
	user.LastName = lastName
	user.Email = email
	user.Phone = phone

	return user, s.repo.Update(*user)
}
