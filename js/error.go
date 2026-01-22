package js

import "github.com/akshayganeshen/napi-go"

type Error struct {
	Value
}

func (v Value) IsError() (bool, error) {
	b, st := napi.IsError(v.Env.Env, v.Value)
	if err := st.AsError(); err != nil {
		return false, err
	}

	return b, nil
}

func (v Value) AsErrorUnsafe() Error {
	return Error{
		Value: v,
	}
}

func (v Value) AsError() (Error, error) {
	if ok, err := v.IsError(); err != nil {
		return Error{}, err
	} else if !ok {
		return Error{}, ErrWrongType
	}

	return v.AsErrorUnsafe(), nil
}

func (e Env) NewError(code string, message string) (Error, error) {
	var codeValue napi.Value
	if code != "" {
		value, err := e.ValueOf(code)
		if err != nil {
			return Error{}, err
		}

		codeValue = value.Value
	}

	msgValue, err := e.ValueOf(message)
	if err != nil {
		return Error{}, err
	}

	v, st := napi.CreateError(e.Env, codeValue, msgValue.Value)
	if err := st.AsError(); err != nil {
		return Error{}, err
	}

	return Error{
		Value: e.WrapValue(v),
	}, nil
}

func (e Env) ThrowError(code string, message string) error {
	jsErr, err := e.NewError(code, message)
	if err != nil {
		return err
	}

	return jsErr.Throw()
}

func (e Error) Throw() error {
	return napi.Throw(e.Env.Env, e.Value.Value).AsError()
}

type ErrorRef struct {
	Ref
}

func (e Error) NewRef() (ErrorRef, error) {
	ref, err := e.Value.NewRef()
	if err != nil {
		return ErrorRef{}, err
	}

	return ErrorRef{
		Ref: ref,
	}, nil
}

func (e ErrorRef) Error() string {
	value, err := e.Ref.GetValue()
	if err != nil {
		panic(err)
	}

	return value.AsErrorUnsafe().String()
}
