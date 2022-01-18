package deepcopy

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

// Example 1
type Ex1StructA struct {
	Foo string
}
type Ex1StructB struct {
	Foo string
}

var (
	ex1objA = Ex1StructA{Foo: "leia"}
	ex1objB = Ex1StructB{}
)

// Example 2
type Ex2StructA struct {
	Zak bool
	Yo  uint32
}
type Ex2StructB struct {
	Zak bool
	Yo  int64
}
type Ex2StructC struct {
	Foo *Ex2StructA
}
type Ex2StructD struct {
	Foo *Ex2StructB
}

var (
	ex2objC = Ex2StructC{
		Foo: &Ex2StructA{
			Zak: true,
			Yo:  uint32(65),
		},
	}
	ex2objD = Ex2StructD{}
)

// Example 3
type Ex3StructA struct {
	Foodx string
}
type Ex3StructB struct {
	Food Ex3StructA
}
type Ex3StructC struct {
	Foo  string
	Bar  int
	Zak  bool
	Yo   *bool
	Sel  []*Ex3StructB
	Mar  map[string]string
	Frey map[uint]float64
	Rand *time.Time
	Dot  *timestamppb.Timestamp
}

var (
	yo           = true
	timeNow      = time.Now()
	timestampNow = timestamppb.New(timeNow)
	ex3objC      = Ex3StructC{
		Foo: "hello there!",
		Bar: 34234,
		Zak: false,
		Yo:  &yo,
		Sel: []*Ex3StructB{
			&Ex3StructB{
				Food: Ex3StructA{
					Foodx: "nested string here",
				},
			},
			&Ex3StructB{
				Food: Ex3StructA{
					Foodx: "second nested string",
				},
			},
		},
		Mar: map[string]string{
			"first":  "f i r s t",
			"second": "the second one",
		},
		Frey: map[uint]float64{
			uint(4): 8.2343,
			uint(5): 23.4322,
			uint(6): 883423.0,
		},
		Rand: &timeNow,
		Dot:  timestampNow,
	}
)

// Example 4
var (
	ex4String   = "42"
	ex4EmptyInt int32
	ex4Int      = int32(42)
)

// Example 5
var (
	ex5String    = "true"
	ex5EmptyBool bool
	ex5Bool      = true
)

func TestDeepCopy(t *testing.T) {
	testCases := []struct {
		name              string
		input             interface{}
		outputPtr         interface{}
		expectedOutputPtr interface{}
		expectedError     error
	}{
		/* **** Case 1: Struct Conversion **** */
		{
			name:              "Example 1: struct conversion",
			input:             ex1objA,
			outputPtr:         &ex1objB,
			expectedOutputPtr: &Ex1StructB{Foo: "leia"},
		},
		{
			name:      "Example 2: struct conversion, nested fields",
			input:     ex2objC,
			outputPtr: &ex2objD,
			expectedOutputPtr: &Ex2StructD{
				Foo: &Ex2StructB{
					Zak: true,
					Yo:  int64(65),
				},
			},
		},
		/* **** Case 2: Identical Copy **** */
		{
			name:      "Example 3: identical copy, nested fields",
			input:     ex3objC,
			outputPtr: &Ex3StructC{},
			expectedOutputPtr: &Ex3StructC{
				Foo: "hello there!",
				Bar: 34234,
				Zak: false,
				Yo:  &yo,
				Sel: []*Ex3StructB{
					&Ex3StructB{
						Food: Ex3StructA{
							Foodx: "nested string here",
						},
					},
					&Ex3StructB{
						Food: Ex3StructA{
							Foodx: "second nested string",
						},
					},
				},
				Mar: map[string]string{
					"first":  "f i r s t",
					"second": "the second one",
				},
				Frey: map[uint]float64{
					uint(4): 8.2343,
					uint(5): 23.4322,
					uint(6): 883423.0,
				},
				Rand: &timeNow,
				Dot:  timestampNow,
			},
		},
		/* **** Case 3: General Type Casting **** */
		{
			name:              "Example 4: type casting, string to int",
			input:             ex4String,    // "42"
			outputPtr:         &ex4EmptyInt, // empty int var
			expectedOutputPtr: &ex4Int,      // 42 (int)
		},
		{
			name:              "Example 4: type casting, string to bool",
			input:             ex5String,     // "true"
			outputPtr:         &ex5EmptyBool, // empty bool
			expectedOutputPtr: &ex5Bool,      // true
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
