package testslicebuilder

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Dummy types for SUT, state, and assertion ---

type DummySUT struct {
	actualCalled []string
}

type DummyState struct {
	actualCalled []string
}

type DummyAssert struct {
	Name string
}

// ===============================================================
// == Helper append functions ==
// ===============================================================

func appendSUT(sut *DummySUT, label string) {
	sut.actualCalled = append(sut.actualCalled, "sut-"+label)
}

func appendState(state *DummyState, label string) {
	state.actualCalled = append(state.actualCalled, "state-"+label)
}

// ===============================================================
// == TESTS ==
// ===============================================================

func Test_TestDataFromSlice_FullBehavior(t *testing.T) {
	tests := []TableTestItem[DummySUT, DummyState, DummyAssert]{
		{
			Name: "test0",
			StateBuilder: func(t *testing.T, sut *DummySUT, state *DummyState) {
				t.Helper()

				appendSUT(sut, "statebuilder0")
				appendState(state, "statebuilder0")
			},
			SpecificBuilder: func(t *testing.T, sut *DummySUT, state *DummyState) {
				t.Helper()

				appendSUT(sut, "specificbuilder0")
				appendState(state, "specificbuilder0")
			},
			Assertion: DummyAssert{"assert0"},
		},
		{
			Name: "test1",
			StateBuilder: func(t *testing.T, sut *DummySUT, state *DummyState) {
				t.Helper()

				appendSUT(sut, "statebuilder1")
				appendState(state, "statebuilder1")
			},
			SpecificBuilder: func(t *testing.T, sut *DummySUT, state *DummyState) {
				t.Helper()

				appendSUT(sut, "specificbuilder1")
				appendState(state, "specificbuilder1")
			},
			Assertion: DummyAssert{"assert1"},
		},
		{
			Name: "test2",
			StateBuilder: func(t *testing.T, sut *DummySUT, state *DummyState) {
				t.Helper()

				appendSUT(sut, "statebuilder2")
				appendState(state, "statebuilder2")
			},
			SpecificBuilder: func(t *testing.T, sut *DummySUT, state *DummyState) {
				t.Helper()

				appendSUT(sut, "specificbuilder2")
				appendState(state, "specificbuilder2")
			},
			Assertion: DummyAssert{"assert2"},
		},
	}

	t.Run("run all indices like a normal for-loop", func(t *testing.T) {
		sutAllActualCalled := make([][]string, 0)
		stateAllActualCalled := make([][]string, 0)
		actualNames := make([]string, 0)

		expectedSUTAll := [][]string{
			{"sut-statebuilder0", "sut-specificbuilder0"},
			{"sut-statebuilder0", "sut-statebuilder1", "sut-specificbuilder1"},
			{"sut-statebuilder0", "sut-statebuilder1", "sut-statebuilder2", "sut-specificbuilder2"},
		}

		expectedStateAll := [][]string{
			{"state-statebuilder0", "state-specificbuilder0"},
			{"state-statebuilder0", "state-statebuilder1", "state-specificbuilder1"},
			{"state-statebuilder0", "state-statebuilder1", "state-statebuilder2", "state-specificbuilder2"},
		}

		expectedNames := []string{"test0", "test1", "test2"}

		for i := range tests {
			data, err := TestDataFromSlice(t, i, tests)
			require.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("assert%v", i), data.Assert.Name)
			assert.Equal(t, expectedNames[i], tests[i].Name)

			sutAllActualCalled = append(sutAllActualCalled, data.SUT.actualCalled)
			stateAllActualCalled = append(stateAllActualCalled, data.State.actualCalled)
			actualNames = append(actualNames, tests[i].Name)
		}

		assert.Equal(t, expectedSUTAll, sutAllActualCalled)
		assert.Equal(t, expectedStateAll, stateAllActualCalled)
		assert.Equal(t, expectedNames, actualNames)
	})
}

// ===============================================================

func Test_TestDataFromSlice_MissingBuilders(t *testing.T) {
	tests := []TableTestItem[DummySUT, DummyState, DummyAssert]{
		{
			Name:         "missingStateBuilder",
			StateBuilder: nil,
			SpecificBuilder: func(t *testing.T, sut *DummySUT, state *DummyState) {
				t.Helper()

				appendSUT(sut, "specific-only")
				appendState(state, "specific-only")
			},
			Assertion: DummyAssert{"assert-specific-only"},
		},
		{
			Name: "missingSpecificBuilder",
			StateBuilder: func(t *testing.T, sut *DummySUT, state *DummyState) {
				t.Helper()

				appendSUT(sut, "state-only")
				appendState(state, "state-only")
			},
			SpecificBuilder: nil,
			Assertion:       DummyAssert{"assert-state-only"},
		},
		{
			Name:            "bothMissing",
			StateBuilder:    nil,
			SpecificBuilder: nil,
			Assertion:       DummyAssert{"assert-none"},
		},
	}

	t.Run("run all indices (loop form)", func(t *testing.T) {
		sutAllActualCalled := make([][]string, 0)
		stateAllActualCalled := make([][]string, 0)
		actualNames := make([]string, 0)

		expectedSUTAll := [][]string{
			{"sut-specific-only"},
			{"sut-state-only"},
			{"sut-state-only"},
		}
		expectedStateAll := [][]string{
			{"state-specific-only"},
			{"state-state-only"},
			{"state-state-only"},
		}
		expectedNames := []string{"missingStateBuilder", "missingSpecificBuilder", "bothMissing"}

		for i := range tests {
			data, err := TestDataFromSlice(t, i, tests)

			require.NoError(t, err)
			assert.Equal(t, expectedNames[i], tests[i].Name)

			sutAllActualCalled = append(sutAllActualCalled, data.SUT.actualCalled)
			stateAllActualCalled = append(stateAllActualCalled, data.State.actualCalled)
			actualNames = append(actualNames, tests[i].Name)
		}

		assert.Equal(t, expectedSUTAll, sutAllActualCalled)
		assert.Equal(t, expectedStateAll, stateAllActualCalled)
		assert.Equal(t, expectedNames, actualNames)
	})
}

// ===============================================================

func Test_TestDataFromSlice_RepeatedIndices_And_Order(t *testing.T) {
	t.Parallel()

	tests := []TableTestItem[DummySUT, DummyState, DummyAssert]{
		{
			Name: "A",
			StateBuilder: func(t *testing.T, sut *DummySUT, state *DummyState) {
				t.Helper()

				appendSUT(sut, "stateA")
				appendState(state, "stateA")
			},
			SpecificBuilder: func(t *testing.T, sut *DummySUT, state *DummyState) {
				t.Helper()

				appendSUT(sut, "specA")
				appendState(state, "specA")
			},
			Assertion: DummyAssert{"assertA"},
		},
		{
			Name: "B",
			StateBuilder: func(t *testing.T, sut *DummySUT, state *DummyState) {
				t.Helper()

				appendSUT(sut, "stateB")
				appendState(state, "stateB")
			},
			SpecificBuilder: func(t *testing.T, sut *DummySUT, state *DummyState) {
				t.Helper()

				appendSUT(sut, "specB")
				appendState(state, "specB")
			},
			Assertion: DummyAssert{"assertB"},
		},
	}

	runIndices := []int{0, 1, 1, 0}
	expectedSUTAll := [][]string{
		{"sut-stateA", "sut-specA"},
		{"sut-stateA", "sut-stateB", "sut-specB"},
		{"sut-stateA", "sut-stateB", "sut-specB"},
		{"sut-stateA", "sut-specA"},
	}

	expectedStateAll := [][]string{
		{"state-stateA", "state-specA"},
		{"state-stateA", "state-stateB", "state-specB"},
		{"state-stateA", "state-stateB", "state-specB"},
		{"state-stateA", "state-specA"},
	}

	sutAllActualCalled := make([][]string, 0)
	stateAllActualCalled := make([][]string, 0)
	actualNames := make([]string, 0)

	for _, idx := range runIndices {
		data, err := TestDataFromSlice(t, idx, tests)
		require.NoError(t, err)

		sutAllActualCalled = append(sutAllActualCalled, data.SUT.actualCalled)
		stateAllActualCalled = append(stateAllActualCalled, data.State.actualCalled)
		actualNames = append(actualNames, tests[idx].Name)
	}

	assert.Equal(t, expectedSUTAll, sutAllActualCalled)
	assert.Equal(t, expectedStateAll, stateAllActualCalled)
	assert.Equal(t, []string{"A", "B", "B", "A"}, actualNames)
}

// ===============================================================

func Test_TestDataFromSlice_IndexErrors(t *testing.T) {
	tests := []TableTestItem[DummySUT, DummyState, DummyAssert]{
		{Name: "A", Assertion: DummyAssert{"A"}},
	}

	t.Run("negative index", func(t *testing.T) {
		data, err := TestDataFromSlice(t, -1, tests)
		require.ErrorIs(t, err, ErrIndexOutOfRange)
		assert.Empty(t, data.SUT)
		assert.Empty(t, data.State)
	})

	t.Run("too large index", func(t *testing.T) {
		data, err := TestDataFromSlice(t, 5, tests)
		require.ErrorIs(t, err, ErrIndexOutOfRange)
		assert.Empty(t, data.SUT)
		assert.Empty(t, data.State)
	})

	t.Run("no tests defined", func(t *testing.T) {
		var empty []TableTestItem[DummySUT, DummyState, DummyAssert]

		data, err := TestDataFromSlice(t, 0, empty)

		require.ErrorIs(t, err, ErrNoTestsDefined)
		assert.Empty(t, data.SUT)
		assert.Empty(t, data.State)
	})
}

func Test_TestDataFromSlice_PanicInBuilder_DoesNotStopOthers(t *testing.T) {
	tests := []TableTestItem[DummySUT, DummyState, DummyAssert]{
		{
			Name: "normal-before-panics",
			StateBuilder: func(t *testing.T, sut *DummySUT, state *DummyState) {
				t.Helper()

				appendSUT(sut, "pre-panic-statebuilder")
				appendState(state, "pre-panic-statebuilder")
			},
			SpecificBuilder: func(t *testing.T, sut *DummySUT, state *DummyState) {
				t.Helper()

				appendSUT(sut, "should-run-before-panic-statebuilder")
				appendState(state, "should-run-before-panic-statebuilder")
			},
			Assertion: DummyAssert{"assert-pre-panic-statebuilder"},
		},
		{
			Name: "panic-in-specificbuilder",
			StateBuilder: func(t *testing.T, sut *DummySUT, state *DummyState) {
				t.Helper()

				appendSUT(sut, "ok-statebuilder")
				appendState(state, "ok-statebuilder")
			},
			SpecificBuilder: func(t *testing.T, sut *DummySUT, state *DummyState) {
				t.Helper()

				panic("boom-in-specificbuilder")
			},
			Assertion: DummyAssert{"assert-panic-specificbuilder"},
		},
		{
			Name: "normal-after-panics",
			StateBuilder: func(t *testing.T, sut *DummySUT, state *DummyState) {
				t.Helper()

				appendSUT(sut, "normal-state")
				appendState(state, "normal-state")
			},
			SpecificBuilder: func(t *testing.T, sut *DummySUT, state *DummyState) {
				t.Helper()

				appendSUT(sut, "normal-specific")
				appendState(state, "normal-specific")
			},
			Assertion: DummyAssert{"assert-normal"},
		},
		{
			Name: "panic-in-statebuilder",
			StateBuilder: func(t *testing.T, sut *DummySUT, state *DummyState) {
				t.Helper()

				panic("boom-in-statebuilder")
			},
			SpecificBuilder: func(t *testing.T, sut *DummySUT, state *DummyState) {
				t.Helper()
			},
			Assertion: DummyAssert{"assert-normal"},
		},
	}

	sutAllActualCalled := make([][]string, 0)
	stateAllActualCalled := make([][]string, 0)
	actualAsserts := make([]string, 0)
	actualNames := make([]string, 0)
	actualPanics := make([]string, 0)

	expectedSUTAll := [][]string{
		{"sut-pre-panic-statebuilder", "sut-should-run-before-panic-statebuilder"},
		{"sut-pre-panic-statebuilder", "sut-ok-statebuilder", "sut-normal-state", "sut-normal-specific"},
	}

	expectedStateAll := [][]string{
		{"state-pre-panic-statebuilder", "state-should-run-before-panic-statebuilder"},
		{"state-pre-panic-statebuilder", "state-ok-statebuilder", "state-normal-state", "state-normal-specific"},
	}
	expectedAssert := []string{
		"assert-pre-panic-statebuilder",
		"assert-normal",
	}

	expectedNames := []string{
		"normal-before-panics",
		"panic-in-specificbuilder",
		"normal-after-panics",
		"panic-in-statebuilder",
	}
	expectedPanics := []string{
		"",
		"boom-in-specificbuilder",
		"",
		"boom-in-statebuilder",
	}

	for i, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			actualNames = append(actualNames, tests[i].Name)

			defer func() {
				if r := recover(); r != nil {
					t.Logf("Recovered panic in subtest %q: %v", tt.Name, r)

					actualPanic, ok := r.(string)

					assert.True(t, ok)

					actualPanics = append(actualPanics, actualPanic)
				}
			}()

			data, err := TestDataFromSlice(t, i, tests)
			require.NoError(t, err)

			sutAllActualCalled = append(sutAllActualCalled, data.SUT.actualCalled)
			stateAllActualCalled = append(stateAllActualCalled, data.State.actualCalled)
			actualAsserts = append(actualAsserts, data.Assert.Name)
			actualPanics = append(actualPanics, "")
		})
	}

	t.Run("Assert correct call with panics", func(t *testing.T) {
		assert.Equal(t, expectedSUTAll, sutAllActualCalled)
		assert.Equal(t, expectedStateAll, stateAllActualCalled)
		assert.Equal(t, expectedAssert, actualAsserts)
		assert.Equal(t, expectedNames, actualNames)
		assert.Equal(t, expectedPanics, actualPanics)
	})
}
