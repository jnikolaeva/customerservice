package transport

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-kit/kit/log"
	gokittransport "github.com/go-kit/kit/transport"
	gokithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/jnikolaeva/eshop-common/uuid"
	"github.com/pkg/errors"

	"github.com/jnikolaeva/customerservice/internal/customer/application"
)

const userIDHeader = "X-Auth-User-Id"

var (
	ErrBadRouting       = errors.New("bad routing")
	ErrNotAuthenticated = errors.New("user is not authenticated")
	ErrBadRequest       = errors.New("bad request")
)

func MakeHandler(pathPrefix string, endpoints Endpoints, errorLogger log.Logger) http.Handler {
	options := []gokithttp.ServerOption{
		gokithttp.ServerErrorEncoder(encodeErrorResponse),
		gokithttp.ServerErrorHandler(gokittransport.NewLogErrorHandler(errorLogger)),
	}

	registerCustomerHandler := gokithttp.NewServer(endpoints.RegisterCustomer, decodeRegisterCustomerRequest, encodeResponse, options...)
	getCurrentCustomerHandler := gokithttp.NewServer(endpoints.GetCurrentCustomer, decodeGetCurrentCustomerRequest, encodeResponse, options...)
	findCustomerHandler := gokithttp.NewServer(endpoints.FindCustomer, decodeFindCustomerRequest, encodeResponse, options...)
	updateCustomerHandler := gokithttp.NewServer(endpoints.UpdateCustomer, decodeUpdateCustomerRequest, encodeResponse, options...)

	r := mux.NewRouter()
	s := r.PathPrefix(pathPrefix).Subrouter()
	s.Handle("", registerCustomerHandler).Methods(http.MethodPost)
	s.Handle("/me", authMiddleware(getCurrentCustomerHandler)).Methods(http.MethodGet)
	s.Handle("/{userId}", authMiddleware(findCustomerHandler)).Methods(http.MethodGet)
	s.Handle("/{userId}", authMiddleware(updateCustomerHandler)).Methods(http.MethodPut)
	return r
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := uuid.FromString(r.Header.Get(userIDHeader))
		if err != nil {
			encodeErrorResponse(r.Context(), ErrNotAuthenticated, w)
			return
		}
		ctx := application.WithUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func decodeRegisterCustomerRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req registerCustomerRequest
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil && e != io.EOF {
		return nil, e
	}
	if req.Username == "" {
		return nil, errors.WithMessage(ErrBadRequest, "missing required parameter 'username'")
	}
	if req.Password == "" {
		return nil, errors.WithMessage(ErrBadRequest, "missing required parameter 'password'")
	}
	if req.Email == "" {
		return nil, errors.WithMessage(ErrBadRequest, "missing required parameter 'email'")
	}
	return req, nil
}

func decodeGetCurrentCustomerRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	return nil, nil
}

func decodeFindCustomerRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	sID, ok := vars["userId"]
	if !ok {
		return nil, ErrBadRouting
	}
	id, err := uuid.FromString(sID)
	if err != nil {
		return nil, ErrBadRouting
	}
	return findCustomerRequest{ID: id}, nil
}

func decodeUpdateCustomerRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	sID, ok := vars["userId"]
	if !ok {
		return nil, ErrBadRouting
	}
	var req updateCustomerRequest
	if e := json.NewDecoder(r.Body).Decode(&req.userDetails); e != nil && e != io.EOF {
		return nil, e
	}
	id, err := uuid.FromString(sID)
	if err != nil {
		return nil, ErrBadRouting
	}
	req.ID = id
	if req.Email == "" {
		return nil, errors.WithMessage(ErrBadRequest, "missing required parameter 'email'")
	}
	return req, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if response == nil {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeErrorResponse(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var errorResponse = translateError(err)
	w.WriteHeader(errorResponse.Status)
	_ = json.NewEncoder(w).Encode(errorResponse.Response)
}

type transportError struct {
	Status   int
	Response errorResponse
}

func translateError(err error) transportError {
	if errors.Is(err, ErrBadRequest) {
		return transportError{
			Status: http.StatusBadRequest,
			Response: errorResponse{
				Code:    101,
				Message: err.Error(),
			},
		}
	}
	switch err {
	case application.ErrCustomerNotFound:
		return transportError{
			Status: http.StatusNotFound,
			Response: errorResponse{
				Code:    102,
				Message: err.Error(),
			},
		}
	case application.ErrDuplicateUser:
		return transportError{
			Status: http.StatusConflict,
			Response: errorResponse{
				Code:    103,
				Message: err.Error(),
			},
		}
	case ErrNotAuthenticated:
		return transportError{
			Status: http.StatusUnauthorized,
			Response: errorResponse{
				Code:    104,
				Message: err.Error(),
			},
		}
	case application.ErrNotAuthorized:
		return transportError{
			Status: http.StatusForbidden,
			Response: errorResponse{
				Code:    105,
				Message: err.Error(),
			},
		}
	default:
		return transportError{
			Status: http.StatusInternalServerError,
			Response: errorResponse{
				Code:    100,
				Message: "unexpected error",
			},
		}
	}
}
