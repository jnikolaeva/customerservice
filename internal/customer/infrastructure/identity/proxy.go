package identity

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jnikolaeva/eshop-common/uuid"
	"github.com/pkg/errors"
)

type Proxy struct {
	baseURL    string
	httpClient *http.Client
}

type registerUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type userResponse struct {
	ID string `json:"id"`
}

func NewProviderProxy(baseURL string) *Proxy {
	client := http.DefaultClient
	client.Timeout = 20 * time.Second
	return &Proxy{
		baseURL:    baseURL,
		httpClient: client,
	}
}

func (p *Proxy) Register(username, password string) (uuid.UUID, error) {
	registerURL := p.baseURL + "/users"
	request := &registerUserRequest{
		Username: username,
		Password: password,
	}
	data := new(bytes.Buffer)
	if err := json.NewEncoder(data).Encode(request); err != nil {
		return uuid.UUID{}, errors.Wrap(err, "failed to register user")
	}
	r, err := p.httpClient.Post(registerURL, "application/json", data)
	if err != nil {
		return uuid.UUID{}, errors.Wrap(err, "identity provider failed to register user")
	}
	if r.StatusCode != http.StatusOK {
		return uuid.UUID{}, errors.WithStack(errors.Errorf("identity provider failed to register user with status code: %d", r.StatusCode))
	}
	defer r.Body.Close()
	var user userResponse
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		return uuid.UUID{}, errors.Wrap(err, "failed to decode response from identity provider")
	}
	id, err := uuid.FromString(user.ID)
	if err != nil {
		return uuid.UUID{}, errors.Wrapf(err, "failed to convert user id from identity provider: %v", user.ID)
	}
	return id, nil
}

func (p *Proxy) Delete(userID uuid.UUID) error {
	deleteURL := p.baseURL + "/users/" + userID.String()
	req, err := http.NewRequest(http.MethodDelete, deleteURL, nil)
	if err != nil {
		return err
	}
	r, err := p.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to delete user")
	}
	if r.StatusCode != http.StatusNoContent {
		return errors.WithStack(errors.Errorf("identity provider failed to delete user with status code: %d", r.StatusCode))
	}
	return nil
}
