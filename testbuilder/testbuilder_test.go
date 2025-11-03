package testbuilder

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestCase_WithStateBuilder(t *testing.T) {
	t.Parallel()
	// Arrange
	testcase := &TestCase[string, string, func()]{}

	// Act
	res := testcase.
		WithStateBuilder(func(t *testing.T, sut *string, state *string) {
			t.Helper()

			*sut = "sut"
			*state = "state"
		})

	// Assert
	assert.Equal(t, testcase, res) // pointer equal
	require.NotNil(t, testcase.StateBuilder)

	var sut string

	var state string

	testcase.StateBuilder(t, &sut, &state)
	assert.Equal(t, "sut", sut)
	assert.Equal(t, "state", state)
}

func TestTestCase_WithSpecificBuilder(t *testing.T) {
	t.Parallel()
	// Arrange
	testcase := &TestCase[string, string, func()]{}

	// Act
	res := testcase.
		WithSpecificBuilder(func(t *testing.T, sut *string, state *string) {
			t.Helper()

			*sut = "sut"
			*state = "state"
		})

	// Assert
	assert.Equal(t, testcase, res) // pointer equal
	require.NotNil(t, testcase.SpecificBuilder)

	var (
		sut   string
		state string
	)

	testcase.SpecificBuilder(t, &sut, &state)

	assert.Equal(t, "sut", sut)
	assert.Equal(t, "state", state)
}

func TestTestCase_WithAssertion(t *testing.T) {
	t.Parallel()
	// Arrange
	testcase := &TestCase[string, string, func(t *testing.T, sut string, state string)]{}

	// Act
	res := testcase.
		WithAssertion(func(t *testing.T, sut string, state string) {
			t.Helper()

			assert.Equal(t, "sut", sut)
			assert.Equal(t, "state", state)
		})

	// Assert
	assert.Equal(t, testcase, res) // pointer equal
	require.NotNil(t, testcase.WithAssertion)
	testcase.Assertion(t, "sut", "state")
}

func TestTestsBuilder_Register(t *testing.T) {
	t.Parallel()
	// Arrange
	builder := TestsBuilder[string, string, func(t *testing.T, sut string, state string)]{}

	// Act
	res := builder.Register("test").
		WithStateBuilder(func(t *testing.T, sut *string, state *string) {
			t.Helper()

			*sut = "sut"
			*state = "state"
		}).
		WithSpecificBuilder(func(t *testing.T, sut *string, state *string) {
			t.Helper()

			*sut = "specific-sut"
			*state = "specific-state"
		}).
		WithAssertion(func(t *testing.T, sut string, state string) {
			t.Helper()

			assert.Equal(t, "specific-sut", sut)
			assert.Equal(t, "specific-state", state)
		})

	// Assert
	require.Len(t, builder.TestCases, 1)
	assert.Equal(t, builder.TestCases[0], res)
}

func TestTestsBuilder_Tests_NoTests(t *testing.T) {
	t.Parallel()
	// Arrange
	builder := TestsBuilder[string, string, func(t *testing.T, sut string, state string)]{}

	// Act
	for testName := range builder.Tests() {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			// Arrange
			// nothing to do
		})
	}
}

func TestTestsBuilder_Tests_StopDuringYield(t *testing.T) {
	t.Parallel()
	// Arrange
	builder := TestsBuilder[string, string, func(t *testing.T, sut string, state string)]{}
	builder.Register("test1")
	builder.Register("test2")

	// Act
	for testName := range builder.Tests() {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			// nothing to do
		})

		return
	}
}

func TestTestsBuilder_SingleTest_SpecificBuilderTakesPrecedence(t *testing.T) {
	t.Parallel()
	// Arrange
	builder := TestsBuilder[string, string, func(t *testing.T, sut string, state string)]{}
	builder.Register("test").
		WithStateBuilder(func(t *testing.T, sut *string, state *string) {
			t.Helper()

			*sut = "sut"
			*state = "state"
		}).
		WithSpecificBuilder(func(t *testing.T, sut *string, state *string) {
			t.Helper()

			*sut = "specific-sut"
			*state = "specific-state"
		})

	for testName, testBuilder := range builder.Tests() {
		assert.Equal(t, "test", testName)
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			// Arrange
			testData := testBuilder(t)

			// Assert
			assert.Equal(t, "specific-sut", testData.SUT)
			assert.Equal(t, "specific-state", testData.State)
		})
	}
}

func TestTestsBuilder_MultipleTests_StateIsRepeated(t *testing.T) {
	t.Parallel()
	// Arrange
	builder := TestsBuilder[string, int, func(t *testing.T)]{}
	builder.Register("1").
		WithStateBuilder(func(t *testing.T, _ *string, state *int) {
			t.Helper()

			*state += 1
		})
	builder.Register("2").
		WithStateBuilder(func(t *testing.T, _ *string, state *int) {
			t.Helper()

			*state += 1
		})
	builder.Register("3").
		WithStateBuilder(func(t *testing.T, _ *string, state *int) {
			t.Helper()

			*state += 1
		})
	builder.Register("4").
		WithStateBuilder(func(t *testing.T, _ *string, state *int) {
			t.Helper()

			*state += 1
		})

	for testName, testBuilder := range builder.Tests() {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			// Arrange
			testData := testBuilder(t)
			counter, err := strconv.Atoi(testName) // using testName because of t.Parallel()
			require.NoError(t, err)

			// Assert
			assert.Equal(t, counter, testData.State)
		})
	}
}

func TestTestsBuilder_MultipleTests_PreviousSpecificTestsAreIgnored(t *testing.T) {
	t.Parallel()
	// Arrange
	builder := TestsBuilder[string, int, func(t *testing.T)]{}
	builder.Register("1").
		WithStateBuilder(func(t *testing.T, sut *string, state *int) {
			t.Helper()

			*sut = "a"
		})
	builder.Register("2").
		WithStateBuilder(func(t *testing.T, sut *string, state *int) {
			t.Helper()

			*sut += "b"
		})
	builder.Register("3").
		WithStateBuilder(func(t *testing.T, sut *string, state *int) {
			t.Helper()

			*sut += "c"
		}).
		WithSpecificBuilder(func(t *testing.T, sut *string, state *int) {
			t.Helper()

			*sut = "-" // this should not be visible in test4
		})
	builder.Register("4").
		WithStateBuilder(func(t *testing.T, sut *string, state *int) {
			t.Helper()

			*sut += "d"
		})

	results := map[string]string{
		"1": "a",
		"2": "ab",
		"3": "-",
		"4": "abcd",
	}

	for testName, testBuilder := range builder.Tests() {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()
			// Arrange
			testData := testBuilder(t)

			// Assert
			assert.Equal(t, results[testName], testData.SUT)
		})
	}
}
