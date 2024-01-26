package properties

import (
	"errors"
)

var (
	ErrInvalidUnmarshal = errors.New("v must be a non-nil struct pointer")
	ErrInvalidMarshal   = errors.New("v must be of type map, map pointer, struct or struct pointer")
	ErrInvalidPropBytes = errors.New("bytes are not from valid .properties config")
	ErrUnsupportedType  = errors.New("unsupported type")
)

func Marshal(v interface{}) ([]byte, error) {
	return marshal(v)
}

func Unmarshal(data []byte, v interface{}) error {
	p, err := propsFromBytes(data, "")
	if err != nil {
		return err
	}
	return UnmarshalKV(p.kv, v)
}

func UnmarshalKV(kv map[string]string, v interface{}) error {
	return unmarshalKV(kv, v)
}

func UnmarshalKey(key string, data []byte, v interface{}) error {
	p, err := propsFromBytes(data, key+".")
	if err != nil {
		return err
	}
	return UnmarshalKV(p.kv, v)
}
