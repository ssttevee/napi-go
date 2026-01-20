package js

import (
	"github.com/akshayganeshen/napi-go"
)

type Value struct {
	Env   Env
	Value napi.Value
}

func (e Env) AsValue(value napi.Value) Value {
	return Value{
		Env:   e,
		Value: value,
	}
}

func (v Value) GetEnv() Env {
	return v.Env
}

func (v Value) Type() napi.ValueType {
	vt, st := napi.Typeof(v.Env.Env, v.Value)
	if st != napi.StatusOK {
		panic(napi.StatusError(st))
	}

	return vt
}
