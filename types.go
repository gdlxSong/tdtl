/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package tdtl

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var (
	UNDEFINED_RESULT = &DefaultNode{typ: Undefined}
	NULL_RESULT      = &DefaultNode{typ: Undefined}
)

// Type node type
type Type int

const (
	// Undefine is Not a value
	// This isn't explicitly representable in JSON except by omitting the value.
	Undefined Type = iota
	// Null is a null json value
	Null
	// Bool is a json boolean
	Bool
	// Number is json number, include Int and Float
	Number
	// Int is json number, a discrete Int
	Int
	// Float is json number
	Float
	// String is a json string
	String
	// Array is a json string
	Array
	// JSON is a raw block of JSON
	JSON
)

// String returns a string representation of the type.
func (t Type) String() string {
	switch t {
	default:
		return "Undefined"
	case Null:
		return "Null"
	case Bool:
		return "Bool"
	case Int:
		return "Int"
	case Float:
		return "Float"
	case String:
		return "String"
	case JSON:
		return "JSON"
	}
}

//Node interface
type Node interface {
	Type() Type
	To(Type) Node
	String() string
	Value() interface{}
}

//DefaultNode interface
type DefaultNode struct {
	// Type is the json type
	typ Type
	// raw is the raw json
	raw string
}

func (r DefaultNode) Type() Type { return r.typ }
func (r DefaultNode) To(Type) Node {
	return r
}
func (r DefaultNode) String() string {
	return r.raw
}
func (r DefaultNode) Value() interface{} { return r.raw }

type BoolNode bool

func (r BoolNode) Type() Type         { return Bool }
func (r BoolNode) Value() interface{} { return bool(r) }
func (r BoolNode) To(typ Type) Node {
	switch typ {
	case Bool:
		return r
	case String:
		return StringNode(fmt.Sprintf("%t", r))
	}
	return UNDEFINED_RESULT
}
func (r BoolNode) String() string {
	return fmt.Sprintf("%t", r)
}

type IntNode int64

func (r IntNode) Type() Type         { return Int }
func (r IntNode) Value() interface{} { return int64(r) }
func (r IntNode) To(typ Type) Node {
	switch typ {
	case Number, Int:
		return r
	case Float:
		return FloatNode(r)
	case String:
		return StringNode(fmt.Sprintf("%d", r))
	}
	return UNDEFINED_RESULT
}
func (r IntNode) String() string {
	return fmt.Sprintf("%d", r)
}

type FloatNode float64

func (r FloatNode) Type() Type         { return Float }
func (r FloatNode) Value() interface{} { return float64(r) }
func (r FloatNode) To(typ Type) Node {
	switch typ {
	case Number, Float:
		return r
	case Int:
		return IntNode(r)
	case String:
		return StringNode(fmt.Sprintf("%f", r))
	}
	return UNDEFINED_RESULT
}
func (r FloatNode) String() string {
	return fmt.Sprintf("%f", r)
}

type StringNode string

func (r StringNode) Type() Type         { return String }
func (r StringNode) Value() interface{} { return string(r) }
func (r StringNode) To(typ Type) Node {
	switch typ {
	case String:
		return r
	case Bool:
		b, err := strconv.ParseBool(string(r))
		if err != nil {
			return UNDEFINED_RESULT
		}
		return BoolNode(b)
	case Number:
		if !strings.Contains(string(r), ".") {
			return r.To(Int)
		}
		return r.To(Float)
	case Int:
		b, err := strconv.ParseInt(string(r), 10, 64)
		if err != nil {
			return UNDEFINED_RESULT
		}
		return IntNode(b)
	case Float:
		b, err := strconv.ParseFloat(string(r), 64)
		if err != nil {
			return UNDEFINED_RESULT
		}
		return FloatNode(b)
	}
	return UNDEFINED_RESULT
}
func (r StringNode) String() string {
	return string(r)
}

type NullNode struct{}

func (r NullNode) Type() Type         { return Null }
func (r NullNode) String() string     { return "null" }
func (r NullNode) Value() interface{} { return nil }
func (r NullNode) To(typ Type) Node {
	switch typ {
	case Null:
		return r
	case JSON:
		return JSONNode("{}")
	case Array:
		return ArrayNode("[]")
	default:
		return UNDEFINED_RESULT
	}
}

type ArrayNode []byte

func (r ArrayNode) Type() Type     { return Array }
func (r ArrayNode) String() string { return string(r) }
func (r ArrayNode) Value() interface{} {
	var data interface{}
	_ = json.Unmarshal(r, &data)
	return data
}

func (r ArrayNode) To(typ Type) Node {
	switch typ {
	case String:
		return StringNode(r)
	case Array:
		return r
	case JSON:
		return JSONNode(r)
	default:
		return UNDEFINED_RESULT
	}
}

// JSONNode maybe Object or Array
type JSONNode string

func (r JSONNode) Type() Type { return JSON }
func (r JSONNode) To(typ Type) Node {
	return UNDEFINED_RESULT
}
func (r JSONNode) Value() interface{} {
	var data interface{}
	_ = json.Unmarshal([]byte(r), &data)
	return data
}
func (r JSONNode) Update(key string, value Node) (val string, err error) {
	switch value := value.(type) {
	case FloatNode, IntNode, BoolNode:
		v := value.To(String)
		switch v := v.(type) {
		case StringNode:
			val, err = updateJSON(r, key, v)
		}
	case StringNode:
		val, err = updateJSON(r, key, "\""+value+"\"")
	case JSONNode:
		if key == "" {
			val = string(value)
		} else {
			val, err = updateJSON(r, key, StringNode(value))
		}
	default:
		val, err = "", fmt.Errorf("unknown type")
	}
	return
}
func (r JSONNode) String() string {
	return string(r)
}

//Expr
type Expr interface {
	expr()
}

func (*SelectStatementExpr) expr() {}
func (FieldsExpr) expr()           {}
func (*FieldExpr) expr()           {}
func (TopicExpr) expr()            {}
func (*FilterExpr) expr()          {}
func (*BinaryExpr) expr()          {}
func (*JSONPathExpr) expr()        {}
func (*SwitchExpr) expr()          {}
func (CaseListExpr) expr()         {}
func (*CaseExpr) expr()            {}

func (DefaultNode) expr() {}
func (BoolNode) expr()    {}
func (IntNode) expr()     {}
func (FloatNode) expr()   {}
func (StringNode) expr()  {}
func (*CallExpr) expr()   {}
func (JSONNode) expr()    {}

//BinaryExpr
type BinaryExpr struct {
	Op  int
	LHS Expr
	RHS Expr
}

//JSONPathExpr xpath
type JSONPathExpr struct {
	val string
}

//CallExpr
type CallExpr struct {
	raw  string
	key  string
	args []Expr
}

func (e *CallExpr) String() string {
	return fmt.Sprintf("%s", e.raw)
}

func (e *CallExpr) FuncName() string {
	return fmt.Sprintf("%s", e.key)
}

func (e *CallExpr) Args() []Expr {
	return e.args
}

//SwitchExpr
type SwitchExpr struct {
	exp  Expr
	list []*CaseExpr
	last Expr
}

//CaseListExpr
type CaseListExpr []*CaseExpr

//CaseExpr
type CaseExpr struct {
	when Expr
	then Expr
}

//FieldExpr
type FieldExpr struct {
	exp   Expr
	alias string
}

func (r *FieldExpr) String() string {
	panic("todo")
}

//FieldExpr
type FieldsExpr []*FieldExpr

func (r FieldsExpr) String() string {
	panic("todo")
}

//TopicExpr
type TopicExpr []string

func (r TopicExpr) String() string {
	panic("todo")
}

//FilterExpr
type FilterExpr struct {
	exp Expr
}

func (r *FilterExpr) String() string {
	panic("todo")
}

//GroupExpr
type DimensionsExpr struct {
	exprs  []*JSONPathExpr
	window *WindowExpr
}

func (*DimensionsExpr) expr() {}

type DimensionExpr struct {
	exp Expr
}

func (*DimensionExpr) expr() {}

type WindowType int

const (
	NOT_WINDOW WindowType = iota
	TUMBLING_WINDOW
	HOPPING_WINDOW
	SLIDING_WINDOW
	SESSION_WINDOW
)

type WindowLength int

func (WindowLength) expr() {}

type WindowInterval int

func (WindowInterval) expr() {}

type WindowExpr struct {
	WindowType WindowType
	Length     WindowLength
	Interval   WindowInterval
}

func (WindowExpr) expr() {}

//SelectStatementExpr
type SelectStatementExpr struct {
	fields     FieldsExpr
	topic      TopicExpr
	filter     *FilterExpr
	dimensions *DimensionsExpr
}

func (r *SelectStatementExpr) String() string {
	return "Root Expr"
}

func NewNode(v interface{}) Node {
	switch val := v.(type) {
	case float32:
		return FloatNode(val)
	case float64:
		return FloatNode(val)
	case uint8, int8, uint16, int16, uint, int, uint32, int32, int64, uint64:
		return StringNode(fmt.Sprintf("%v", val)).To(Int)
	case string:
		return StringNode(val)
	case []byte:
		return JSONNode(val)
	case bool:
		return BoolNode(val)
	case map[string]interface{}:
		data, _ := json.Marshal(v)
		return JSONNode(string(data))
	case nil:
		return NullNode{}
	default:
		valKind := reflect.TypeOf(val).Kind()
		if reflect.Ptr == valKind {
			// deference pointer.
			return NewNode(reflect.ValueOf(val).Elem().Interface())
		} else if reflect.Slice == valKind {
			data, _ := json.Marshal(v)
			return JSONNode(string(data))
		}

		return UNDEFINED_RESULT
	}
}
func ToBytesWithWrapString(val Node) []byte {
	if nil == val {
		return []byte{}
	}

	switch val.Type() {
	case JSON:
		jsonVal, _ := val.(JSONNode)
		return []byte(jsonVal)
	case String:
		return []byte("\"" + val.String() + "\"")
	default:
		return []byte(val.String())
	}
}
