package napi

/*
#include <stdlib.h>
#include <node/node_api.h>

extern void napiCallWrappedFinalizeFn(
	napi_env env,
	void *finalize_data,
	void *finalize_hint
);
*/
import "C"
import (
	"runtime/cgo"
	"unsafe"
)

type FinalizeFn func(env Env, finalizeData any, finalizeHint any)

type napiFinalizeFnData struct {
	fn   FinalizeFn
	data any
}

//export napiCallWrappedFinalizeFn
func napiCallWrappedFinalizeFn(
	env C.napi_env,
	finalizeData, finalizeHint unsafe.Pointer,
) {
	dataHandle := cgo.Handle(finalizeData)
	hintHandle := cgo.Handle(finalizeHint)
	data := dataHandle.Value().(napiFinalizeFnData)
	data.fn(
		Env(env),
		data.data,
		hintHandle.Value(),
	)

	dataHandle.Delete()
	hintHandle.Delete()
}

var _cCallWrappedFinalizeFn = C.napiCallWrappedFinalizeFn

func wrapFinalizeFnData(fn FinalizeFn, data any) unsafe.Pointer {
	return unsafe.Pointer(cgo.NewHandle(napiFinalizeFnData{
		fn:   fn,
		data: data,
	}))
}
