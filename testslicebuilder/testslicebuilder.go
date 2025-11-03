package testslicebuilder

import (
	"errors"
	"testing"

	"github.com/Emptyless/go-testbuilder/testbuilder"
)

type TableTestItem[SUT any, STATE any, ASSERT any] struct {
	Name            string
	StateBuilder    func(t *testing.T, sut *SUT, state *STATE)
	SpecificBuilder func(t *testing.T, sut *SUT, state *STATE)
	Assertion       ASSERT
}

// Sentinel errors for clarity and better testability
var (
	ErrIndexOutOfRange = errors.New("index out of range")
	ErrNoTestsDefined  = errors.New("no tests defined")
)

func TestDataFromSlice[SUT any, STATE any, ASSERT any](
	t *testing.T,
	testIndex int,
	tests []TableTestItem[SUT, STATE, ASSERT],
) (testbuilder.TestData[SUT, STATE, ASSERT], error) {
	var sut SUT

	var state STATE

	if len(tests) == 0 {
		return testbuilder.TestData[SUT, STATE, ASSERT]{}, ErrNoTestsDefined
	}

	if testIndex < 0 || testIndex >= len(tests) {
		return testbuilder.TestData[SUT, STATE, ASSERT]{}, ErrIndexOutOfRange
	}

	// Build up to the index
	for _, tc := range tests[:testIndex+1] {
		if tc.StateBuilder != nil {
			tc.StateBuilder(t, &sut, &state)
		}
	}

	// Then run the specific builder at that index
	target := tests[testIndex]
	if target.SpecificBuilder != nil {
		target.SpecificBuilder(t, &sut, &state)
	}

	return testbuilder.TestData[SUT, STATE, ASSERT]{
		SUT:    sut,
		State:  state,
		Assert: target.Assertion,
	}, nil
}
