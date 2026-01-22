package promise

import (
	"errors"

	"github.com/akshayganeshen/napi-go"
	"github.com/akshayganeshen/napi-go/js"
)

var (
	ErrAlreadySettled = errors.New("promise is already settled")
)

type Promise struct {
	js.Value

	tsfn    js.ThreadsafeFunction
	settled bool
}

type Deferred struct {
	Env      js.Env
	Deferred napi.Deferred
}

func (d Deferred) Resolve(value any) error {
	v, err := d.Env.ValueOf(value)
	if err != nil {
		return err
	}

	st := napi.ResolveDeferred(d.Env.Env, d.Deferred, v.Value)
	if st != napi.StatusOK {
		return napi.StatusError(st)
	}

	return nil
}

func (d Deferred) Reject(value any) error {
	v, err := d.Env.ValueOf(value)
	if err != nil {
		return err
	}

	st := napi.RejectDeferred(d.Env.Env, d.Deferred, v.Value)
	if st != napi.StatusOK {
		return napi.StatusError(st)
	}

	return nil
}

type Settler interface {
	Settle(env js.Env, deferred Deferred, data any) error
}

type SettlerFunc func(env js.Env, deferred Deferred, data any) error

func (f SettlerFunc) Settle(env js.Env, deferred Deferred, data any) error {
	return f(env, deferred, data)
}

type promiseTsfnContext struct {
	deferred napi.Deferred
	settler  Settler
}

func (c promiseTsfnContext) CallJs(env js.Env, fn js.Value, data any) error {
	return c.settler.Settle(env, Deferred{Env: env, Deferred: c.deferred}, data)
}

func NewPromise(env js.Env, settler Settler) (Promise, error) {
	p, st := napi.CreatePromise(env.Env)
	if st != napi.StatusOK {
		return Promise{}, napi.StatusError(st)
	}

	tsfn, err := env.NewThreadsafeFunction(
		nil,
		"napi-go/promise",
		promiseTsfnContext{
			deferred: p.Deferred,
			settler:  settler,
		},
		nil,
	)
	if err != nil {
		return Promise{}, err
	}

	return Promise{
		Value: env.WrapValue(p.Value),
		tsfn:  tsfn,
	}, nil
}

func (p Promise) Settle(data any) error {
	if p.settled {
		return ErrAlreadySettled
	}

	p.settled = true

	defer p.tsfn.Release()

	return p.tsfn.Call(data)
}
