JTUG - JSON Tagged Union Go
===========================

Deserialize (unmarshal) tagged unions from JSON into go structs.
----------------------------------------------------------------

You have some JSON with a tagged union:

```json
[
  { "type": "A", "count": 10 },
  { "type": "B", "data": "hello" }
]
```

You have some go:

```go
type Tag string

const (
	A Tag = "A"
	B Tag = "B"
)

type StructA struct {
	Type  Tag `json:"type"`
	Count int `json:"count"`
}

type StructB struct {
	Type Tag    `json:"type"`
	Data string `json:"data"`
}
```

Define some types and write a tag-to-struct mapper:

```go
type Union = jtug.Union[Tag]
type List = jtug.UnionList[Tag, Mapper]

type Mapper struct{}

func (Mapper) Unmarshal(b []byte, t Tag) (jtug.Union[Tag], error) {
	switch t {
	case A:
		var value StructA
		return value, json.Unmarshal(b, &value)
	case B:
		var value StructB
		return value, json.Unmarshal(b, &value)
	default:
		return nil, fmt.Errorf("unknown tag: \"%s\"", t)
	}
}
```

Now you can parse for example a list of tags:

```go
var list List
err := json.Unmarshal([]byte(`[
    {"type":"A","count":10},
    {"type":"B","data":"hello"}
]`), &list)
for i := range list {
    switch t := list[i].(type) {
        case StructA:
            println(t.Count)
        case StructB:
            println(t.Data)
        // etc.
    }
}
```

You can also parse some wrapping container struct:

```go
type Container struct {
    Target Union `json:"target"`
}

func (t *Container) UnmarshalJSON(b []byte) error {
    return jtug.UnmarshalTaggedField[Mapper](t, &t.Target, b)
}

var container Container
err := json.Unmarshal([]byte(`{"target": {"type":"A","count":10}}`), &container)
switch t := container.Target.(type) {
    case StructA:
    // ...
}
```

If your tag field is not "type", then the Mapper can implement `JSONTag() string` (it must return the whole go tag, for example `json:"mytag"`).

Serialize (marshal)
-------------------

Just serialize your structs normally.

Nested structs
--------------

It is possible for your struct types to have fields that are tagged unions themselves.
See the `nested_test.go` for an example.

Advantages of `jtug`
--------------------

* Zero dependencies.
* Limited amount of boilerplate, only mapper and some aliases.
* The JSON format works well with TypeScript and Python.

Limitations
-----------

Go does not have tagged unions, sum types, or real enums.
Every way to represent those concepts in go has some imperfections.
Interfaces are the blessed way to do polymorphism in go.

Cannot unmarshal a root level tagged union
------------------------------------------

```go
var item jtug.Union[Tag]
json.Unmarshal(`{"type":"A"}`, &item)
```

AFAIK it is impossible to directly unmarshal to an interface, it must be wrapped by a container struct that implements `UnmarshalJSON([]byte) error`, where you can use the `jtug.UnmarshalTaggedField[Mapper](t, &t.TargetField, b)` function.

The container can also be a slice, like in the first example above, without needing to do this.

UnmarshalJSON is already implemented
------------------------------------

If you must have custom unmarshaling on the structs for some reason, then you would probably need to copy some code riddled with `reflect` from the library.
