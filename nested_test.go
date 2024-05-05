package jtug

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

type TestTag string

const (
	TagA TestTag = "A"
	TagB TestTag = "B"
	TagC TestTag = "C"
)

type TestUnion = Union[TestTag]
type TestUnionList = UnionList[TestTag, TestMapper]

type TestMapper struct{}

func (TestMapper) Unmarshal(b []byte, t TestTag) (Union[TestTag], error) {
	switch t {
	case TagA:
		var value TypeA
		return value, json.Unmarshal(b, &value)
	case TagB:
		var value TypeB
		return value, json.Unmarshal(b, &value)
	case TagC:
		var value TypeC
		return value, json.Unmarshal(b, &value)
	default:
		return nil, fmt.Errorf("unknown tag: \"%s\"", t)
	}
}

func (TestMapper) JSONTag() string {
	return `json:"tag"`
}

type TypeA struct {
	Tag  TestTag `json:"tag"`
	Data int     `json:"data"`
}

type TypeB struct {
	Tag   TestTag   `json:"tag"`
	Child TestUnion `json:"child"`
}

func (t *TypeB) UnmarshalJSON(b []byte) error {
	return UnmarshalTaggedField[TestMapper](t, &t.Child, b)
}

type TypeC struct {
	Tag      TestTag       `json:"tag"`
	Children TestUnionList `json:"children"`
}

type TestWrapper struct {
	List TestUnionList `json:"list"`
	One  TestUnion     `json:"one"`
}

func (t *TestWrapper) UnmarshalJSON(b []byte) error {
	return UnmarshalTaggedField[TestMapper](t, &t.One, b)
}

func Must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func TestMarshalUnmarshal(t *testing.T) {
	fileContent := Must(os.ReadFile("test.json"))
	var wrapper TestWrapper
	Must(0, json.Unmarshal(fileContent, &wrapper))

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	Must(0, encoder.Encode(wrapper))

	if buf.String() != string(fileContent) {
		t.Log("want:")
		t.Log(string(fileContent))
		t.Log("have:")
		t.Log(buf.String())
		panic("output does not equal input")
	}
}
