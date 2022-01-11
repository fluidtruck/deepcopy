package deepcopy

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// Case 1: Struct Conversion
type Ex1StructA struct {
	Foo string
}

type Ex1StructB struct {
	Foo string
	Bar uint64
}

var (
	ex1objA = Ex1StructA{Foo: "leia"}
	ex1objB = Ex1StructB{}
)

func TestDeepCopy(t *testing.T) {
	testCases := []struct {
		name              string
		input             interface{}
		outputPtr         interface{}
		expectedOutputPtr interface{}
		expectedError     error
	}{
		{
			name:              "Case 1: struct conversion",
			input:             ex1objA,
			outputPtr:         &ex1objB,
			expectedOutputPtr: &Ex1StructB{Foo: "leia"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := DeepCopy(tc.input, tc.outputPtr)
			if tc.expectedError != nil {
				require.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedOutputPtr, tc.outputPtr)
			}
		})
	}
}
