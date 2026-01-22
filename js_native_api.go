package napi

/*
#include <stdlib.h>
#include <node/node_api.h>
*/
import "C"

import (
	"runtime"
	"unsafe"
)

func GetUndefined(env Env) (Value, Status) {
	var result Value
	status := Status(C.napi_get_undefined(
		C.napi_env(env),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetNull(env Env) (Value, Status) {
	var result Value
	status := Status(C.napi_get_null(
		C.napi_env(env),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetGlobal(env Env) (Value, Status) {
	var result Value
	status := Status(C.napi_get_global(
		C.napi_env(env),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetBoolean(env Env, value bool) (Value, Status) {
	var result Value
	status := Status(C.napi_get_boolean(
		C.napi_env(env),
		C.bool(value),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateObject(env Env) (Value, Status) {
	var result Value
	status := Status(C.napi_create_object(
		C.napi_env(env),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateArray(env Env) (Value, Status) {
	var result Value
	status := Status(C.napi_create_array(
		C.napi_env(env),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateArrayWithLength(env Env, length int) (Value, Status) {
	var result Value
	status := Status(C.napi_create_array_with_length(
		C.napi_env(env),
		C.size_t(length),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateDouble(env Env, value float64) (Value, Status) {
	var result Value
	status := Status(C.napi_create_double(
		C.napi_env(env),
		C.double(value),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateStringUtf8(env Env, str string) (Value, Status) {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))

	var result Value
	status := Status(C.napi_create_string_utf8(
		C.napi_env(env),
		cstr,
		C.size_t(len([]byte(str))), // must pass number of bytes
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateSymbol(env Env, description Value) (Value, Status) {
	var result Value
	status := Status(C.napi_create_symbol(
		C.napi_env(env),
		C.napi_value(description),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateFunction(env Env, name string, cb Callback) (Value, Status) {
	provider, status := getInstanceData(env)
	if status != StatusOK || provider == nil {
		return nil, status
	}

	return provider.GetCallbackData().CreateCallback(env, name, cb)
}

func CreateError(env Env, code, msg Value) (Value, Status) {
	var result Value
	status := Status(C.napi_create_error(
		C.napi_env(env),
		C.napi_value(code),
		C.napi_value(msg),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func Typeof(env Env, value Value) (ValueType, Status) {
	var result ValueType
	status := Status(C.napi_typeof(
		C.napi_env(env),
		C.napi_value(value),
		(*C.napi_valuetype)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetValueDouble(env Env, value Value) (float64, Status) {
	var result float64
	status := Status(C.napi_get_value_double(
		C.napi_env(env),
		C.napi_value(value),
		(*C.double)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetValueInt64(env Env, value Value) (int64, Status) {
	var result int64
	status := Status(C.napi_get_value_int64(
		C.napi_env(env),
		C.napi_value(value),
		(*C.int64_t)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetValueInt32(env Env, value Value) (int32, Status) {
	var result int32
	status := Status(C.napi_get_value_int32(
		C.napi_env(env),
		C.napi_value(value),
		(*C.int32_t)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetValueBigintInt64(env Env, value Value) (int64, bool, Status) {
	var result int64
	var lossless bool
	status := Status(C.napi_get_value_bigint_int64(
		C.napi_env(env),
		C.napi_value(value),
		(*C.int64_t)(unsafe.Pointer(&result)),
		(*C.bool)(unsafe.Pointer(&lossless)),
	))
	return result, lossless, status
}

func GetValueBool(env Env, value Value) (bool, Status) {
	var result bool
	status := Status(C.napi_get_value_bool(
		C.napi_env(env),
		C.napi_value(value),
		(*C.bool)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetValueStringUtf8(env Env, value Value) (string, Status) {
	// call napi_get_value_string_utf8 twice
	// first is to get number of bytes
	// second is to populate the actual string buffer
	bufsize := C.size_t(0)
	var strsize C.size_t

	status := Status(C.napi_get_value_string_utf8(
		C.napi_env(env),
		C.napi_value(value),
		nil,
		bufsize,
		&strsize,
	))

	if status != StatusOK {
		return "", status
	}

	// ensure there is room for the null terminator as well
	strsize++
	cstr := (*C.char)(C.malloc(C.sizeof_char * strsize))
	defer C.free(unsafe.Pointer(cstr))

	status = Status(C.napi_get_value_string_utf8(
		C.napi_env(env),
		C.napi_value(value),
		cstr,
		strsize,
		&strsize,
	))

	if status != StatusOK {
		return "", status
	}

	return C.GoStringN(
		(*C.char)(cstr),
		(C.int)(strsize),
	), status
}

func SetProperty(env Env, object, key, value Value) Status {
	return Status(C.napi_set_property(
		C.napi_env(env),
		C.napi_value(object),
		C.napi_value(key),
		C.napi_value(value),
	))
}

func SetElement(env Env, object Value, index int, value Value) Status {
	return Status(C.napi_set_element(
		C.napi_env(env),
		C.napi_value(object),
		C.uint32_t(index),
		C.napi_value(value),
	))
}

func StrictEquals(env Env, lhs, rhs Value) (bool, Status) {
	var result bool
	status := Status(C.napi_strict_equals(
		C.napi_env(env),
		C.napi_value(lhs),
		C.napi_value(rhs),
		(*C.bool)(&result),
	))
	return result, status
}

type GetCbInfoResult struct {
	Args []Value
	This Value
}

func GetCbInfo(env Env, info CallbackInfo) (GetCbInfoResult, Status) {
	// call napi_get_cb_info twice
	// first is to get total number of arguments
	// second is to populate the actual arguments
	argc := C.size_t(0)
	status := Status(C.napi_get_cb_info(
		C.napi_env(env),
		C.napi_callback_info(info),
		&argc,
		nil,
		nil,
		nil,
	))

	if status != StatusOK {
		return GetCbInfoResult{}, status
	}

	argv := make([]Value, int(argc))
	var cArgv unsafe.Pointer
	if argc > 0 {
		cArgv = unsafe.Pointer(&argv[0]) // must pass element pointer
	}

	var thisArg Value

	status = Status(C.napi_get_cb_info(
		C.napi_env(env),
		C.napi_callback_info(info),
		&argc,
		(*C.napi_value)(cArgv),
		(*C.napi_value)(unsafe.Pointer(&thisArg)),
		nil,
	))

	return GetCbInfoResult{
		Args: argv,
		This: thisArg,
	}, status
}

func Throw(env Env, err Value) Status {
	return Status(C.napi_throw(
		C.napi_env(env),
		C.napi_value(err),
	))
}

func ThrowError(env Env, code, msg string) Status {
	codeCStr, msgCCstr := C.CString(code), C.CString(msg)
	defer C.free(unsafe.Pointer(codeCStr))
	defer C.free(unsafe.Pointer(msgCCstr))

	return Status(C.napi_throw_error(
		C.napi_env(env),
		codeCStr,
		msgCCstr,
	))
}

func CreatePromise(env Env) (Promise, Status) {
	var result Promise
	status := Status(C.napi_create_promise(
		C.napi_env(env),
		(*C.napi_deferred)(unsafe.Pointer(&result.Deferred)),
		(*C.napi_value)(unsafe.Pointer(&result.Value)),
	))
	return result, status
}

func ResolveDeferred(env Env, deferred Deferred, resolution Value) Status {
	return Status(C.napi_resolve_deferred(
		C.napi_env(env),
		C.napi_deferred(deferred),
		C.napi_value(resolution),
	))
}

func RejectDeferred(env Env, deferred Deferred, rejection Value) Status {
	return Status(C.napi_reject_deferred(
		C.napi_env(env),
		C.napi_deferred(deferred),
		C.napi_value(rejection),
	))
}

func SetInstanceData(env Env, data any) Status {
	provider, status := getInstanceData(env)
	if status != StatusOK || provider == nil {
		return status
	}

	provider.SetUserData(data)
	return status
}

func GetInstanceData(env Env) (any, Status) {
	provider, status := getInstanceData(env)
	if status != StatusOK || provider == nil {
		return nil, status
	}

	return provider.GetUserData(), status
}

func CoerceToString(env Env, value Value) (Value, Status) {
	var result Value
	status := Status(C.napi_coerce_to_string(
		C.napi_env(env),
		C.napi_value(value),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CoerceToObject(env Env, value Value) (Value, Status) {
	var result Value
	status := Status(C.napi_coerce_to_object(
		C.napi_env(env),
		C.napi_value(value),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CoerceToNumber(env Env, value Value) (Value, Status) {
	var result Value
	status := Status(C.napi_coerce_to_number(
		C.napi_env(env),
		C.napi_value(value),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CoerceToBool(env Env, value Value) (Value, Status) {
	var result Value
	status := Status(C.napi_coerce_to_bool(
		C.napi_env(env),
		C.napi_value(value),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func IsError(env Env, value Value) (bool, Status) {
	var result bool
	status := Status(C.napi_is_error(
		C.napi_env(env),
		C.napi_value(value),
		(*C.bool)(unsafe.Pointer(&result)),
	))
	return result, status
}

func IsPromise(env Env, value Value) (bool, Status) {
	var result bool
	status := Status(C.napi_is_promise(
		C.napi_env(env),
		C.napi_value(value),
		(*C.bool)(unsafe.Pointer(&result)),
	))
	return result, status
}

func IsBuffer(env Env, value Value) (bool, Status) {
	var result bool
	status := Status(C.napi_is_buffer(
		C.napi_env(env),
		C.napi_value(value),
		(*C.bool)(unsafe.Pointer(&result)),
	))
	return result, status
}

func GetBufferInfo(env Env, value Value) ([]byte, Status) {
	var data unsafe.Pointer
	var length C.size_t
	status := Status(C.napi_get_buffer_info(
		C.napi_env(env),
		C.napi_value(value),
		&data,
		&length,
	))
	return unsafe.Slice((*byte)(data), length), status
}

func HasProperty(env Env, object Value, key Value) (bool, Status) {
	var result bool
	status := Status(C.napi_has_property(
		C.napi_env(env),
		C.napi_value(object),
		C.napi_value(key),
		(*C.bool)(unsafe.Pointer(&result)),
	))
	if status != StatusOK {
		return false, status
	}

	return result, status
}

func HasOwnProperty(env Env, object Value, key Value) (bool, Status) {
	var result bool
	status := Status(C.napi_has_own_property(
		C.napi_env(env),
		C.napi_value(object),
		C.napi_value(key),
		(*C.bool)(unsafe.Pointer(&result)),
	))
	if status != StatusOK {
		return false, status
	}

	return result, status
}

func GetProperty(env Env, object Value, key Value) (Value, Status) {
	var result Value
	status := Status(C.napi_get_property(
		C.napi_env(env),
		C.napi_value(object),
		C.napi_value(key),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	if status != StatusOK {
		return nil, status
	}

	return result, status
}

func CallFunction(env Env, recv Value, fn Value, args []Value) (Value, Status) {
	defer runtime.KeepAlive(args)

	var argsPtr *C.napi_value
	if len(args) > 0 {
		argsPtr = (*C.napi_value)(unsafe.Pointer(&args[0]))
	}

	var result Value
	status := Status(C.napi_call_function(
		C.napi_env(env),
		C.napi_value(recv),
		C.napi_value(fn),
		C.size_t(len(args)),
		argsPtr,
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	if status != StatusOK {
		return nil, status
	}

	return result, status
}
