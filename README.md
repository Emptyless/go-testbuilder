# Go TestBuilder

_⚠️ This tool is under active development and must be considered alpha. It's API may be changed in a breaking way until a 1.0 version is released. Submit issues to the Github issue tracker if found.⚠️_

A workflow like `TestsBuilder` that uses generics for type-safety. The aim of this library is to make it easier to test

- large more use-case oriented functions
- ... that don't necessarily have a high branch complexity
- ... but do have a lot of methods

And that without repetition!

## Installation

```bash
go get github.com/Emptyless/go-testbuilder
```

## Usage

```go
package main

import (
	"errors"
	"github.com/Emptyless/go-testbuilder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
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
	builder := testbuilder.TestsBuilder[UserController, State, func(t *testing.T, controller UserController, state State, user *User, error error)]{}

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

```

## How It Works

TestBuilder manages test cases with three generic types:

1. **SUT** (System Under Test): The component being tested
2. **STATE**: The test state that can be shared and modified across test cases
3. **ASSERT**: The assertion logic, typically a function `func(t *testing.T, ...)`

When iterating through test cases:

1. A clean SUT and STATE are initialized before each test
2. State builders from all previous test cases are applied in order
3. The specific builder for the current test case is applied
4. The test's assertion logic is executed

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.