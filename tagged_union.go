package jtug

import (
	"encoding/json"
	"reflect"
)

type Tag interface {
	comparable
}

type Mapper[T Tag] interface {
	Unmarshal(b []byte, t T) (Union[T], error)
}

type Tagger interface {
	JSONTag() string
}

type Union[T Tag] any

type UnionList[T Tag, M Mapper[T]] []Union[T]

func UnmarshalTaggedField[M Mapper[T], S any, T comparable](
	structPointer *S,
	field any,
	bytes []byte,
) error {
	originalType := reflect.TypeOf(*structPointer)

	fieldType := reflect.ValueOf(field).Elem().Type()
	fieldCount := originalType.NumField()
	var index int

	// Make a copy of type but with json.RawMessage instead of field.
	fields := make([]reflect.StructField, fieldCount)
	for i := 0; i < fieldCount; i++ {
		fields[i] = originalType.Field(i)
		if fields[i].Type == fieldType {
			fields[i].Type = reflect.TypeOf(json.RawMessage{})
			index = i
		}
	}
	newType := reflect.StructOf(fields)
	newValue := reflect.New(newType)
	err := json.Unmarshal(bytes, newValue.Interface())
	if err != nil {
		return err
	}

	fieldBytes := newValue.Elem().Field(index).Bytes()
	var temp tempUnionAlias[T, M]
	err = json.Unmarshal(fieldBytes, &temp)
	if err != nil {
		return err
	}

	result := reflect.ValueOf(structPointer)
	elem := result.Elem()
	for i := 0; i < fieldCount; i++ {
		if i == index {
			elem.Field(i).Set(reflect.ValueOf(temp.out))
		} else {
			elem.Field(i).Set(newValue.Elem().Field(i))
		}
	}
	return nil
}

func (f *UnionList[T, M]) UnmarshalJSON(data []byte) error {
	factTypes := []tempUnionAlias[T, M]{}
	err := json.Unmarshal(data, &factTypes)
	if err != nil {
		return err
	}
	for _, factType := range factTypes {
		*f = append(*f, factType.out)
	}
	return nil
}

type tempUnion[T Tag] struct {
	Tag T
	out any
}

type tempUnionAlias[T Tag, M Mapper[T]] tempUnion[T]

func (f *tempUnionAlias[T, M]) UnmarshalJSON(b []byte) error {
	var mapper M

	// Umarshal as tempUnion once to get type, the cast/alias is used to avoid infinite recursion.

	originalType := reflect.TypeOf((*tempUnion[T])(f)).Elem()
	fields := []reflect.StructField{
		originalType.Field(0),
		originalType.Field(1),
	}
	jsonTag := `json:"type"`
	if tagger, ok := any(mapper).(Tagger); ok {
		jsonTag = tagger.JSONTag()
	}
	fields[0].Tag = reflect.StructTag(jsonTag)
	newType := reflect.StructOf(fields)

	temp := reflect.New(newType)
	i := temp.Interface()

	err := json.Unmarshal(b, i)
	if err != nil {
		return err
	}

	// Unmarshal to the actual type.
	tag := temp.Elem().Field(0).Interface().(T)
	f.out, err = mapper.Unmarshal(b, tag)
	return err
}
