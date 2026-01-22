package js

import (
	"log"

	"github.com/akshayganeshen/napi-go"
)

type TsfnContext interface {
	CallJs(env Env, fn Value, data any) error
}

type TsfnContextFunc func(env Env, fn Value, data any) error

func (f TsfnContextFunc) CallJs(env Env, fn Value, data any) error {
	return f(env, fn, data)
}

func tsfnCallJsWrapper(e napi.Env, callback napi.Value, context any, data any) {
	env := WrapEnv(e)
	err := context.(TsfnContext).CallJs(
		env,
		env.WrapValue(callback),
		data,
	)
	if err != nil {
		if err2 := env.ThrowError("", err.Error()); err2 != nil {
			log.Println("WARN: threadsafe_function_call_js_wrapper: an error occurred while handling another")
			log.Println("WARN: threadsafe_function_call_js_wrapper: original", err)
			log.Println("WARN: threadsafe_function_call_js_wrapper: secondary", err2)
		}
	}
}

type TsfnFinalizer interface {
	Finalize(env Env, context TsfnContext) error
}

type TsfnFinalizerFunc func(env Env, context TsfnContext) error

func (f TsfnFinalizerFunc) Finalize(env Env, context TsfnContext) error {
	return f(env, context)
}

func tsfnFinalizerWrapper(e napi.Env, data any, hint any) {
	if data == nil {
		return
	}

	env := WrapEnv(e)
	err := data.(TsfnFinalizer).Finalize(env, hint.(TsfnContext))
	if err != nil {
		if err2 := env.ThrowError("", err.Error()); err2 != nil {
			log.Println("WARN: threadsafe_function_finalizer_wrapper: an error occurred while handling another")
			log.Println("WARN: threadsafe_function_finalizer_wrapper: original", err)
			log.Println("WARN: threadsafe_function_finalizer_wrapper: secondary", err2)
		}
	}
}

func defaultTsfnCallJs(env Env, fn Value, data any) error {
	obj, err := env.ValueOf(data)
	if err != nil {
		return err
	}

	if _, err := fn.AsFunctionUnsafe().Call(nil, obj); err != nil {
		return err
	}

	return nil
}

var DefaultTsfnContext = TsfnContextFunc(defaultTsfnCallJs)

func (e Env) NewThreadsafeFunction(v AnyValue, name string, context TsfnContext, finalizer TsfnFinalizer) (ThreadsafeFunction, error) {
	var value napi.Value
	if v != nil {
		value = v.GetValue().Value
	}

	nameValue, err := e.ValueOf(name)
	if err != nil {
		return ThreadsafeFunction{}, err
	}

	if context == nil {
		context = DefaultTsfnContext
	}

	tsfn, st := napi.CreateThreadsafeFunction(
		e.Env,
		value,
		nil,
		nameValue.Value,
		0,
		1,
		finalizer,
		tsfnFinalizerWrapper,
		context,
		tsfnCallJsWrapper,
	)
	if st != napi.StatusOK {
		return ThreadsafeFunction{}, napi.StatusError(st)
	}

	return ThreadsafeFunction{
		Tsfn: tsfn,
	}, nil
}

type ThreadsafeFunction struct {
	Tsfn napi.ThreadsafeFunction
}

func (f ThreadsafeFunction) Valid() bool {
	return f.Tsfn != nil
}

func finalizeTsfn(f ThreadsafeFunction) {
	st := napi.ReleaseThreadsafeFunction(f.Tsfn)
	if st == napi.StatusOK {
		return
	}

	log.Printf("Failed to release ThreadSafeFunction: %v", st)
}

func (f ThreadsafeFunction) Call(data any) error {
	st := napi.CallThreadsafeFunction(f.Tsfn, data)
	if st != napi.StatusOK {
		return napi.StatusError(st)
	}

	return nil
}

func (f ThreadsafeFunction) Acquire() error {
	st := napi.AcquireThreadsafeFunction(f.Tsfn)
	if st != napi.StatusOK {
		return napi.StatusError(st)
	}

	return nil
}

func (f ThreadsafeFunction) Release() error {
	st := napi.ReleaseThreadsafeFunction(f.Tsfn)
	if st != napi.StatusOK {
		return napi.StatusError(st)
	}

	return nil
}
