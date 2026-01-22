package js

import (
	"github.com/akshayganeshen/napi-go"
)

type Ref struct {
	Env       Env
	Reference napi.Ref
}

func (v Value) NewRef() (Ref, error) {
	ref, st := napi.CreateReference(v.Env.Env, v.Value, 1)
	if err := st.AsError(); err != nil {
		return Ref{}, err
	}

	return Ref{
		Env:       v.Env,
		Reference: ref,
	}, nil
}

func (r Ref) Valid() bool {
	return r.Reference != nil
}

func (r Ref) GetValue() (Value, error) {
	value, st := napi.GetReferenceValue(r.Env.Env, r.Reference)
	if err := st.AsError(); err != nil {
		return Value{}, err
	}

	return r.Env.WrapValue(value), nil
}

func (r Ref) Ref() (int, error) {
	n, st := napi.ReferenceRef(r.Env.Env, r.Reference)
	if err := st.AsError(); err != nil {
		return 0, err
	}

	return n, nil
}

func (r Ref) Unref() (int, error) {
	n, st := napi.ReferenceUnref(r.Env.Env, r.Reference)
	if err := st.AsError(); err != nil {
		return 0, err
	}

	if n == 0 {
		napi.DeleteReference(r.Env.Env, r.Reference)
	}

	return n, nil
}
