package js

import "github.com/akshayganeshen/napi-go"

type Buffer struct {
	Value
}

func (v Value) IsBuffer() (bool, error) {
	b, st := napi.IsBuffer(v.Env.Env, v.Value)
	if err := st.AsError(); err != nil {
		return false, err
	}

	return b, nil
}

func (v Value) AsBufferUnsafe() Buffer {
	return Buffer{
		Value: v,
	}
}

func (v Value) AsBuffer() (Buffer, error) {
	if ok, err := v.IsBuffer(); err != nil {
		return Buffer{}, err
	} else if !ok {
		return Buffer{}, ErrWrongType
	}

	return v.AsBufferUnsafe(), nil
}

func (v Buffer) GetBytes() ([]byte, error) {
	data, st := napi.GetBufferInfo(v.Env.Env, v.Value.Value)
	if st != napi.StatusOK {
		return nil, napi.StatusError(st)
	}

	return data, nil
}
