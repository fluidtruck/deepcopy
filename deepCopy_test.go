package deepcopy

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

type LocalInspection struct {
	UUID          string
	ID            uint
	UserID        uint64
	IsValid       bool
	ReservationID uint // different from pbInspection ReservationId (uint64)
	Keys          map[string]string
	OddKeys       map[uint]uint64
	RandomTime    *time.Time
	RandomTime2   *timestamppb.Timestamp
}

type PbInspection struct {
	Uuid          string
	Id            uint
	UserId        uint64
	IsValid       bool
	ReservationId uint64
	Keys          map[string]string
	OddKeys       map[uint]uint64
	RandomTime    *time.Time
	RandomTime2   *time.Time
}

type JustTimeA struct {
	RandomTime *time.Time
}

type JustTimeB struct {
	RandomTime *timestamppb.Timestamp
}

type PbTime struct {
	Time0 timestamppb.Timestamp
	Time  *timestamppb.Timestamp
	Time2 *timestamppb.Timestamp
	Time3 *time.Time
	Time4 *time.Time
}

type LocalTime struct {
	Time  *time.Time
	Time2 *timestamppb.Timestamp
	Time3 *time.Time
	Time4 *timestamppb.Timestamp
}

type LocalFunky struct {
	Hello string `dc:"hi"`
	Sup   int
}

type DcFunky struct {
	Hi   string
	Sup2 uint32 `dc:"sup"`
}

type LocalTeacher struct {
	TeacherID uint
}

type PbTeacher struct {
	TeacherId uint64
}

type LocalClassroom struct {
	Teacher1   LocalTeacher
	Teacher2   *LocalTeacher
	Parent     LocalTeacher
	Child      *LocalTeacher
	InsertedAt *time.Time
	UpdatedAt  *time.Time
}

type PbClassroom struct {
	Teacher1   PbTeacher
	Teacher2   *PbTeacher
	Parent     *PbTeacher
	Child      PbTeacher
	InsertedAt *timestamppb.Timestamp
	UpdatedAt  *timestamppb.Timestamp
}

type Thing struct {
	PointerPointer *string
	DoublePointer  **string
	A              string
	B              *uint64
}

type Thing2 struct {
	PointerPointer *string
	DoublePointer  **string
	A              *string
	B              uint
}

type LocalInspectionType string

const (
	LocalInspectionType_Pickup LocalInspectionType = "pickup"
)

type InspectionType int32

type NewStruct1 struct {
	Field                  int64
	ExtraInputFieldNoMatch string
	Enum                   LocalInspectionType
	BadConversion1         int32
}

type NewStruct2 struct {
	Field                   int32
	ExtraOutputFieldNoMatch *time.Time
	Enum                    InspectionType
	BadConversion1          timestamppb.Timestamp
}

type ClassroomA struct {
	Teacher LocalTeacher
}

type ClassroomB struct {
	Classroom ClassroomA
	Locality  string
}

func TestDeepCopy(t *testing.T) {
	optionalString1 := "pointer string"
	stringPointer1 := &optionalString1
	num1 := uint64(5)
	emptyString := ""
	otherNum1 := uint(5)
	utc, _ := time.LoadLocation("UTC")
	time1 := time.Now().Add(time.Hour * 5).In(utc)
	time2 := time.Now().Add(time.Hour * 11).In(utc)
	timestamppb1 := timestamppb.New(time1)
	timestamppb2 := timestamppb.New(time2)

	testCases := []struct {
		name            string
		input           interface{}
		outputPtr       interface{}
		expectedRespPtr interface{}
		expectedErr     error
	}{
		{
			name: "bad conversion to timestamp",
			input: NewStruct1{
				BadConversion1: int32(12),
			},
			outputPtr:   &NewStruct2{},
			expectedErr: errors.New("unable to convert %!s(int32=12) (type int32) to type timestamppb.Timestamp"),
		},
		{
			name: "extra input field with no match",
			input: NewStruct1{
				ExtraInputFieldNoMatch: "thing",
			},
			outputPtr:       &NewStruct2{},
			expectedRespPtr: &NewStruct2{},
		},
		{
			name: "extra output field with no match",
			input: NewStruct1{
				Field: int64(10),
			},
			outputPtr: &NewStruct2{
				ExtraOutputFieldNoMatch: &time1,
			},
			expectedRespPtr: &NewStruct2{
				Field:                   int32(10),
				ExtraOutputFieldNoMatch: &time1,
			},
		},
		{
			name: "output field embedded in input, no matching fields",
			input: ClassroomA{
				Teacher: LocalTeacher{
					TeacherID: uint(32),
				},
			},
			outputPtr: &ClassroomB{
				Classroom: ClassroomA{},
				Locality:  "locality",
			},
			expectedRespPtr: &ClassroomB{
				Classroom: ClassroomA{},
				Locality:  "locality",
			},
		},
		{
			name: "enum with no match, should fail",
			input: NewStruct1{
				Enum: LocalInspectionType_Pickup,
			},
			outputPtr:   &NewStruct2{},
			expectedErr: errors.New("unable to convert pickup (type deepcopy.LocalInspectionType) to type deepcopy.InspectionType"),
		},
		{
			name:        "non-pointer outputType, should fail",
			input:       NewStruct1{},
			outputPtr:   4,
			expectedErr: errors.New("expected pointer for arg1 %!s(int=4) but received int"),
		},
		{
			name: "enum but null, should pass",
			input: NewStruct1{
				Enum: "",
			},
			outputPtr:       &NewStruct2{},
			expectedRespPtr: &NewStruct2{},
		},
		{
			name: "int64 to int32, under limit",
			input: NewStruct1{
				Field: int64(100),
			},
			outputPtr: &NewStruct2{},
			expectedRespPtr: &NewStruct2{
				Field: int32(100),
			},
		},
		{
			name: "dc struct tag",
			input: LocalFunky{
				Hello: "mystring",
				Sup:   3,
			},
			outputPtr: &DcFunky{},
			expectedRespPtr: &DcFunky{
				Hi:   "mystring",
				Sup2: uint32(3),
			},
		},
		{
			name: "dc struct tag, part 2",
			input: DcFunky{
				Hi:   "mystring",
				Sup2: uint32(3),
			},
			outputPtr: &LocalFunky{},
			expectedRespPtr: &LocalFunky{
				Hello: "mystring",
				Sup:   3,
			},
		},
		{
			name: "timestamp to time",
			input: &JustTimeB{
				RandomTime: timestamppb1,
			},
			outputPtr: &JustTimeA{},
			expectedRespPtr: &JustTimeA{
				RandomTime: &time1,
			},
		},
		{
			name: "CreatedUpdated to *InsertedAt *UpdatedAt",
			input: PbClassroom{
				InsertedAt: timestamppb1,
				UpdatedAt:  timestamppb2,
			},
			outputPtr: &LocalClassroom{},
			expectedRespPtr: &LocalClassroom{
				InsertedAt: &time1,
				UpdatedAt:  &time2,
			},
		},
		{
			name: "inspections obj, should ignore bad field InspectionType",
			input: &LocalInspection{
				UUID:          "hello there",
				ID:            uint(2),
				UserID:        uint64(2),
				IsValid:       true,
				ReservationID: otherNum1,
				Keys: map[string]string{
					"hi": "you",
				},
				OddKeys: map[uint]uint64{
					uint(5): uint64(55),
				},
				RandomTime: &time1,
			},
			outputPtr: &PbInspection{},
			expectedRespPtr: &PbInspection{
				Uuid:          "hello there",
				Id:            uint(2),
				UserId:        uint64(2),
				IsValid:       true,
				ReservationId: num1,
				Keys: map[string]string{
					"hi": "you",
				},
				OddKeys: map[uint]uint64{
					uint(5): uint64(55),
				},
				RandomTime: &time1,
			},
		},
		{
			name: "inspections obj, but in opposite direction",
			input: &PbInspection{
				Uuid:    "hello there",
				Id:      uint(2),
				UserId:  uint64(2),
				IsValid: true,
				Keys: map[string]string{
					"hi": "you",
				},
				OddKeys: map[uint]uint64{
					uint(5): uint64(55),
				},
			},
			outputPtr: &LocalInspection{},
			expectedRespPtr: &LocalInspection{
				UUID:    "hello there",
				ID:      uint(2),
				UserID:  uint64(2),
				IsValid: true,
				Keys: map[string]string{
					"hi": "you",
				},
				OddKeys: map[uint]uint64{
					uint(5): uint64(55),
				},
			},
		},
		{
			name: "uint to uint64",
			input: &LocalInspection{
				ReservationID: uint(34),
			},
			outputPtr: &PbInspection{},
			expectedRespPtr: &PbInspection{
				ReservationId: uint64(34),
			},
		},
		{
			name:            "*string to *string",
			input:           &optionalString1,
			outputPtr:       &emptyString,
			expectedRespPtr: &optionalString1,
		},
		{
			name: "**string to **string",
			input: &Thing{
				DoublePointer: &stringPointer1,
			},
			outputPtr: &Thing2{},
			expectedRespPtr: &Thing2{
				DoublePointer: &stringPointer1,
			},
		},
		{
			name: "string to *string",
			input: &Thing{
				A: optionalString1,
			},
			outputPtr: &Thing2{},
			expectedRespPtr: &Thing2{
				A: &optionalString1,
			},
		},
		{
			name: "*string to string",
			input: &Thing2{
				A: &optionalString1,
			},
			outputPtr: &Thing{},
			expectedRespPtr: &Thing{
				A: optionalString1,
			},
		},
		{
			name: "*uint64 to uint",
			input: &Thing{
				B: &num1,
			},
			outputPtr: &Thing2{},
			expectedRespPtr: &Thing2{
				B: otherNum1,
			},
		},
		{
			name: "uint to *uint64",
			input: &Thing2{
				B: otherNum1,
			},
			outputPtr: &Thing{},
			expectedRespPtr: &Thing{
				B: &num1,
			},
		},
		{
			name: "nested struct, only top-level pointers",
			input: &LocalClassroom{
				Teacher1: LocalTeacher{
					TeacherID: uint(1),
				},
			},
			outputPtr: &PbClassroom{},
			expectedRespPtr: &PbClassroom{
				Teacher1: PbTeacher{
					TeacherId: uint64(1),
				},
			},
		},
		{
			name: "nested struct, pointer to nonpointer",
			input: &PbClassroom{
				Parent: &PbTeacher{
					TeacherId: uint64(2),
				},
			},
			outputPtr: &LocalClassroom{},
			expectedRespPtr: &LocalClassroom{
				Parent: LocalTeacher{
					TeacherID: uint(2),
				},
			},
		},
		{
			name: "nested struct, nested pointers",
			input: &PbClassroom{
				Teacher2: &PbTeacher{
					TeacherId: uint64(2),
				},
			},
			outputPtr: &LocalClassroom{},
			expectedRespPtr: &LocalClassroom{
				Teacher2: &LocalTeacher{
					TeacherID: uint(2),
				},
			},
		},
		{
			name: "nested struct, nonpointer to pointer",
			input: &LocalClassroom{
				Parent: LocalTeacher{
					TeacherID: uint(1),
				},
			},
			outputPtr: &PbClassroom{},
			expectedRespPtr: &PbClassroom{
				Parent: &PbTeacher{
					TeacherId: uint64(1),
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := DeepCopy(tc.input, tc.outputPtr)
			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedRespPtr, tc.outputPtr)
			}
		})
	}
}

func TestDeepCopyToTimestamp(t *testing.T) {
	time1 := time.Now()
	timestamppb1 := timestamppb.New(time1)

	testCases := []struct {
		name            string
		input           interface{}
		outputPtr       interface{}
		expectedRespPtr interface{}
		expectedErr     error
	}{
		{
			name: "*time to *timestamppb",
			input: &LocalTime{
				Time: &time1,
			},
			outputPtr: &PbTime{},
			expectedRespPtr: &PbTime{
				Time: timestamppb1,
			},
		},
		{
			name: "*timestamppb to *timestamppb",
			input: &LocalTime{
				Time2: timestamppb1,
			},
			outputPtr: &PbTime{},
			expectedRespPtr: &PbTime{
				Time2: timestamppb1,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := DeepCopy(tc.input, tc.outputPtr)
			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
				e := tc.expectedRespPtr.(*PbTime)
				r := tc.outputPtr.(*PbTime)
				assert.Equal(t, e.Time0.Seconds, r.Time0.Seconds)
				assert.Equal(t, e.Time0.Nanos, r.Time0.Nanos)
				if e.Time != nil {
					ee := *e.Time
					rr := *r.Time
					assert.Equal(t, ee.Seconds, rr.Seconds)
					assert.Equal(t, ee.Nanos, rr.Nanos)
				}
				if e.Time2 != nil {
					ee := *e.Time2
					rr := *r.Time2
					assert.Equal(t, ee.Seconds, rr.Seconds)
					assert.Equal(t, ee.Nanos, rr.Nanos)
				}
				if e.Time3 != nil {
					assert.Equal(t, e.Time3, r.Time3)
				}
				if e.Time4 != nil {
					assert.Equal(t, e.Time4, r.Time4)
				}
			}
		})
	}
}
