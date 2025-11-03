package examples

import (
	"testing"

	"github.com/Emptyless/go-testbuilder/testbuilder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserController_Handle(t *testing.T) {
	t.Parallel()

	// Mocks object
	type Mocks struct {
		MockMailer     *MockMailService
		MockRepository *MockUserRepository
	}

	// State object
	type State struct {
		// Inputs
		userName string
		payload  string

		// Mocks
		mocks Mocks

		// Returned user
		user User
	}

	type Sut = UserController

	type Assert = func(t *testing.T, controller UserController, state State, user *User, err error)

	// builder
	builder := testbuilder.TestsBuilder[Sut, State, Assert]{}

	builder.Register("invalid payload").
		WithSpecificBuilder(func(t *testing.T, sut *UserController, state *State) {
			state.payload = ""
		}).
		WithAssertion(func(t *testing.T, _ UserController, state State, user *User, err error) {
			assert.Nil(t, user)
			require.EqualError(t, err, "invalid payload")
		})

	builder.Register("get user failure").
		WithStateBuilder(func(t *testing.T, sut *UserController, state *State) {
			ctrl := gomock.NewController(t)

			state.userName = "my-user"
			state.payload = "my-payload"
			state.user = User{Name: state.userName}

			state.mocks.MockMailer = NewMockMailService(ctrl)
			state.mocks.MockRepository = NewMockUserRepository(ctrl)
		}).
		WithSpecificBuilder(func(t *testing.T, sut *UserController, state *State) {
			state.mocks.MockRepository.EXPECT().GetUser(state.userName).Return(User{}, assert.AnError)
		}).
		WithAssertion(func(t *testing.T, controller UserController, state State, user *User, err error) {
			assert.Nil(t, user)
			require.ErrorIs(t, assert.AnError, err)
		})

	builder.Register("send mail failure").
		WithStateBuilder(func(t *testing.T, sut *UserController, state *State) {
			state.mocks.MockRepository.EXPECT().GetUser(state.userName).Return(state.user, nil) // Now this needs to succeed to continue to the next test
		}).
		WithSpecificBuilder(func(t *testing.T, sut *UserController, state *State) {
			state.mocks.MockMailer.EXPECT().SendMail().Return(assert.AnError)
		}).
		WithAssertion(func(t *testing.T, controller UserController, state State, user *User, err error) {
			assert.Nil(t, user)
			require.ErrorIs(t, assert.AnError, err)
		})

	builder.Register("store user failure").
		WithStateBuilder(func(t *testing.T, sut *UserController, state *State) {
			state.mocks.MockMailer.EXPECT().SendMail().Return(nil)
		}).
		WithSpecificBuilder(func(t *testing.T, sut *UserController, state *State) {
			state.mocks.MockRepository.EXPECT().StoreUser(state.user).Return(assert.AnError)
		}).
		WithAssertion(func(t *testing.T, controller UserController, state State, user *User, err error) {
			assert.Nil(t, user)
			require.EqualError(t, err, assert.AnError.Error())
		})

	builder.Register("success").
		WithStateBuilder(func(t *testing.T, sut *UserController, state *State) {
			state.mocks.MockRepository.EXPECT().StoreUser(state.user).Return(nil)
		}).
		WithAssertion(func(t *testing.T, controller UserController, state State, user *User, err error) {
			require.NoError(t, err)
			assert.NotNil(t, user)
			assert.Equal(t, state.user, *user)
		})

	// Run all test cases
	for name, buildTest := range builder.Tests() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Arrange
			testData := buildTest(t)
			ctrl := testData.SUT

			// Use Mocks to populate actual interfaces
			ctrl.Mailer = testData.State.mocks.MockMailer
			ctrl.Repository = testData.State.mocks.MockRepository

			// Act
			user, err := ctrl.Handle(testData.State.userName, testData.State.payload)

			// Assert
			testData.Assert(t, ctrl, testData.State, user, err)
		})
	}
}
