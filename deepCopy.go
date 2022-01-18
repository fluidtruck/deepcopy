package deepcopy

import (
	"errors"
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
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
	DC_STRUCT_TAG = "dc"
)

var (
	timeType           = reflect.TypeOf(time.Time{})
	timePtrType        = reflect.TypeOf(&time.Time{})
	timestamppbPtrType = reflect.TypeOf(&timestamppb.Timestamp{})
)

func smartCopy(inValue reflect.Value, outValue reflect.Value) (err error) {
	errCouldNotConvert := fmt.Errorf("unable to convert %s (type %s) to type %s", inValue.Interface(), inValue.Type(), outValue.Type())
	if !outValue.CanSet() {
		err := fmt.Errorf("value of %s cannot be set", outValue.Interface())
		return err
	}
	done := false

	// handle string -> number
	if inValue.Kind() == reflect.String {
		attempted, noError := parseStringFlexibly(inValue, outValue)
		if attempted {
			if !noError {
				return errCouldNotConvert
			}
			return
		}
	}

	// handle *timestamppb.Timestamp
	if inValue.Type() == timestamppbPtrType {
		err = convertFromTimestampPbPointer(inValue, outValue)
		if err != nil {
			return err
		}
		return
	} else if outValue.Type() == timestamppbPtrType {
		err = convertToTimestampPbPointer(inValue, outValue)
		if err != nil {
			return err
		}
		return
	}

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
	case reflect.Struct:
		startingCount := 0
		if outValue.Type() == timeType {
			err = convertToTime(inValue, outValue)
			if err != nil {
				return err
			}
			startingCount = inValue.NumField()
		} else if inValue.Kind() != reflect.Struct {
			return errCouldNotConvert
		}

		inValNumField := inValue.NumField()
		for i := startingCount; i < inValNumField; i++ {
			foundMatchingOutputField := false
			inputField := inValue.Field(i)
			inputFieldName := inValue.Type().Field(i).Name
			if !inputField.CanInterface() {
				// skip unexported fields
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
				outputFieldName := outValue.Type().Field(j).Name
				rune0, _ := utf8.DecodeRuneInString(outputFieldName)
				if unicode.IsLower(rune0) {
					// skip unexported fields
					continue
				}

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
	case timestamppbPtrType:
		inTimePreConvert := inValue.Interface().(*timestamppb.Timestamp)
		inTime := inTimePreConvert.AsTime()
		inTimeVal := reflect.ValueOf(inTime)
		newOutVal := reflect.New(reflect.TypeOf(inTime))
		newOutVal.Elem().Set(inTimeVal)
		outValue.Set(newOutVal.Elem())
	default:
		return errCouldNotConvert
	}
	return nil
}

func convertFromTimestampPbPointer(inValue, outValue reflect.Value) error {
	errCouldNotConvert := fmt.Errorf("unable to convert %s (type %s) to type %s", inValue.Interface(), inValue.Type(), outValue.Type())
	if inValue.Type() != timestamppbPtrType {
		return errCouldNotConvert
	}
	switch outValue.Type() {
	case timestamppbPtrType:
		outValue.Set(inValue)
	case timeType:
		inTimePreConvert := inValue.Interface().(*timestamppb.Timestamp)
		inTime := inTimePreConvert.AsTime()
		inTimeVal := reflect.ValueOf(inTime)
		newOutVal := reflect.New(reflect.TypeOf(inTime))
		newOutVal.Elem().Set(inTimeVal)
		outValue.Set(newOutVal.Elem())
	case timePtrType:
		inTimePreConvert := inValue.Interface().(*timestamppb.Timestamp)
		inTime := inTimePreConvert.AsTime()
		inTimeVal := reflect.ValueOf(&inTime)
		newOutVal := reflect.New(reflect.TypeOf(&inTime))
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
	inFieldTag := strings.ToLower(inField.Tag.Get(DC_STRUCT_TAG))
	outFieldTag := strings.ToLower(outField.Tag.Get(DC_STRUCT_TAG))

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

// TODO: test for converting string to every one of these types
func parseStringFlexibly(inValue, outValue reflect.Value) (didAttempt bool, worked bool) {
	// bool #1 represents "Did we try to convert?"
	didAttempt = true
	// bool #2 represents whether conversion worked
	worked = true
	// if bool #1 is false, then bool #2 is ignored
	if inValue.Kind() != reflect.String {
		return false, true
	}
	s := strings.ToLower(inValue.String())
	switch outValue.Kind() {
	default:
		didAttempt = false
	case reflect.Bool:
		if s == "t" || s == "true" {
			outValue.SetBool(true)
		} else if s == "f" || s == "false" {
			outValue.SetBool(false)
		} else {
			worked = false
		}
	case reflect.Int, reflect.Int64:
		parsedString, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return true, false
		}
		val := reflect.ValueOf(parsedString)
		outValue.Set(val)
	case reflect.Int8:
		parsedString, err := strconv.ParseInt(s, 10, 8)
		if err != nil {
			return true, false
		}
		num := int8(parsedString)
		val := reflect.ValueOf(num)
		outValue.Set(val)
	case reflect.Int16:
		parsedString, err := strconv.ParseInt(s, 10, 16)
		if err != nil {
			return true, false
		}
		num := int16(parsedString)
		val := reflect.ValueOf(num)
		outValue.Set(val)
	case reflect.Int32:
		parsedString, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return true, false
		}
		num := int32(parsedString)
		val := reflect.ValueOf(num)
		outValue.Set(val)
	case reflect.Uint, reflect.Uint64:
		parsedString, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return true, false
		}
		val := reflect.ValueOf(parsedString)
		outValue.Set(val)
	case reflect.Uint8:
		parsedString, err := strconv.ParseUint(s, 10, 8)
		if err != nil {
			return true, false
		}
		num := uint8(parsedString)
		val := reflect.ValueOf(num)
		outValue.Set(val)
	case reflect.Uint16:
		parsedString, err := strconv.ParseUint(s, 10, 16)
		if err != nil {
			return true, false
		}
		num := uint16(parsedString)
		val := reflect.ValueOf(num)
		outValue.Set(val)
	case reflect.Uint32:
		parsedString, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return true, false
		}
		num := uint32(parsedString)
		val := reflect.ValueOf(num)
		outValue.Set(val)
	case reflect.Float64:
		parsedString, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return true, false
		}
		val := reflect.ValueOf(parsedString)
		outValue.Set(val)
	case reflect.Float32:
		parsedString, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return true, false
		}
		num := float32(parsedString)
		val := reflect.ValueOf(num)
		outValue.Set(val)
	}

	return
}
