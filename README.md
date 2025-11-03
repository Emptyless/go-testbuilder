# Go TestBuilder

_⚠️ This tool is under active development and must be considered alpha. It's API may be changed in a breaking way until a 1.0 version is released. Submit issues to the Github issue tracker if found.⚠️_

A workflow like `TestsBuilder` that uses generics for type-safety. The aim of this library is to make it easier to test

- large more use-case oriented functions
- ... that don't necessarily have a high branch complexity
- ... but do have a lot of methods

And that without repetition!


---

## Usage

### The Problem

When testing complex functions with many dependencies or steps, traditional table-driven tests often lead to repetition. Each test case must set up mocks and state independently, even when many setup steps are shared.

For example, if you're testing a workflow with 5 steps, your second test might need to:
- Set up the same mocks as test 1
- Configure them identically up to step 2
- Then diverge with a specific failure scenario for step 2

### The Solution: Builder Pattern with State Inheritance

Go TestBuilder solves this by allowing test cases to **inherit setup from previous tests**. Each test builds on the previous one's state.

When you register test cases, they form a dependency chain:

```
Test 0 (invalid payload):
  Setup: specificBuilder0
  Run & Assert

Test 1 (get user failure):
  Setup: stateBuilder0 + specificBuilder1
  Run & Assert

Test 2 (send mail failure):
  Setup: stateBuilder0 + stateBuilder1 + specificBuilder2
  Run & Assert
```

Each test runs with:
1. A fresh SUT (System Under Test)
2. A fresh initial state
3. All previous tests' **StateBuilder** functions applied in order
4. The current test's **SpecificBuilder** applied
5. Finally, the test's assertion logic

This eliminates duplication. Setup logic is written once and reused by dependent tests.

---

### Execution Model Explained

Consider five test cases in order:

```
Test 0 - invalid payload
Test 1 - get user failure
Test 2 - send mail failure
Test 3 - store user failure
Test 4 - success
```

Their builder execution would be:

```
Running Test 0:
  specificBuilder0
  Assertion

Running Test 1:
  stateBuilder0
  specificBuilder1
  Assertion

Running Test 2:
  stateBuilder0
  stateBuilder1
  specificBuilder2
  Assertion

Running Test 3:
  stateBuilder0
  stateBuilder1
  stateBuilder2
  specificBuilder3
  Assertion

Running Test 4:
  stateBuilder0
  stateBuilder1
  stateBuilder2
  stateBuilder3
  specificBuilder4
  Assertion
```

Each test **inherits accumulated setup** from all tests before it. This allows you to build progressive test scenarios where later tests assume earlier setup is complete.


---

### Two APIs: Same Behavior, Different Syntax

This library provides **two equivalent ways** to define tests:

1. **Fluent Builder API** (`testbuilder.TestsBuilder`) — chain method calls for a concise, expressive style
2. **Declarative Table Style** (`testslicebuilder.TableTestItem`) — define tests as struct entries in a slice for clarity

Both execute identically. Choose the one that fits your team's preferences and codebase style.

**Note:** The table-driven style is preferred when working with JetBrains IDEs (GoLand, IntelliJ with Go plugin), as it provides individual play buttons for each test case in the editor.

---

## Example System Under Test

For the following examples, we'll test this `UserController`:

```go
package examples

import (
	"errors"
)

type UserController struct {
	Mailer     MailService
	Repository UserRepository
}

var ErrInvalidPayload = errors.New("invalid payload")

func (c *UserController) Handle(userName string, payload string) (*User, error) {
	if payload == "" {
		return nil, ErrInvalidPayload
	}

	user, err := c.Repository.GetUser(userName)
	if err != nil {
		return nil, err
	}

	if err := c.Mailer.SendMail(); err != nil {
		return nil, err
	}

	if err := c.Repository.StoreUser(user); err != nil {
		return nil, err
	}

	return &user, nil
}

type MailService interface {
	SendMail() error
}

type UserRepository interface {
	GetUser(username string) (User, error)
	StoreUser(user User) error
}

type User struct {
	Name string
}
```

---

## Example 1: Fluent Builder API

This style chains method calls for a concise, expressive structure. Great when you want to group related setup and assertions together.

```go
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

	// Define supporting types for mocks and state
	type Mocks struct {
		MockMailer     *MockMailService
		MockRepository *MockUserRepository
	}

	type State struct {
		userName string
		payload  string
		mocks    Mocks
		user     User
	}

	type Sut = UserController
	type Assert = func(t *testing.T, controller UserController, state State, user *User, err error)

	// Create the builder
	builder := testbuilder.TestsBuilder[Sut, State, Assert]{}

	// Register test cases
	builder.Register("invalid payload").
		WithSpecificBuilder(func(t *testing.T, sut *Sut, state *State) {
			state.payload = ""
		}).
		WithAssertion(func(t *testing.T, _ UserController, state State, user *User, err error) {
			assert.Nil(t, user)
			require.EqualError(t, err, "invalid payload")
		})

	builder.Register("get user failure").
		WithStateBuilder(func(t *testing.T, sut *Sut, state *State) {
			ctrl := gomock.NewController(t)
			state.userName = "my-user"
			state.payload = "my-payload"
			state.user = User{Name: state.userName}
			state.mocks.MockMailer = NewMockMailService(ctrl)
			state.mocks.MockRepository = NewMockUserRepository(ctrl)
		}).
		WithSpecificBuilder(func(t *testing.T, sut *Sut, state *State) {
			state.mocks.MockRepository.EXPECT().GetUser(state.userName).
				Return(User{}, assert.AnError)
		}).
		WithAssertion(func(t *testing.T, controller UserController, state State, user *User, err error) {
			assert.Nil(t, user)
			require.ErrorIs(t, assert.AnError, err)
		})

	builder.Register("send mail failure").
		WithStateBuilder(func(t *testing.T, sut *Sut, state *State) {
			state.mocks.MockRepository.EXPECT().GetUser(state.userName).
				Return(state.user, nil)
		}).
		WithSpecificBuilder(func(t *testing.T, sut *Sut, state *State) {
			state.mocks.MockMailer.EXPECT().SendMail().Return(assert.AnError)
		}).
		WithAssertion(func(t *testing.T, controller UserController, state State, user *User, err error) {
			assert.Nil(t, user)
			require.ErrorIs(t, assert.AnError, err)
		})

	builder.Register("store user failure").
		WithStateBuilder(func(t *testing.T, sut *Sut, state *State) {
			state.mocks.MockMailer.EXPECT().SendMail().Return(nil)
		}).
		WithSpecificBuilder(func(t *testing.T, sut *Sut, state *State) {
			state.mocks.MockRepository.EXPECT().StoreUser(state.user).
				Return(assert.AnError)
		}).
		WithAssertion(func(t *testing.T, controller UserController, state State, user *User, err error) {
			assert.Nil(t, user)
			require.ErrorIs(t, assert.AnError, err)
		})

	builder.Register("success").
		WithStateBuilder(func(t *testing.T, sut *Sut, state *State) {
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

			// Arrange: build fresh test data (applies builder chain)
			testData := buildTest(t)
			ctrl := testData.SUT

			// Wire mocks
			ctrl.Mailer = testData.State.mocks.MockMailer
			ctrl.Repository = testData.State.mocks.MockRepository

			// Act
			user, err := ctrl.Handle(testData.State.userName, testData.State.payload)

			// Assert
			testData.Assert(t, ctrl, testData.State, user, err)
		})
	}
}
```

---

## Example 2: Declarative Table Style

This style expresses each test case as a struct entry in a slice. It's easier to visually scan all test cases at once, and IDEs will detect subtest names readily.

```go
package examples

import (
	"testing"

	"github.com/Emptyless/go-testbuilder/testslicebuilder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserController_TableTest_Handle(t *testing.T) {
	type Mocks struct {
		MockMailer     *MockMailService
		MockRepository *MockUserRepository
	}

	type State struct {
		userName string
		payload  string
		mocks    Mocks
		user     User
	}

	type Sut = UserController
	type Assert = func(t *testing.T, controller UserController, state State, user *User, err error)

	tests := []testslicebuilder.TableTestItem[Sut, State, Assert]{
		{
			Name: "invalid payload",
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
				state.mocks.MockRepository.EXPECT().GetUser(state.userName).
					Return(User{}, assert.AnError)
			},
			Assertion: func(t *testing.T, controller UserController, state State, user *User, err error) {
				assert.Nil(t, user)
				require.ErrorIs(t, assert.AnError, err)
			},
		},
		{
			Name: "send mail failure",
			StateBuilder: func(t *testing.T, sut *Sut, state *State) {
				state.mocks.MockRepository.EXPECT().GetUser(state.userName).
					Return(state.user, nil)
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
				state.mocks.MockRepository.EXPECT().StoreUser(state.user).
					Return(assert.AnError)
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
			Assertion: func(t *testing.T, controller UserController, state State, user *User, err error) {
				require.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, state.user, *user)
			},
		},
	}

	// Run all test cases
	for i, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			// Arrange: build test data from slice (applies builder chain)
			testData, err := testslicebuilder.TestDataFromSlice(t, i, tests)
			require.NoError(t, err)

			ctrl := testData.SUT

			// Wire mocks
			ctrl.Mailer = testData.State.mocks.MockMailer
			ctrl.Repository = testData.State.mocks.MockRepository

			// Act
			user, err := ctrl.Handle(testData.State.userName, testData.State.payload)

			// Assert
			testData.Assert(t, ctrl, testData.State, user, err)
		})
	}
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.