package napi

/*
#include <node/node_api.h>

extern void napiCallWrappedTsfnCallJsFn(
	napi_env env,
	napi_value js_callback,
	void *context,
	void *data
);
*/
import "C"

import (
	"runtime/cgo"
	"unsafe"
)

type TsfnCallJsFn func(env Env, jsCallback Value, context any, data any)

type tsfnContext struct {
	realFn      TsfnCallJsFn
	realContext any
}

//export napiCallWrappedTsfnCallJsFn
func napiCallWrappedTsfnCallJsFn(
	env C.napi_env,
	jsCallback C.napi_value,
	contextPtr, dataPtr unsafe.Pointer,
) {
	dataHandle := cgo.Handle(dataPtr)
	context := cgo.Handle(contextPtr).Value().(tsfnContext)
	context.realFn(
		Env(env),
		Value(jsCallback),
		context.realContext,
		dataHandle.Value(),
	)

	dataHandle.Delete()
}

type tsfnFinalizeData struct {
	realFn   FinalizeFn
	realData any
}

func tsfnFinalize(env Env, finalizeData any, finalizeHint any) {
	data := finalizeData.(tsfnFinalizeData)
	data.realFn(env, data.realData, finalizeHint.(tsfnContext).realContext)
}

func CreateThreadsafeFunction(
	env Env,
	fn Value,
	asyncResource, asyncResourceName Value,
	maxQueueSize, initialThreadCount int,
	threadFinalizeFnHint any, threadFinalizeFn FinalizeFn,
	context any, callJsFn TsfnCallJsFn,
) (ThreadsafeFunction, Status) {
	var result ThreadsafeFunction
	status := Status(C.napi_create_threadsafe_function(
		C.napi_env(env),
		C.napi_value(fn),
		C.napi_value(asyncResource),
		C.napi_value(asyncResourceName),
		C.size_t(maxQueueSize),
		C.size_t(initialThreadCount),
		wrapFinalizeFnData(tsfnFinalize, tsfnFinalizeData{
			realFn:   threadFinalizeFn,
			realData: threadFinalizeFnHint,
		}),
		(*[0]byte)(_cCallWrappedFinalizeFn),
		unsafe.Pointer(cgo.NewHandle(tsfnContext{
			realFn:      callJsFn,
			realContext: context,
		})),
		(*[0]byte)(C.napiCallWrappedTsfnCallJsFn),
		(*C.napi_threadsafe_function)(unsafe.Pointer(&result)),
	))
	return result, status
}

// CallThreadsafeFunction defaults to blocking call mode
func CallThreadsafeFunction(
	fn ThreadsafeFunction,
	data any,
	blocking ...bool,
) Status {
	var callMode C.napi_threadsafe_function_call_mode = C.napi_tsfn_blocking
	if len(blocking) > 0 && !blocking[0] {
		callMode = C.napi_tsfn_nonblocking
	}
	return Status(C.napi_call_threadsafe_function(
		C.napi_threadsafe_function(fn),
		unsafe.Pointer(cgo.NewHandle(data)),
		callMode,
	))
}

func AcquireThreadsafeFunction(fn ThreadsafeFunction) Status {
	return Status(C.napi_acquire_threadsafe_function(
		C.napi_threadsafe_function(fn),
	))
}

func ReleaseThreadsafeFunction(
	fn ThreadsafeFunction,
) Status {
	return Status(C.napi_release_threadsafe_function(
		C.napi_threadsafe_function(fn),
		C.napi_tsfn_release,
	))
}
