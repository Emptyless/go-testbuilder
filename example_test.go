package testbuilder

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// An example system under test
type UserController struct {
	Mailer     MailService
	Repository UserRepository
}

func (c *UserController) Handle(userName string, payload string) (*User, error) {
	if payload == "" {
		return nil, errors.New("invalid payload")
	}

	user, getUserErr := c.Repository.GetUser(userName)
	if getUserErr != nil {
		return nil, getUserErr // do something with the error
	}

	if err := c.Mailer.SendMail(); err != nil {
		return nil, err // do something with the error
	}

	if err := c.Repository.StoreUser(user); err != nil {
		return nil, err // do something with the error
	}

	return &user, nil
}

type MailService interface {
	SendMail() error
}

type MockMailService struct {
	Error error
}

func (m *MockMailService) SendMail() error {
	return m.Error
}

type User struct {
	Name string
}

type UserRepository interface {
	GetUser(string) (User, error)
	StoreUser(User) error
}

type MockUserRepository struct {
	GetUserProvidedUserName string
	GetUserUser             User
	GetUserError            error

	StoreUserProvidedUser User
	StoreUserError        error
}

func (m *MockUserRepository) GetUser(s string) (User, error) {
	m.GetUserProvidedUserName = s
	return m.GetUserUser, m.GetUserError
}

func (m *MockUserRepository) StoreUser(user User) error {
	m.StoreUserProvidedUser = user
	return m.StoreUserError
}

func TestUserController_Handle(t *testing.T) {
	t.Parallel()

	// State object
	type State struct {
		// Inputs
		userName string
		payload  string

		// Returned user
		user User
	}

	// builder
	builder := TestsBuilder[UserController, State, func(t *testing.T, controller UserController, state State, user *User, error error)]{}

	builder.Register("invalid payload").WithSpecificBuilder(func(t *testing.T, sut *UserController, state *State) {
		state.payload = ""
	}).WithAssertion(func(t *testing.T, controller UserController, state State, user *User, error error) {
		assert.Nil(t, user)
		require.EqualError(t, error, "invalid payload")
	})

	builder.Register("get user failure").WithStateBuilder(func(t *testing.T, sut *UserController, state *State) {
		state.userName = "my-user"
		state.payload = "my-payload"
		sut.Repository = &MockUserRepository{}
	}).WithSpecificBuilder(func(t *testing.T, sut *UserController, state *State) {
		sut.Repository.(*MockUserRepository).GetUserError = assert.AnError
	}).WithAssertion(func(t *testing.T, controller UserController, state State, user *User, error error) {
		assert.Nil(t, user)
		assert.Equal(t, controller.Repository.(*MockUserRepository).GetUserProvidedUserName, state.userName) // typically something like this would be easier with go-mock
		require.EqualError(t, error, assert.AnError.Error())
	})

	builder.Register("send mail failure").WithStateBuilder(func(t *testing.T, sut *UserController, state *State) {
		state.user = User{Name: state.userName}
		sut.Repository.(*MockUserRepository).GetUserUser = state.user
		sut.Repository.(*MockUserRepository).GetUserError = nil // is already nil, just added for verbosity
	}).WithSpecificBuilder(func(t *testing.T, sut *UserController, state *State) {
		sut.Mailer = &MockMailService{Error: assert.AnError}
	}).WithAssertion(func(t *testing.T, controller UserController, state State, user *User, error error) {
		assert.Nil(t, user)
		require.EqualError(t, error, assert.AnError.Error())
	})

	builder.Register("store user failure").WithStateBuilder(func(t *testing.T, sut *UserController, state *State) {
		sut.Mailer = &MockMailService{Error: nil}
	}).WithSpecificBuilder(func(t *testing.T, sut *UserController, state *State) {
		sut.Repository.(*MockUserRepository).StoreUserError = assert.AnError
	}).WithAssertion(func(t *testing.T, controller UserController, state State, user *User, error error) {
		assert.Nil(t, user)
		require.EqualError(t, error, assert.AnError.Error())
	})

	builder.Register("success").WithStateBuilder(func(t *testing.T, sut *UserController, state *State) {
		sut.Repository.(*MockUserRepository).StoreUserError = nil
	}).WithAssertion(func(t *testing.T, controller UserController, state State, user *User, error error) {
		require.Nil(t, error)
		require.NotNil(t, user)
		assert.Equal(t, state.user, *user)
	})

	// Run all test cases
	for name, buildTest := range builder.Tests() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Arrange
			testData := buildTest(t)
			ctrl := testData.SUT

			// Act
			user, err := ctrl.Handle(testData.State.userName, testData.State.payload)

			// Assert
			testData.Assert(t, ctrl, testData.State, user, err)
		})
	}
}
