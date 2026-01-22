package js

import (
	"errors"

	"github.com/akshayganeshen/napi-go"
)

var (
	ErrWrongType      = errors.New("wrong type")
	ErrBigintLostData = errors.New("bigint convertion lost data")
)

type AnyValue interface {
	GetValue() Value
}

type Value struct {
	Env   Env
	Value napi.Value
}

func (e Env) WrapValue(value napi.Value) Value {
	return Value{
		Env:   e,
		Value: value,
	}
}

func (v Value) GetEnv() Env {
	return v.Env
}

func (v Value) GetValue() Value {
	return v
}

func (v Value) Valid() bool {
	return v.Value != nil
}

func (v Value) GetType() (napi.ValueType, error) {
	vt, st := napi.Typeof(v.Env.Env, v.Value)
	if err := st.AsError(); err != nil {
		return napi.ValueTypeUndefined, err
	}

	return vt, nil
}

func (v Value) Type() napi.ValueType {
	vt, err := v.GetType()
	if err != nil {
		panic(err)
	}

	return vt
}

func (v Value) IsTruthy() (bool, error) {
	boolValue, err := v.IntoBool()
	if err != nil {
		return false, err
	}

	return boolValue.AsBool()
}

func (v Value) IsFalsy() (bool, error) {
	truthy, err := v.IsTruthy()
	if err != nil {
		return false, err
	}

	return !truthy, nil
}

func (v Value) IsUndefined() (bool, error) {
	t, err := v.GetType()
	if err != nil {
		return false, err
	}

	return t == napi.ValueTypeUndefined, nil
}

func (v Value) IsBool() (bool, error) {
	t, err := v.GetType()
	if err != nil {
		return false, err
	}

	return t == napi.ValueTypeBoolean, nil
}

func (v Value) AsBool() (bool, error) {
	if ok, err := v.IsBool(); err != nil {
		return false, err
	} else if !ok {
		return false, ErrWrongType
	}

	b, st := napi.GetValueBool(v.Env.Env, v.Value)
	if err := st.AsError(); err != nil {
		return false, err
	}

	return b, nil
}

func (v Value) IsNumber() (bool, error) {
	t, err := v.GetType()
	if err != nil {
		return false, err
	}

	return t == napi.ValueTypeNumber, nil
}

func (v Value) IsBigint() (bool, error) {
	t, err := v.GetType()
	if err != nil {
		return false, err
	}

	return t == napi.ValueTypeBigint, nil
}

func (v Value) AsInt64() (int64, error) {
	if ok, err := v.IsNumber(); err != nil {
		return 0, err
	} else if ok {
		n, st := napi.GetValueInt64(v.Env.Env, v.Value)
		if err := st.AsError(); err != nil {
			return 0, err
		}

		return n, nil
	}

	if ok, err := v.IsBigint(); err != nil {
		return 0, err
	} else if !ok {
		return 0, ErrWrongType
	}

	n, lossless, st := napi.GetValueBigintInt64(v.Env.Env, v.Value)
	if err := st.AsError(); err != nil {
		return 0, err
	}

	if !lossless {
		return n, ErrBigintLostData
	}

	return n, nil
}

func (v Value) IsString() (bool, error) {
	t, err := v.GetType()
	if err != nil {
		return false, err
	}

	return t == napi.ValueTypeString, nil
}

func (v Value) AsString() (string, error) {
	if ok, err := v.IsString(); err != nil {
		return "", err
	} else if !ok {
		return "", ErrWrongType
	}

	str, st := napi.GetValueStringUtf8(v.Env.Env, v.Value)
	if err := st.AsError(); err != nil {
		return "", err
	}

	return str, nil
}

func (v Value) IntoBool() (Value, error) {
	result, st := napi.CoerceToBool(v.Env.Env, v.Value)
	if err := st.AsError(); err != nil {
		return Value{}, err
	}

	return v.Env.WrapValue(result), nil
}

func (v Value) IntoNumber() (Value, error) {
	result, st := napi.CoerceToNumber(v.Env.Env, v.Value)
	if err := st.AsError(); err != nil {
		return Value{}, err
	}

	return v.Env.WrapValue(result), nil
}

func (v Value) IntoObject() (Value, error) {
	result, st := napi.CoerceToObject(v.Env.Env, v.Value)
	if err := st.AsError(); err != nil {
		return Value{}, err
	}

	return v.Env.WrapValue(result), nil
}

func (v Value) IntoString() (Value, error) {
	result, st := napi.CoerceToString(v.Env.Env, v.Value)
	if err := st.AsError(); err != nil {
		return Value{}, err
	}

	return v.Env.WrapValue(result), nil
}

func (v Value) IntoGoString() (string, error) {
	result, err := v.IntoString()
	if err != nil {
		return "", err
	}

	return result.AsString()
}

func (v Value) String() string {
	s, err := v.IntoGoString()
	if err != nil {
		panic(err)
	}

	return s
}
