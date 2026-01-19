package js

import (
	"github.com/akshayganeshen/napi-go"
)

type ValueType napi.ValueType

func (t ValueType) String() string {
	switch napi.ValueType(t) {
	case napi.ValueTypeUndefined:
		return "undefined"
	case napi.ValueTypeNull:
		return "null"
	case napi.ValueTypeBoolean:
		return "boolean"
	case napi.ValueTypeNumber:
		return "number"
	case napi.ValueTypeString:
		return "string"
	case napi.ValueTypeSymbol:
		return "symbol"
	case napi.ValueTypeObject:
		return "object"
	case napi.ValueTypeFunction:
		return "function"
	case napi.ValueTypeExternal:
		return "external"
	case napi.ValueTypeBigint:
		return "bigint"

	default:
		return "other"
	}
}

type Value struct {
	Env   Env
	Value napi.Value
}

func (v Value) GetEnv() Env {
	return v.Env
}

func (v Value) Type() ValueType {
	vt, st := napi.Typeof(v.Env.Env, v.Value)
	if st != napi.StatusOK {
		panic(napi.StatusError(st))
	}

	return ValueType(vt)
}
