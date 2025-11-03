package examples

import (
	"testing"

	"github.com/Emptyless/go-testbuilder/testslicebuilder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserController_TableTest_Handle(t *testing.T) {
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

	tests := []testslicebuilder.TableTestItem[Sut, State, Assert]{
		{
			Name:         "invalid payload",
			StateBuilder: func(t *testing.T, sut *Sut, state *State) {},
			SpecificBuilder: func(t *testing.T, sut *Sut, state *State) {
				state.payload = ""
			},
			Assertion: func(t *testing.T, controller UserController, state State, user *User, err error) {
				assert.Nil(t, user)
				assert.EqualError(t, err, "invalid payload")
			},
		},
		{
			Name: "get user failure",
			StateBuilder: func(t *testing.T, sut *Sut, state *State) {
				ctrl := gomock.NewController(t)

				state.userName = "my-user"
				state.payload = "my-payload"
				state.user = User{Name: state.userName}

				state.mocks.MockMailer = NewMockMailService(ctrl)
				state.mocks.MockRepository = NewMockUserRepository(ctrl)
			},
			SpecificBuilder: func(t *testing.T, sut *Sut, state *State) {
				state.mocks.MockRepository.EXPECT().GetUser(state.userName).Return(User{}, assert.AnError)
			},
			Assertion: func(t *testing.T, controller UserController, state State, user *User, err error) {
				assert.Nil(t, user)
				require.ErrorIs(t, assert.AnError, err)
			},
		},
		{
			Name: "send mail failure",
			StateBuilder: func(t *testing.T, sut *Sut, state *State) {
				// Now this needs to succeed to continue to the next test
				state.mocks.MockRepository.EXPECT().GetUser(state.userName).Return(state.user, nil)
			},
			SpecificBuilder: func(t *testing.T, sut *Sut, state *State) {
				state.mocks.MockMailer.EXPECT().SendMail().Return(assert.AnError)
			},
			Assertion: func(t *testing.T, controller UserController, state State, user *User, err error) {
				assert.Nil(t, user)
				require.ErrorIs(t, assert.AnError, err)
			},
		},
		{
			Name: "store user failure",
			StateBuilder: func(t *testing.T, sut *Sut, state *State) {
				state.mocks.MockMailer.EXPECT().SendMail().Return(nil)
			},
			SpecificBuilder: func(t *testing.T, sut *Sut, state *State) {
				state.mocks.MockRepository.EXPECT().StoreUser(state.user).Return(assert.AnError)
			},
			Assertion: func(t *testing.T, controller UserController, state State, user *User, err error) {
				assert.Nil(t, user)
				require.ErrorIs(t, assert.AnError, err)
			},
		},
		{
			Name: "success",
			StateBuilder: func(t *testing.T, sut *Sut, state *State) {
				state.mocks.MockRepository.EXPECT().StoreUser(state.user).Return(nil)
			},
			SpecificBuilder: func(t *testing.T, sut *Sut, state *State) {},
			Assertion: func(t *testing.T, controller UserController, state State, user *User, err error) {
				require.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, state.user, *user)
			},
		},
	}

	// To enable Goland's/Intellij's integration, the first use of the slice 'tests' must be in this form:
	for i, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Arrange
			testData, err := testslicebuilder.TestDataFromSlice(t, i, tests)
			require.NoError(t, err)

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
