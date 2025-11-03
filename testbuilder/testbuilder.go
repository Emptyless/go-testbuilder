package testbuilder

import (
	"iter"
	"testing"
)

// TestsBuilder manages a collection of test cases for a system under test (SUT).
// SUT represents the system under test
// STATE represents the test state
// ASSERT defines the test assertions.
//
// After Register'ing all tests, range over the Tests iterator
//
// TestCase's contain a TestCase.StateBuilder which is the primary feature of the TestBuilder to manage. While ranging
// over the tests, a clean SUT and STATE are initialized before each test. Then the SUT and STATE are modified by
// - First TestCase: TestCase[0].SpecificBuilder(TestCase[0].StateBuilder(SUT, STATE))
// - Second TestCase: TestCase[1].SpecificBuilder(TestCase[0..1].StateBuilder(SUT, STATE))
// - Third TestCase: TestCase[2].SpecificBuilder(TestCase[0..2].StateBuilder(SUT, STATE))
// - ...
// - Nth TestCase: TestCase[n].SpecificBuilder(TestCase[0..n].StateBuilder(SUT, STATE))
type TestsBuilder[SUT any, STATE any, ASSERT any] struct {
	TestCases []*TestCase[SUT, STATE, ASSERT]
}

// TestData defines a generic structure for test data, including the system under test, state, and assertion logic.
// SUT represents the system under test.
// STATE represents the test's initial or modified state.
// ASSERT contains the logic or value used for verification.
type TestData[SUT any, STATE any, ASSERT any] struct {
	// SUT represents the system under test within a test case.
	SUT SUT

	// State tracks the state produced by the TestCase.StateBuilder's
	State STATE

	// Assert function that can be specified to be any type. Typically, it is a good idea to use a function signature
	// like func(t *testing.T, state STATE, ...) where the ... is replaced by the output of the SUT
	Assert ASSERT
}

// TestCase is yielded to the TestsBuilder.Tests range loop. See TestsBuilder for documentation on the types
type TestCase[SUT any, STATE any, ASSERT any] struct {
	// TestName for the test case
	TestName string
	// StateBuilder that is subsequently used to build up state for the tests. The distinction between the StateBuilder
	// and the SpecificBuilder is that StateBuilder is subsequently called for all TestCase's that are registered to the
	// TestsBuilder.
	StateBuilder func(t *testing.T, sut *SUT, state *STATE)
	// SpecificBuilder is only run for this case
	SpecificBuilder func(t *testing.T, sut *SUT, state *STATE)
	// Assertion logic
	Assertion ASSERT
}

// WithStateBuilder mutates the SUT and STATE for the current and all further tests
func (ts *TestCase[SUT, STATE, ASSERT]) WithStateBuilder(f func(t *testing.T, sut *SUT, state *STATE)) *TestCase[SUT, STATE, ASSERT] {
	ts.StateBuilder = f
	return ts
}

// WithSpecificBuilder mutates the SUT and STATE only for this particular test
func (ts *TestCase[SUT, STATE, ASSERT]) WithSpecificBuilder(f func(t *testing.T, sut *SUT, state *STATE)) *TestCase[SUT, STATE, ASSERT] {
	ts.SpecificBuilder = f
	return ts
}

// WithAssertion holds any assertion logic associated with this TestCase
func (ts *TestCase[SUT, STATE, ASSERT]) WithAssertion(f ASSERT) *TestCase[SUT, STATE, ASSERT] {
	ts.Assertion = f
	return ts
}

// Register the test to the TestsBuilder
func (ts *TestsBuilder[SUT, STATE, ASSERT]) Register(name string) *TestCase[SUT, STATE, ASSERT] {
	testcase := &TestCase[SUT, STATE, ASSERT]{
		TestName: name,
	}
	ts.TestCases = append(ts.TestCases, testcase)

	return testcase
}

// Tests iterator that yields a TestName and TestData structure
//
// TestCase's contain a TestCase.StateBuilder which is the primary feature of the TestBuilder to manage. While ranging
// over the tests, a clean SUT and STATE are initialized before each test. Then the SUT and STATE are modified by
// - First TestCase: TestCase[0].SpecificBuilder(TestCase[0].StateBuilder(SUT, STATE))
// - Second TestCase: TestCase[1].SpecificBuilder(TestCase[0..1].StateBuilder(SUT, STATE))
// - Third TestCase: TestCase[2].SpecificBuilder(TestCase[0..2].StateBuilder(SUT, STATE))
// - ...
// - Nth TestCase: TestCase[n].SpecificBuilder(TestCase[0..n].StateBuilder(SUT, STATE))
func (ts *TestsBuilder[SUT, STATE, ASSERT]) Tests() iter.Seq2[string, func(t *testing.T) TestData[SUT, STATE, ASSERT]] {
	return func(yield func(string, func(t *testing.T) TestData[SUT, STATE, ASSERT]) bool) {
		for i, curcase := range ts.TestCases {
			build := func(t *testing.T) TestData[SUT, STATE, ASSERT] {
				t.Helper()

				var (
					sut   SUT
					state STATE
				)

				for j, testcase := range ts.TestCases {
					if builder := testcase.StateBuilder; builder != nil {
						builder(t, &sut, &state)
					}

					if j < i {
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
