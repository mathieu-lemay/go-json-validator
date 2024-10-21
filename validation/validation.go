package validation

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/rs/zerolog/log"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

var structsCache = map[reflect.Type]reflect.Type{}

func createPointerStruct(src interface{}) reflect.Type {
	srcType := reflect.TypeOf(src)
	if ptrType, ok := structsCache[srcType]; ok {
		return ptrType
	}

	var sfs []reflect.StructField
	var t reflect.Type

	vt := reflect.ValueOf(src)
	if vt.Kind() == reflect.Ptr {
		t = vt.Elem().Type()
	} else {
		t = vt.Type()
	}

	// TODO: Add required tag on non ptr values

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		switch field.Type.Kind() {
		case reflect.Struct:
			interfaceType := createPointerStruct(reflect.New(field.Type).Elem().Interface())
			sf := reflect.StructField{
				Name: field.Name,
				Type: reflect.PointerTo(interfaceType),
				Tag:  field.Tag,
			}
			sfs = append(sfs, sf)
		case reflect.Ptr:
			sf := reflect.StructField{
				Name: field.Name,
				Type: field.Type,
				Tag:  field.Tag,
			}
			sfs = append(sfs, sf)
		default:
			sf := reflect.StructField{
				Name: field.Name,
				Type: reflect.PointerTo(field.Type),
				Tag:  field.Tag,
			}
			sfs = append(sfs, sf)
		}

	}

	ptrType := reflect.StructOf(sfs)
	structsCache[srcType] = ptrType

	return ptrType
}

func copyStruct(src interface{}, dst interface{}) error {
	sv := reflect.ValueOf(src)
	if sv.Kind() == reflect.Ptr {
		sv = sv.Elem()
	}
	dv := reflect.ValueOf(dst)
	if dv.Kind() == reflect.Ptr {
		dv = dv.Elem()
	}

	dt := dv.Type()

	log.Debug().Str("dvKind", dv.Kind().String()).Str("dt", dt.String()).Send()

	for i := range dv.NumField() {
		f := dt.Field(i)

		sf := sv.FieldByName(f.Name)
		sfIsPtr := sf.Kind() == reflect.Ptr

		df := dv.FieldByName(f.Name)
		dfIsPtr := df.Kind() == reflect.Ptr

		sVal := sf
		if !dfIsPtr {
			if sf.IsNil() {
				return fmt.Errorf("can't set nil value to non ptr field: %s", f.Name)
			}
			sVal = sf.Elem()
		}

		log.Debug().
			Interface("field", f).
			Interface("sfType", sf.Type().String()).
			Interface("sfIsPtr", sfIsPtr).
			Interface("sfIsNil", sf.IsNil()).
			Interface("dfType", df.Type().String()).
			Interface("dfIsPtr", dfIsPtr).
			Interface("sVal", sVal.String()).
			Msg("field")

		df.Set(sVal)
	}

	return nil
}

func parse(data []byte, target interface{}) (interface{}, error) {
	// Create copy of the given type, but with all fields as pointers
	ptrType := createPointerStruct(target)

	// Something something interface, something something golang
	// https://stackoverflow.com/a/45680060
	ptrInstance := reflect.New(ptrType).Interface()

	if err := json.Unmarshal(data, &ptrInstance); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return ptrInstance, nil
}

func parseAndValidate(data []byte, target interface{}) error {
	ptrInstance, err := parse(data, target)
	if err != nil {
		return fmt.Errorf("failed to parse data: %w", err)
	}

	err = validate.Struct(ptrInstance)
	if err != nil {
		return fmt.Errorf("invalid data: %w", err)
	}

	// Fill the original struct with validated values
	return copyStruct(ptrInstance, target)
}
