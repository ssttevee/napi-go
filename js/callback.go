package js

import (
	"fmt"
	"reflect"

	"github.com/akshayganeshen/napi-go"
)

var (
	errorInterface = reflect.TypeOf((*error)(nil)).Elem()

	envType      = reflect.TypeOf(Env{})
	valueType    = reflect.TypeOf(Value{})
	stringType   = reflect.TypeOf("")
	objectType   = reflect.TypeOf(Object{})
	bufferType   = reflect.TypeOf(Buffer{})
	functionType = reflect.TypeOf(Function{})
	promiseType  = reflect.TypeOf(Promise{})
	errorType    = reflect.TypeOf(Error{})
)

func MustCallback(fn any) napi.Callback {
	cb, err := Callback(fn)
	if err != nil {
		panic(err)
	}

	return cb
}

func Callback(fn any) (napi.Callback, error) {
	if cb, ok := fn.(napi.Callback); ok {
		return cb, nil
	}

	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	// Validate that fn is a function
	if fnType.Kind() != reflect.Func {
		return nil, fmt.Errorf("AsCallback: expected function, got %v", fnType)
	}

	// Validate return type: must be any, (any, error), or concrete types
	numOut := fnType.NumOut()
	if numOut > 2 {
		return nil, fmt.Errorf("AsCallback: function must return any or (any, error), got %d return values", numOut)
	}
	if numOut == 2 {
		// Second return must be error
		if !fnType.Out(1).Implements(errorInterface) {
			return nil, fmt.Errorf("AsCallback: second return value must be error, got %v", fnType.Out(1))
		}
	}

	// Validate parameters
	numIn := fnType.NumIn()
	if numIn == 0 {
		return nil, fmt.Errorf("AsCallback: function must have at least a 'this' parameter")
	}

	paramIdx := 0
	hasEnv := false
	hasVariadic := fnType.IsVariadic()

	// Check for optional Env parameter
	if fnType.In(0) == envType {
		hasEnv = true
		paramIdx++
	}

	// Check for required 'this' parameter
	if paramIdx >= numIn {
		return nil, fmt.Errorf("AsCallback: function must have a 'this' parameter")
	}

	if err := validateCallbackArgType(fnType.In(paramIdx)); err != nil {
		return nil, fmt.Errorf("AsCallback: 'this' parameter type %w", err)
	}
	paramIdx++

	// Validate remaining parameters
	if paramIdx < numIn {
		// Check if it's a single slice parameter
		if numIn == paramIdx+1 && fnType.In(paramIdx).Kind() == reflect.Slice {
			if err := validateCallbackArgType(fnType.In(paramIdx).Elem()); err != nil {
				return nil, fmt.Errorf("AsCallback: slice element type %w", err)
			}
		} else {
			// Multiple individual parameters
			for i := paramIdx; i < numIn; i++ {
				paramType := fnType.In(i)
				if hasVariadic && i == numIn-1 {
					// For variadic functions, check the slice element type
					if paramType.Kind() != reflect.Slice {
						panic(fmt.Sprintf("AsCallback: variadic parameter must be a slice, got %v", paramType))
					}
					paramType = paramType.Elem()
				}
				if err := validateCallbackArgType(paramType); err != nil {
					return nil, fmt.Errorf("AsCallback: parameter %d %w", i, err)
				}
			}
		}
	}

	// Create the actual callback
	return func(env napi.Env, info napi.CallbackInfo) napi.Value {
		jsEnv := WrapEnv(env)

		cbInfo, st := napi.GetCbInfo(env, info)
		if st != napi.StatusOK {
			napi.ThrowError(env, "", napi.StatusError(st).Error())
			undef, _ := jsEnv.Undefined()
			return undef.Value
		}

		thisValue := Value{
			Env:   jsEnv,
			Value: cbInfo.This,
		}
		args := make([]Value, len(cbInfo.Args))
		for i, cbArg := range cbInfo.Args {
			args[i] = Value{
				Env:   jsEnv,
				Value: cbArg,
			}
		}

		// Build call arguments
		callArgs := make([]reflect.Value, 0, numIn)

		paramIdx := 0

		// Add Env if needed
		if hasEnv {
			callArgs = append(callArgs, reflect.ValueOf(jsEnv))
			paramIdx++
		}

		// Add 'this'
		thisType := fnType.In(paramIdx)
		convertedThisArg, ok := convertCallbackArgType(thisValue, thisType)
		if !ok {
			napi.ThrowError(env, "", fmt.Sprintf("this: expected %v, got incompatible type: %s", thisType, thisValue.Type()))
			undef, _ := jsEnv.Undefined()
			return undef.Value
		}
		callArgs = append(callArgs, convertedThisArg)
		paramIdx++

		// Add remaining arguments
		if paramIdx < numIn {
			// Check if it's a single slice parameter
			if numIn == paramIdx+1 && fnType.In(paramIdx).Kind() == reflect.Slice {
				// Single slice parameter
				sliceType := fnType.In(paramIdx)
				elemType := sliceType.Elem()
				slice := reflect.MakeSlice(sliceType, len(args), len(args))
				for i, arg := range args {
					convertedArg, ok := convertCallbackArgType(arg, elemType)
					if !ok {
						napi.ThrowError(env, "", fmt.Sprintf("Argument %d: expected %v, got incompatible type: %s", i, fnType.In(paramIdx), arg.Type()))
						undef, _ := jsEnv.Undefined()
						return undef.Value
					}
					slice.Index(i).Set(convertedArg)
				}
				callArgs = append(callArgs, slice)
			} else {
				// Multiple individual parameters
				argsNeeded := numIn - paramIdx
				if hasVariadic {
					argsNeeded = numIn - paramIdx - 1
				}

				for i := 0; i < argsNeeded && paramIdx < numIn; i++ {
					if i < len(args) {
						convertedArg, ok := convertCallbackArgType(args[i], fnType.In(paramIdx))
						if !ok {
							napi.ThrowError(env, "", fmt.Sprintf("Argument %d: expected %v, got incompatible type: %s", i, fnType.In(paramIdx), args[i].Type()))
							undef, _ := jsEnv.Undefined()
							return undef.Value
						}
						callArgs = append(callArgs, convertedArg)
					} else {
						if hasVariadic {
							napi.ThrowError(env, "", fmt.Sprintf("Expected at least %d argument(s), got %d", argsNeeded, len(args)))
						} else {
							napi.ThrowError(env, "", fmt.Sprintf("Expected %d argument(s), got %d", argsNeeded, len(args)))
						}
						undef, _ := jsEnv.Undefined()
						return undef.Value
					}
					paramIdx++
				}

				// Handle variadic arguments
				if hasVariadic {
					variadicType := fnType.In(numIn - 1).Elem()
					remainingArgs := args[argsNeeded:]
					for i, arg := range remainingArgs {
						convertedArg, ok := convertCallbackArgType(arg, variadicType)
						if !ok {
							napi.ThrowError(env, "", fmt.Sprintf("Argument %d: expected %v, got incompatible type: %s", argsNeeded+i, variadicType, arg.Type()))
							undef, _ := jsEnv.Undefined()
							return undef.Value
						}
						callArgs = append(callArgs, convertedArg)
					}
				}
			}
		}

		// Call the function
		results := fnValue.Call(callArgs)
		if len(results) == 0 {
			undef, _ := jsEnv.Undefined()
			return undef.Value
		}

		// Check for error (if two return values)
		if len(results) == 2 {
			if errVal := results[1]; !errVal.IsNil() {
				napi.ThrowError(env, "", errVal.Interface().(error).Error())
				undef, _ := jsEnv.Undefined()
				return undef.Value
			}
		}

		// Return the first result
		result, err := jsEnv.ValueOf(results[0].Interface())
		if err != nil {
			napi.ThrowError(env, "", err.Error())
			undef, _ := jsEnv.Undefined()
			return undef.Value
		}

		return result.Value
	}, nil
}

func validateCallbackArgType(targetType reflect.Type) error {
	switch targetType {
	default:
		return fmt.Errorf("must be Value, string, Object, Buffer, Function, Promise, or Error but got %v", targetType)

	case valueType:
	case stringType:
	case objectType:
	case bufferType:
	case functionType:
	case promiseType:
	case errorType:
	}

	return nil
}

func convertCallbackArgType(val Value, targetType reflect.Type) (reflect.Value, bool) {
	switch targetType {
	case valueType:
		return reflect.ValueOf(val), true

	case stringType:
		if s, err := val.AsString(); err == nil {
			return reflect.ValueOf(s), true
		}

	case objectType:
		if obj, err := val.AsObject(); err == nil {
			return reflect.ValueOf(obj), true
		}

	case bufferType:
		if buf, err := val.AsBuffer(); err == nil {
			return reflect.ValueOf(buf), true
		}

	case functionType:
		if fn, err := val.AsFunction(); err == nil {
			return reflect.ValueOf(fn), true
		}

	case promiseType:
		if fn, err := val.AsPromise(); err == nil {
			return reflect.ValueOf(fn), true
		}

	case errorType:
		if fn, err := val.AsError(); err != nil {
			return reflect.ValueOf(fn), true
		}
	}

	return reflect.Value{}, false
}
