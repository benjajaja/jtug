package jtug

import (
	"encoding/json"
	"fmt"
	"testing"
)

type SimpleTag string

const (
	A SimpleTag = "A"
	B SimpleTag = "B"
)

type SimpleA struct {
	Type  TestTag `json:"type"`
	Count int     `json:"count"`
}

type SimpleB struct {
	Type TestTag `json:"type"`
	Data string  `json:"data"`
}

type SimpleUnion = Union[SimpleTag]
type SimpleUnions = UnionList[SimpleTag, SimpleMapper]

type SimpleMapper struct{}

func (SimpleMapper) Unmarshal(b []byte, t SimpleTag) (Union[SimpleTag], error) {
	switch t {
	case A:
		var value SimpleA
		return value, json.Unmarshal(b, &value)
	case B:
		var value SimpleB
		return value, json.Unmarshal(b, &value)
	default:
		return nil, fmt.Errorf("unknown tag: \"%s\"", t)
	}
}

func TestList(t *testing.T) {
	var list SimpleUnions
	err := json.Unmarshal([]byte(`[
		{"type":"A","count":10},
		{"type":"B","data":"hello"}
	]`), &list)
	if err != nil {
		panic(err)
	}
	if list[0].(SimpleA).Count != 10 {
		panic("did not Unmarshal")
	}
	if list[1].(SimpleB).Data != "hello" {
		panic("did not Unmarshal")
	}
}
