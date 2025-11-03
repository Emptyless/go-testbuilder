package examples

import (
	"errors"
)

//go:generate mockgen -source=./user_controller.go -destination=mocks.go -package=examples

// UserController An example system under test
type UserController struct {
	Mailer     MailService
	Repository UserRepository
}

var ErrInvalidPayload = errors.New("invalid payload")

func (c *UserController) Handle(userName string, payload string) (*User, error) {
	if payload == "" {
		return nil, ErrInvalidPayload
	}

	user, getUserErr := c.Repository.GetUser(userName)
	if getUserErr != nil {
		return nil, getUserErr // do something with the error
	}

	err := c.Mailer.SendMail()
	if err != nil {
		return nil, err // do something with the error
	}

	err = c.Repository.StoreUser(user)
	if err != nil {
		return nil, err // do something with the error
	}

	return &user, nil
}

type MailService interface {
	SendMail() error
}

type User struct {
	Name string
}

type UserRepository interface {
	GetUser(username string) (User, error)
	StoreUser(user User) error
}
