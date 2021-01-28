// Package pocket Create at 2020-11-06 10:20
package pocket

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

// md5 md5 32 lowercase
func Md5(txt string) string {
	md5Hash := md5.New()
	io.WriteString(md5Hash, txt)
	md5Bytes := md5Hash.Sum(nil)
	return strings.ToLower(hex.EncodeToString(md5Bytes))
}

// SnakeString, XxYy to xx_yy , XxYY to xx_yy
func SnakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(string(data[:]))
}

// DecodeQuery decode get query
func DecodeQuery(dst interface{}, src map[string][]string) error {
	t := reflect.TypeOf(dst)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		return errors.New("schema: interface must be a pointer to struct")
	}
	v := reflect.ValueOf(dst).Elem()
	for i := 0; i < t.Elem().NumField(); i++ {
		tag := t.Elem().Field(i).Tag.Get("schema")
		if "" != tag {
			if param, ok := src[tag]; ok {
				if len(param) > 0 && len(param[0]) > 0 {
					switch v.Field(i).Kind() {
					case reflect.Ptr:
						switch v.Field(i).Type().String() {
						case "*int":
							val, err := strconv.Atoi(param[0])
							if nil != err {
								DefaultLogger.Error(err.Error())
								break
							}
							p := new(int)
							*p = val
							v.Field(i).Set(reflect.ValueOf(p))
						case "*int64":
							val, err := strconv.ParseInt(param[0], 10, 64)
							if nil != err {
								DefaultLogger.Error(err.Error())
								break
							}
							p := new(int64)
							*p = val
							v.Field(i).Set(reflect.ValueOf(p))
						case "*string":
							p := new(string)
							*p = param[0]
							v.Field(i).Set(reflect.ValueOf(p))
						case "*float32":
							val, err := strconv.ParseFloat(param[0], 32)
							if nil != err {
								DefaultLogger.Error(err.Error())
								break
							}
							p := new(float32)
							*p = float32(val)
							v.Field(i).Set(reflect.ValueOf(p))
						case "*float64":
							val, err := strconv.ParseFloat(param[0], 64)
							if nil != err {
								DefaultLogger.Error(err.Error())
								break
							}
							p := new(float64)
							*p = val
							v.Field(i).Set(reflect.ValueOf(p))
						case "*bool":
							b := false
							if "true" == param[0] {
								b = true
							}
							p := new(bool)
							*p = b
							v.Field(i).Set(reflect.ValueOf(p))
						default:
							return errors.New(fmt.Sprintf("Not Support %s", tag))
						}
					case reflect.String:
						v.Field(i).SetString(param[0])
					case reflect.Int:
						val, err := strconv.ParseInt(param[0], 10, 32)
						if nil != err {
							DefaultLogger.Error(err.Error())
							break
						}
						v.Field(i).SetInt(val)
					case reflect.Int64:
						val, err := strconv.ParseInt(param[0], 10, 64)
						if nil != err {
							DefaultLogger.Error(err.Error())
							break
						}
						v.Field(i).SetInt(val)
					case reflect.Float32, reflect.Float64:
						val, err := strconv.ParseFloat(param[0], 64)
						if nil != err {
							DefaultLogger.Error(err.Error())
							break
						}
						v.Field(i).SetFloat(val)
					case reflect.Bool:
						b := false
						if "true" == param[0] {
							b = true
						}
						v.Field(i).SetBool(b)
					case reflect.Slice, reflect.Array:
						str := param[0]
						arr := strings.Split(str, ",")
						sl := reflect.MakeSlice(v.Field(i).Elem().Type(), len(arr), len(arr))
						switch v.Field(i).Elem().Kind() {
						case reflect.Int:
							for _, k := range arr {
								val, err := strconv.ParseInt(k, 10, 32)
								if nil != err {
									DefaultLogger.Error(err.Error())
									break
								}
								item := int(val)
								sl = reflect.AppendSlice(sl, reflect.ValueOf(item))
							}
						case reflect.Int64:
							for _, k := range arr {
								val, err := strconv.ParseInt(k, 10, 64)
								if nil != err {
									DefaultLogger.Error(err.Error())
									break
								}
								sl = reflect.AppendSlice(sl, reflect.ValueOf(val))
							}
						case reflect.String:
							for _, k := range arr {
								sl = reflect.AppendSlice(sl, reflect.ValueOf(k))
							}
						}
						v.Field(i).Set(sl)
					default:
						return errors.New(fmt.Sprintf("Not Support %s", tag))
					}
				}
			}
		}
	}
	return nil
}

// Utf8Index Index returns the index of the first instance of substr in s, or -1 if substr is not present in s
func Utf8Index(str, substr string) int {
	index := strings.Index(str, substr)
	if index < 0 {
		return -1
	}
	return utf8.RuneCountInString(str[:index])
}

// HexStr2Uint32 hex string to uint32
func HexStr2Uint32(s string) (uint32, error) {
	b, err := hex.DecodeString(s)
	if nil != err {
		DefaultLogger.Error(err)
		return 0, err
	}
	return binary.BigEndian.Uint32(b), nil
}

// HexStr2Uint16 hex string to uint16
func HexStr2Uint16(s string) (uint16, error) {
	b, err := hex.DecodeString(s)
	if nil != err {
		DefaultLogger.Error(err)
		return 0, err
	}
	return binary.BigEndian.Uint16(b), nil
}

// HexStr2Uint64 hex string to uint64
func HexStr2Uint64(s string) (uint64, error) {
	b, err := hex.DecodeString(s)
	if nil != err {
		DefaultLogger.Error(err)
		return 0, err
	}
	return binary.BigEndian.Uint64(b), nil
}
