package deepcopy

import (
	"errors"
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"reflect"
	"strings"
	"time"
)

func DeepCopy(input, output interface{}) error {
	inputVal := reflect.ValueOf(input)
	outputVal := reflect.ValueOf(output)
	if outputVal.Kind() != reflect.Ptr {
		errOutValueNotPtr := fmt.Errorf("expected pointer for arg1 %s but received %s", outputVal, outputVal.Kind())
		return errOutValueNotPtr
	}
	outputVal = outputVal.Elem()
	inputVal = smartMaxDereference(inputVal, outputVal)
	err := smartCopy(inputVal, outputVal)
	if err != nil {
		return err
	}
	return nil
}

const (
	PB_STRUCT_TAG = "pb"
)

var (
	timeType           = reflect.TypeOf(time.Time{})
	timePtrType        = reflect.TypeOf(&time.Time{})
	timestamppbType    = reflect.TypeOf(timestamppb.Timestamp{})
	timestamppbPtrType = reflect.TypeOf(&timestamppb.Timestamp{})
)

func smartCopy(inValue reflect.Value, outValue reflect.Value) (err error) {
	errCouldNotConvert := fmt.Errorf("unable to convert %s (type %s) to type %s", inValue.Interface(), inValue.Type(), outValue.Type())
	if !outValue.CanSet() {
		err := fmt.Errorf("value of %s cannot be set", outValue.Interface())
		return err
	}
	done := false

	switch outValue.Kind() {
	default:
		if inValue.Type() != outValue.Type() && !CanConvert(inValue, outValue.Type()) {
			return errCouldNotConvert
		} else {
			newInValue := inValue.Convert(outValue.Type())
			outValue.Set(newInValue)
			done = true
		}
	case reflect.Array, reflect.Interface, reflect.Func, reflect.Map:
		outValue.Set(inValue)
		done = true
	case reflect.Slice:
		if inValue.Kind() != reflect.Slice {
			return errCouldNotConvert
		}
		sliceType := reflect.TypeOf(outValue.Interface())
		newOutValue := reflect.MakeSlice(sliceType, inValue.Len(), inValue.Len())
		for i := 0; i < inValue.Len(); i++ {
			inVal := inValue.Index(i)
			outVal := reflect.New(sliceType.Elem()).Elem()
			inVal = smartMaxDereference(inVal, outVal)
			err := smartCopy(inVal, outVal)
			if err != nil {
				return err
			}
			newOutValue.Index(i).Set(outVal)
		}
		outValue.Set(newOutValue)
		done = true
	case reflect.Ptr:
		if outValue.Type() == timestamppbPtrType {
			err = convertToTimestampPbPointer(inValue, outValue)
			if err != nil {
				return err
			}
			done = true
		} else {
			outValueInterfaceTypeOfElem := reflect.TypeOf(outValue.Interface()).Elem()
			childOutVal := reflect.New(reflect.TypeOf(inValue.Interface()))
			err := smartCopy(inValue, childOutVal.Elem())
			if err != nil {
				return err
			}
			childOutValOut := reflect.New(outValueInterfaceTypeOfElem)
			childOutValElemNonPtr := smartMaxDereference(childOutVal.Elem(), childOutValOut.Elem())
			err = smartCopy(childOutValElemNonPtr, childOutValOut.Elem())
			if err != nil {
				return err
			}
			outValuePtr := reflect.New(reflect.TypeOf(childOutValOut.Elem().Interface()))
			outValuePtr.Elem().Set(reflect.ValueOf(childOutValOut.Elem().Interface()))
			outValue.Set(outValuePtr)
			done = true
		}
	case reflect.Struct:
		startingCount := 0
		if outValue.Type() == timeType {
			err = convertToTime(inValue, outValue)
			if err != nil {
				return err
			}
			startingCount = inValue.NumField()
		} else if outValue.Type() == timestamppbType {
			// the only time this should ever happen is if inValue is also timestamppb
			// otherwise should be caught while type *timestamppb in reflect.Ptr case
			if inValue.Type() != timestamppbType {
				return errCouldNotConvert
			}
			outValue.Set(inValue)
			startingCount = inValue.NumField()
		} else if inValue.Kind() != reflect.Struct {
			return errCouldNotConvert
		}

		inValNumField := inValue.NumField()
		for i := startingCount; i < inValNumField; i++ {
			foundMatchingOutputField := false
			inputField := inValue.Field(i)
			inputFieldName := strings.ToLower(inValue.Type().Field(i).Name)
			if !inputField.CanInterface() {
				// skip un-exported fields
				continue
			}

			inputFieldInterface := inputField.Interface()
			if reflect.DeepEqual(inputFieldInterface, reflect.Zero(reflect.TypeOf(inputFieldInterface)).Interface()) {
				// skip null fields
				continue
			}
			for j := 0; j < outValue.NumField(); j++ {
				if foundMatchingOutputField {
					continue
				}
				outputField := outValue.Field(j)
				outputFieldName := strings.ToLower(outValue.Type().Field(j).Name)
				if fieldsMatch(reflect.TypeOf(inValue.Interface()).Field(i), reflect.TypeOf(outValue.Interface()).Field(j)) {
					if !inputField.IsValid() {
						err = errors.New(errCouldNotConvert.Error() + fmt.Sprintf(": field %s is invalid", inputFieldName))
						return err
					}
					if !outputField.CanSet() {
						err = errors.New(errCouldNotConvert.Error() + fmt.Sprintf(": cannot set field %s", outputFieldName))
						return err
					}
					inputField = smartMaxDereference(inputField, outputField)
					err = smartCopy(inputField, outputField)
					if err != nil {
						return err
					}
					foundMatchingOutputField = true
				}
			}
		}
		done = true
	}
	if !done {
		return errCouldNotConvert
	}
	return
}

func smartMaxDereference(input, output reflect.Value) reflect.Value {
	if input.Type() == timestamppbPtrType {
		if output.Type() != timeType {
			return input
		}
		return maxDereference(input)
	}
	return maxDereference(input)
}

func maxDereference(value reflect.Value) reflect.Value {
	if value.Kind() != reflect.Ptr {
		return value
	}
	return maxDereference(value.Elem())
}

func convertToTime(inValue, outValue reflect.Value) error {
	// remember: inValue will never be ptr
	errCouldNotConvert := fmt.Errorf("unable to convert %s (type %s) to type %s", inValue.Interface(), inValue.Type(), outValue.Type())
	if outValue.Type() != timeType {
		return errCouldNotConvert
	}
	switch inValue.Type() {
	case timeType:
		inTime := inValue.Interface().(time.Time)
		inTimeVal := reflect.ValueOf(inTime)
		newOutVal := reflect.New(reflect.TypeOf(inTime))
		newOutVal.Elem().Set(inTimeVal)
		outValue.Set(newOutVal.Elem())
	case timestamppbType:
		inTimePreConvert := inValue.Interface().(timestamppb.Timestamp)
		inTimePreConvertPtr := &inTimePreConvert
		inTime := inTimePreConvertPtr.AsTime()
		inTimeVal := reflect.ValueOf(inTime)
		newOutVal := reflect.New(reflect.TypeOf(inTime))
		newOutVal.Elem().Set(inTimeVal)
		outValue.Set(newOutVal.Elem())
	default:
		return errCouldNotConvert
	}
	return nil
}

func convertToTimestampPbPointer(inValue, outValue reflect.Value) error {
	errCouldNotConvert := fmt.Errorf("unable to convert %s (type %s) to type %s", inValue.Interface(), inValue.Type(), outValue.Type())
	if outValue.Type() != timestamppbPtrType {
		return errCouldNotConvert
	}
	switch inValue.Type() {
	case timeType:
		inTimePreConvert := inValue.Interface().(time.Time)
		inTimePtr := timestamppb.New(inTimePreConvert) // returns *timestamppb.Timestamp
		inTimePtrVal := reflect.ValueOf(inTimePtr)
		outValue.Set(inTimePtrVal)
	case timestamppbType:
		inTime := inValue.Interface().(timestamppb.Timestamp)
		inTimeVal := reflect.ValueOf(&inTime)
		outValue.Set(inTimeVal)
	case timestamppbPtrType:
		outValue.Set(inValue)
	default:
		return errCouldNotConvert
	}
	return nil
}

func fieldsMatch(inField, outField reflect.StructField) bool {
	inFieldName := strings.ToLower(inField.Name)
	outFieldName := strings.ToLower(outField.Name)
	if inFieldName == "" || outFieldName == "" {
		return false
	}
	inFieldTag := strings.ToLower(inField.Tag.Get(PB_STRUCT_TAG))
	outFieldTag := strings.ToLower(outField.Tag.Get(PB_STRUCT_TAG))

	if inFieldName == outFieldName || inFieldName == outFieldTag || outFieldName == inFieldTag {
		return true
	}

	return false
}

// taken from reflect in go@1.17:
func CanConvert(v reflect.Value, t reflect.Type) bool {
	vt := v.Type()
	if !vt.ConvertibleTo(t) {
		return false
	}
	// Currently the only conversion that is OK in terms of type
	// but that can panic depending on the value is converting
	// from slice to pointer-to-array.
	if vt.Kind() == reflect.Slice && t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Array {
		n := t.Elem().Len()
		if n > v.Len() {
			return false
		}
	}
	return true
}
