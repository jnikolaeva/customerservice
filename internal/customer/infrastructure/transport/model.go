package transport

import "github.com/jnikolaeva/eshop-common/uuid"

type registerCustomerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	userDetails
}

type registerCustomerResponse struct {
	ID string `json:"id"`
}

type findCustomerRequest struct {
	ID uuid.UUID `json:"userId"`
}

type findCustomerResponse struct {
	userData
}

type updateCustomerRequest struct {
	ID uuid.UUID `json:"userId"`
	userDetails
}

type updateCustomerResponse struct {
	userData
}

type errorResponse struct {
	Code    uint32 `json:"code"`
	Message string `json:"message"`
}

type userData struct {
	ID string `json:"id"`
	userDetails
}

type userDetails struct {
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
}
