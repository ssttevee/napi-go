package js

import (
	"github.com/akshayganeshen/napi-go"
)

type Object struct {
	Value
}

func (v Value) IsObject() (bool, error) {
	t, err := v.GetType()
	if err != nil {
		return false, err
	}

	return t == napi.ValueTypeObject, nil
}

func (v Value) AsObjectUnsafe() Object {
	return Object{
		Value: v,
	}
}

func (v Value) AsObject() (Object, error) {
	if ok, err := v.IsObject(); err != nil {
		return Object{}, err
	} else if !ok {
		return Object{}, ErrWrongType
	}

	return v.AsObjectUnsafe(), nil
}

func (e Env) NewObject() (Object, error) {
	v, st := napi.CreateObject(e.Env)
	if err := st.AsError(); err != nil {
		return Object{}, err
	}

	return Object{
		Value: e.WrapValue(v),
	}, nil
}

func (o Object) HasProperty(key Value) (bool, error) {
	b, st := napi.HasProperty(o.Env.Env, o.Value.Value, key.Value)
	if err := st.AsError(); err != nil {
		return false, err
	}

	return b, nil
}

func (o Object) HasOwnProperty(key Value) (bool, error) {
	b, st := napi.HasOwnProperty(o.Env.Env, o.Value.Value, key.Value)
	if err := st.AsError(); err != nil {
		return false, err
	}

	return b, nil
}

func (o Object) Get(key Value) (Value, error) {
	result, st := napi.GetProperty(o.Env.Env, o.Value.Value, key.Value)
	if err := st.AsError(); err != nil {
		return Value{}, err
	}

	return o.Env.WrapValue(result), nil
}

func (o Object) GetNamed(name string) (Value, error) {
	nameValue, err := o.Env.ValueOf(name)
	if err != nil {
		return Value{}, nil
	}

	return o.Get(nameValue)
}

func (o Object) Set(key, value Value) error {
	return napi.SetProperty(o.Env.Env, o.Value.Value, key.Value, value.Value).AsError()
}

func (o Object) CallNamed(name string, args ...any) (Value, error) {
	nameValue, err := o.Env.ValueOf(name)
	if err != nil {
		return Value{}, nil
	}

	return o.Call(nameValue, args...)
}

func (o Object) Call(key Value, args ...any) (Value, error) {
	prop, err := o.Get(key)
	if err != nil {
		return Value{}, err
	}

	f, err := prop.AsFunction()
	if err != nil {
		return Value{}, err
	}

	return f.Call(o, args...)
}
