package js

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/akshayganeshen/napi-go"
)

type Env struct {
	Env napi.Env
}

func (e Env) Valid() bool {
	return e.Env != nil
}

type InvalidValueTypeError struct {
	Value any
}

var _ error = InvalidValueTypeError{}

func WrapEnv(env napi.Env) Env {
	return Env{
		Env: env,
	}
}

func (e Env) GetGlobal() (Object, error) {
	v, st := napi.GetGlobal(e.Env)
	if err := st.AsError(); err != nil {
		return Object{}, err
	}

	return e.WrapValue(v).AsObjectUnsafe(), nil
}

func (e Env) Null() (Value, error) {
	v, st := napi.GetNull(e.Env)
	if err := st.AsError(); err != nil {
		return Value{}, err
	}

	return e.WrapValue(v), nil
}

func (e Env) Undefined() (Value, error) {
	v, st := napi.GetUndefined(e.Env)
	if err := st.AsError(); err != nil {
		return Value{}, err
	}

	return e.WrapValue(v), nil
}

func (e Env) ValueOf(x any) (Value, error) {
	var (
		v  napi.Value
		st napi.Status
	)

	switch xt := x.(type) {
	case interface{ GetValue() Value }:
		return xt.GetValue(), nil
	case interface{ GetValue() (Value, error) }:
		return xt.GetValue()
	case Value:
		return xt, nil
	case []Value:
		l := len(xt)
		v, st = napi.CreateArrayWithLength(e.Env, l)
		if st != napi.StatusOK {
			break
		}

		for i, xti := range xt {
			// TODO: Use Value.SetIndex helper
			st = napi.SetElement(e.Env, v, i, xti.Value)
			if st != napi.StatusOK {
				break
			}
		}
	case Function:
		return xt.Value, nil
	case napi.Value:
		v, st = xt, napi.StatusOK

	case nil:
		v, st = napi.GetNull(e.Env)
	case bool:
		v, st = napi.GetBoolean(e.Env, xt)
	case int:
		v, st = napi.CreateDouble(e.Env, float64(xt))
	case int8:
		v, st = napi.CreateDouble(e.Env, float64(xt))
	case int16:
		v, st = napi.CreateDouble(e.Env, float64(xt))
	case int64:
		v, st = napi.CreateDouble(e.Env, float64(xt))
	case uint:
		v, st = napi.CreateDouble(e.Env, float64(xt))
	case uint8:
		v, st = napi.CreateDouble(e.Env, float64(xt))
	case uint16:
		v, st = napi.CreateDouble(e.Env, float64(xt))
	case uint64:
		v, st = napi.CreateDouble(e.Env, float64(xt))
	case uintptr:
		v, st = napi.CreateDouble(e.Env, float64(xt))
	case unsafe.Pointer:
		v, st = napi.CreateDouble(e.Env, float64(uintptr(xt)))
	case float32:
		v, st = napi.CreateDouble(e.Env, float64(xt))
	case float64:
		v, st = napi.CreateDouble(e.Env, xt)
	case string:
		v, st = napi.CreateStringUtf8(e.Env, xt)
	case error:
		jsErr, err := e.NewError("", xt.Error())
		if err != nil {
			return Value{}, err
		}

		v = jsErr.Value.Value
	case []any:
		l := len(xt)
		v, st = napi.CreateArrayWithLength(e.Env, l)
		if st != napi.StatusOK {
			break
		}

		for i, xti := range xt {
			// TODO: Use Value.SetIndex helper
			vti, err := e.ValueOf(xti)
			if err != nil {
				return Value{}, err
			}

			st = napi.SetElement(e.Env, v, i, vti.Value)
			if st != napi.StatusOK {
				break
			}
		}

	case map[string]any:
		obj, err := e.NewObject()
		if err != nil {
			return Value{}, err
		}

		for xtk, xtv := range xt {
			// TODO: Use Value.Set helper
			vtk, err := e.ValueOf(xtk)
			if err != nil {
				return Value{}, err
			}

			vtv, err := e.ValueOf(xtv)
			if err != nil {
				return Value{}, err
			}

			if err := obj.Set(vtk, vtv); err != nil {
				return Value{}, err
			}
		}

		v = obj.Value.Value

	default:
		if reflect.ValueOf(x).Kind() == reflect.Func {
			fn, err := e.NewFunction(x)
			if err != nil {
				return Value{}, err
			}

			return fn.Value, nil
		}

		return Value{}, InvalidValueTypeError{x}
	}

	if err := st.AsError(); err != nil {
		return Value{}, err
	}

	return e.WrapValue(v), nil
}

func (e Env) WellKnownSymbol(name string) (Value, error) {
	symbolValue, err := e.ValueOf("Symbol")
	if err != nil {
		return Value{}, err
	}

	nameValue, err := e.ValueOf(name)
	if err != nil {
		return Value{}, err
	}

	global, err := e.GetGlobal()
	if err != nil {
		return Value{}, err
	}

	symbolObj, err := global.Get(symbolValue)
	if err != nil {
		return Value{}, err
	}

	return symbolObj.AsObjectUnsafe().Get(nameValue)
}

func (err InvalidValueTypeError) Error() string {
	return fmt.Sprintf("Value cannot be represented in JS: %T", err.Value)
}
