package napi

/*
#include <stdlib.h>
#include <node/node_api.h>
*/
import "C"

import "unsafe"

type Ref unsafe.Pointer

func GetReferenceValue(env Env, ref Ref) (Value, Status) {
	var result Value
	status := Status(C.napi_get_reference_value(
		C.napi_env(env),
		C.napi_ref(ref),
		(*C.napi_value)(unsafe.Pointer(&result)),
	))
	return result, status
}

func CreateReference(env Env, value Value, initialRefcount int) (Ref, Status) {
	var result Ref
	status := Status(C.napi_create_reference(
		C.napi_env(env),
		C.napi_value(value),
		C.uint32_t(initialRefcount),
		(*C.napi_ref)(unsafe.Pointer(&result)),
	))
	return result, status
}

func DeleteReference(env Env, ref Ref) Status {
	return Status(C.napi_delete_reference(
		C.napi_env(env),
		C.napi_ref(ref),
	))
}

func ReferenceRef(env Env, ref Ref) (int, Status) {
	var result C.uint32_t
	status := Status(C.napi_reference_ref(
		C.napi_env(env),
		C.napi_ref(ref),
		&result,
	))
	return int(result), status
}

func ReferenceUnref(env Env, ref Ref) (int, Status) {
	var result C.uint32_t
	status := Status(C.napi_reference_unref(
		C.napi_env(env),
		C.napi_ref(ref),
		&result,
	))
	return int(result), status
}
