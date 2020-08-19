package transport

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/jnikolaeva/eshop-common/uuid"
	"github.com/pkg/errors"

	"github.com/jnikolaeva/customerservice/internal/customer/application"
	"github.com/jnikolaeva/customerservice/internal/customer/infrastructure/identity"
)

type IdentityProviderProxy interface {
	Register(username, password string) (uuid.UUID, error)
	Delete(userID uuid.UUID) error
}

type Endpoints struct {
	RegisterCustomer   endpoint.Endpoint
	GetCurrentCustomer endpoint.Endpoint
	FindCustomer       endpoint.Endpoint
	UpdateCustomer     endpoint.Endpoint
}

func MakeEndpoints(s application.Service, identityProviderUrl string) Endpoints {
	identityProvider := identity.NewProviderProxy(identityProviderUrl)

	return Endpoints{
		RegisterCustomer:   makeRegisterCustomerEndpoint(s, identityProvider),
		GetCurrentCustomer: makeGetCurrentCustomerEndpoint(s),
		FindCustomer:       makeFindCustomerEndpoint(s),
		UpdateCustomer:     makeUpdateCustomerEndpoint(s),
	}
}

func makeRegisterCustomerEndpoint(s application.Service, identityProviderProxy IdentityProviderProxy) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(registerCustomerRequest)

		id, err := identityProviderProxy.Register(req.Username, req.Password)
		if err != nil {
			return nil, err
		}

		customerID, err := s.Create(ctx, id, req.FirstName, req.LastName, req.Email, req.Phone)
		if err != nil {
			deleteErr := identityProviderProxy.Delete(id)
			if deleteErr != nil {
				err = errors.Wrap(err, deleteErr.Error())
			}
			return nil, err
		}

		return &registerCustomerResponse{ID: customerID.String()}, err
	}
}

func makeGetCurrentCustomerEndpoint(s application.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		user, err := s.FindByID(ctx, *application.GetUserID(ctx))
		if err != nil {
			return nil, err
		}
		return &findCustomerResponse{toUserData(*user)}, err
	}
}

func makeFindCustomerEndpoint(s application.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(findCustomerRequest)
		user, err := s.FindByID(ctx, req.ID)
		if err != nil {
			return nil, err
		}
		return &findCustomerResponse{toUserData(*user)}, err
	}
}

func makeUpdateCustomerEndpoint(s application.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateCustomerRequest)
		user, err := s.Update(ctx, req.ID, req.FirstName, req.LastName, req.Email, req.Phone)
		if err != nil {
			return nil, err
		}
		return &updateCustomerResponse{toUserData(*user)}, err
	}
}

func toUserData(user application.Customer) userData {
	return userData{
		ID: user.ID.String(),
		userDetails: userDetails{
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Phone:     user.Phone,
		},
	}
}
