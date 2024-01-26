package properties

import (
	"encoding"
	"fmt"
	"reflect"
	"strings"
)

var (
	typTextMarshaler   = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	typTextUnmarshaler = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

func toPropLineBytes(key, val string) []byte {
	return []byte(fmt.Sprintf("%s=%s\n", key, escape(val)))
}

func marshal(v interface{}) ([]byte, error) {
	rv := reflect.ValueOf(v)

	if rv.Kind() == reflect.Map || rv.Kind() == reflect.Struct ||
		rv.Kind() == reflect.Ptr && (rv.Elem().Kind() == reflect.Struct || rv.Elem().Kind() == reflect.Map) {
		return devalue("", rv)
	}

	return nil, ErrInvalidMarshal
}

func devalue(key string, v reflect.Value) ([]byte, error) {
	var data []byte
	switch v.Kind() {
	case reflect.Ptr:
		return devalue(key, v.Elem())
	case reflect.Struct:
		if v.Type().Implements(typTextMarshaler) {
			text, err := v.Interface().(encoding.TextMarshaler).MarshalText()
			if err != nil {
				return nil, err
			}
			return toPropLineBytes(key, string(text)), nil
		}
		for i := 0; i < v.NumField(); i++ {
			vf, tf := v.Field(i), v.Type().Field(i)

			kk, _ := parseTag(tf.Tag.Get(tagName))

			if vf.String() == "" {
				continue
			}

			if kk == "-" {
				continue
			}
			if kk == "" {
				kk = tf.Name
			}

			if key != "" {
				kk = fmt.Sprintf("%s.%s", key, kk)
			}

			d, err := devalue(kk, vf)
			if err != nil {
				return nil, err
			}
			data = append(data, d...)
		}
	case reflect.Map:
		for _, kk := range v.MapKeys() {
			vv := v.MapIndex(kk)

			var nkey string
			if key != "" {
				nkey = fmt.Sprintf("%s.%s", key, kk.Interface())
			} else {
				nkey = fmt.Sprint(kk.Interface())
			}

			d, err := devalue(nkey, vv)
			if err != nil {
				return nil, err
			}
			data = append(data, d...)
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			vv := v.Index(i)
			d, err := devalue(fmt.Sprintf("%s[%d]", key, i), vv)
			if err != nil {
				return nil, err
			}
			data = append(data, d...)
		}
	case reflect.String:
		fallthrough
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		fallthrough
	case reflect.Bool:
		fallthrough
	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		return toPropLineBytes(key, fmt.Sprint(v.Interface())), nil
	}
	return data, nil
}

func escape(raw string) string {

	sb := strings.Builder{}

	for _, r := range raw {
		switch r {
		case '\f':
			sb.WriteString("\\f")
		case '\n':
			sb.WriteString("\\n")
		case '\r':
			sb.WriteString("\\r")
		case '\t':
			sb.WriteString("\\t")
		case '\\':
			sb.WriteString("\\\\")
		case ':':
			sb.WriteString("\\:")
		case '=':
			sb.WriteString("\\=")
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
