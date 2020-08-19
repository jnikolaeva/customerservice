package application

import (
	"context"
	"errors"

	"github.com/jnikolaeva/eshop-common/uuid"
)

var ErrNotAuthorized = errors.New("operation not allowed")

type auth struct {
	service Service
}

func NewAuthService(service Service) Service {
	return &auth{
		service: service,
	}
}

func (a auth) Create(ctx context.Context, id uuid.UUID, firstName, lastName, email, phone string) (CustomerID, error) {
	return a.service.Create(ctx, id, firstName, lastName, email, phone)
}

func (a auth) FindByID(ctx context.Context, id uuid.UUID) (*Customer, error) {
	if !isResourceOwner(ctx, id) {
		return nil, ErrNotAuthorized
	}
	return a.service.FindByID(ctx, id)
}

func (a auth) Update(ctx context.Context, id uuid.UUID, firstName, lastName, email, phone string) (*Customer, error) {
	if !isResourceOwner(ctx, id) {
		return nil, ErrNotAuthorized
	}
	return a.service.Update(ctx, id, firstName, lastName, email, phone)
}

func isResourceOwner(ctx context.Context, resourceID uuid.UUID) bool {
	subjectID := GetUserID(ctx)
	return subjectID != nil && resourceID == *subjectID
}