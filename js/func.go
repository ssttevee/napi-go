package js

import (
	"github.com/akshayganeshen/napi-go"
)

type Function struct {
	Value
}

func (v Value) IsFunction() (bool, error) {
	t, err := v.GetType()
	if err != nil {
		return false, err
	}

	return t == napi.ValueTypeFunction, nil
}

func (v Value) AsFunctionUnsafe() Function {
	return Function{
		Value: v,
	}
}

func (v Value) AsFunction() (Function, error) {
	if ok, err := v.IsFunction(); err != nil {
		return Function{}, err
	} else if !ok {
		return Function{}, ErrWrongType
	}

	return v.AsFunctionUnsafe(), nil
}

func (e Env) NewFunction(fn any) (Function, error) {
	cb, err := Callback(fn)
	if err != nil {
		return Function{}, err
	}

	// TODO: Add CreateReference to FuncOf to keep value alive
	v, st := napi.CreateFunction(
		e.Env,
		"",
		cb,
	)

	if err := st.AsError(); err != nil {
		return Function{}, err
	}

	return Function{
		Value: e.WrapValue(v),
	}, nil
}

func (f Function) Call(this any, args ...any) (Value, error) {
	thisValue, err := f.Env.ValueOf(this)
	if err != nil {
		return Value{}, err
	}

	argValues := make([]napi.Value, len(args))
	for i, arg := range args {
		value, err := f.Env.ValueOf(arg)
		if err != nil {
			return Value{}, err
		}

		argValues[i] = value.Value
	}

	result, st := napi.CallFunction(f.Env.Env, thisValue.Value, f.Value.Value, argValues)
	if err := st.AsError(); err != nil {
		return Value{}, err
	}

	return f.Env.WrapValue(result), nil
}

type Finalizer interface {
	Finalize(env Env, data any)
}

type FinalizerFunc func(env Env, data any)

func (f FinalizerFunc) Finalize(env Env, data any) {
	f(env, data)
}

func finalizerWrapper(env napi.Env, data any, hint any) {
	hint.(Finalizer).Finalize(WrapEnv(env), data)
}
