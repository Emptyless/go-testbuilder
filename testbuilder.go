package testbuilder

import (
	"iter"
	"testing"
)

type TestsBuilder[SUT any, STATE any, ASSERT any] struct {
	TestCases []*TestCase[SUT, STATE, ASSERT]
}

type TestData[SUT any, STATE any, ASSERT any] struct {
	SUT    SUT
	State  STATE
	Assert ASSERT
}

type TestCase[SUT any, STATE any, ASSERT any] struct {
	// TestName for current test
	TestName string
	// StateBuilder that is subsequently used to build up state for the tests
	StateBuilder func(t *testing.T, sut *SUT, state *STATE)
	// SpecificBuilder is only run for this case
	SpecificBuilder func(t *testing.T, sut *SUT, state *STATE)
	// Assertion
	Assertion ASSERT
}

func (ts *TestCase[SUT, STATE, ASSERT]) WithStateBuilder(f func(t *testing.T, sut *SUT, state *STATE)) *TestCase[SUT, STATE, ASSERT] {
	ts.StateBuilder = f
	return ts
}

func (ts *TestCase[SUT, STATE, ASSERT]) WithSpecificBuilder(f func(t *testing.T, sut *SUT, state *STATE)) *TestCase[SUT, STATE, ASSERT] {
	ts.SpecificBuilder = f
	return ts
}

func (ts *TestCase[SUT, STATE, ASSERT]) WithAssertion(f ASSERT) *TestCase[SUT, STATE, ASSERT] {
	ts.Assertion = f
	return ts
}

func (ts *TestsBuilder[SUT, STATE, ASSERT]) Register(name string) *TestCase[SUT, STATE, ASSERT] {
	testcase := &TestCase[SUT, STATE, ASSERT]{
		TestName: name,
	}
	ts.TestCases = append(ts.TestCases, testcase)
	return testcase
}

func (ts *TestsBuilder[SUT, STATE, ASSERT]) Tests() iter.Seq2[string, func(t *testing.T) TestData[SUT, STATE, ASSERT]] {
	return func(yield func(string, func(t *testing.T) TestData[SUT, STATE, ASSERT]) bool) {
		for i, curcase := range ts.TestCases {
			build := func(t *testing.T) TestData[SUT, STATE, ASSERT] {
				var sut SUT
				var state STATE

				for j, testcase := range ts.TestCases {
					if builder := testcase.StateBuilder; builder != nil {
						builder(t, &sut, &state)
					}

					if i != j {
						continue
					}

					if testcase.SpecificBuilder != nil {
						testcase.SpecificBuilder(t, &sut, &state)
					}

					break
				}

				return TestData[SUT, STATE, ASSERT]{
					SUT:    sut,
					State:  state,
					Assert: curcase.Assertion,
				}
			}

			if !yield(curcase.TestName, build) {
				return
			}
		}
	}
}
