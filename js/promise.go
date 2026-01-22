package js

import (
	"github.com/akshayganeshen/napi-go"
)

type Promise struct {
	Value
}

func (v Value) IsPromise() (bool, error) {
	b, st := napi.IsPromise(v.Env.Env, v.Value)
	if err := st.AsError(); err != nil {
		return false, err
	}

	return b, nil
}

func (v Value) AsPromiseUnsafe() Promise {
	return Promise{
		Value: v,
	}
}

func (v Value) AsPromise() (Promise, error) {
	if ok, err := v.IsPromise(); err != nil {
		return Promise{}, err
	} else if !ok {
		return Promise{}, ErrWrongType
	}

	return v.AsPromiseUnsafe(), nil
}
