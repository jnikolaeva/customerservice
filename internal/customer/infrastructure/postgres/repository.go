package postgres

import (
	"github.com/jackc/pgx"
	"github.com/jnikolaeva/eshop-common/uuid"
	"github.com/pkg/errors"

	"github.com/jnikolaeva/customerservice/internal/customer/application"
)

const errUniqueConstraint = "23505"

type rawCustomer struct {
	ID        string `db:"id"`
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Email     string `db:"email"`
	Phone     string `db:"phone"`
}

type repository struct {
	connPool *pgx.ConnPool
}

func New(connPool *pgx.ConnPool) application.Repository {
	return &repository{
		connPool: connPool,
	}
}

func (r *repository) Add(customer application.Customer) error {
	_, err := r.connPool.Exec(
		"INSERT INTO customers (id, first_name, last_name, email, phone) VALUES ($1, $2, $3, $4, $5)",
		customer.ID.String(), customer.FirstName, customer.LastName, customer.Email, customer.Phone)
	return r.convertError(err)
}

func (r *repository) FindByID(id application.CustomerID) (*application.Customer, error) {
	var raw rawCustomer
	query := "SELECT id, first_name, last_name, phone, email FROM customers WHERE id = $1"
	err := r.connPool.QueryRow(query, id.String()).Scan(&raw.ID, &raw.FirstName, &raw.LastName, &raw.Phone, &raw.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			err = application.ErrCustomerNotFound
		}
		return nil, errors.WithStack(err)
	}
	customerID, _ := uuid.FromString(raw.ID)
	customer := &application.Customer{
		ID:        application.CustomerID(customerID),
		FirstName: raw.FirstName,
		LastName:  raw.LastName,
		Email:     raw.Email,
		Phone:     raw.Phone,
	}
	return customer, nil
}

func (r *repository) Update(user application.Customer) error {
	_, err := r.connPool.Exec(
		"UPDATE customers SET first_name = $1, last_name = $2, email = $3, phone = $4 WHERE id = $5",
		user.FirstName, user.LastName, user.Email, user.Phone, user.ID.String())
	return r.convertError(err)
}

func (r *repository) convertError(err error) error {
	if err != nil {
		pgErr, ok := err.(pgx.PgError)
		if ok && pgErr.Code == errUniqueConstraint {
			return application.ErrDuplicateUser
		}
		return errors.WithStack(err)
	}
	return nil
}
