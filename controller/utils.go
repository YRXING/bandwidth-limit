package controller

import (
	"errors"
	jsoniter "github.com/json-iterator/go"
	"strconv"
	"strings"
	"unicode"
)

//translate ingress/egress to tc rate
func Translate(tf string) (uint64, error) {
	f := func(r rune) bool {
		return unicode.IsLetter(r)
	}
	index := strings.IndexFunc(tf, f)
	digit, err := strconv.Atoi(tf[0:index])
	if err != nil {
		return 0, err
	}
	unit := tf[index:]
	unit = strings.ToLower(unit)
	var rate uint64
	switch unit {
	case TC_BPS:
		break
	case TC_KPS:
		rate = Pow(digit, 3)
	case TC_MPS:
		rate = Pow(digit, 6)
	case TC_GPS:
		rate = Pow(digit, 9)
	case TC_TPS:
		rate = Pow(digit, 12)
	default:
		err = errors.New("invalid unit")
	}
	if err != nil {
		return 0, err
	}
	return rate, nil
}

func Pow(data, n int) uint64 {
	for i := 0; i < n; i++ {
		data *= 10
	}
	return uint64(data)
}


//When a json object is large and has many fields. It is absolutely unnecessary to
//create a large entity class object in order to access several fields.
func GetFiledFromJason(path []string, data []byte) string {
	var temp jsoniter.Any
	for i,v := range path {
		if i==0 {
			temp = jsoniter.Get(data,v)
			if temp == nil {
				return ""
			}
		}else {
			temp = temp.Get(v)
			if temp == nil {
				return ""
			}
		}
	}

	switch temp.ValueType() {
	case jsoniter.InvalidValue,jsoniter.NilValue,jsoniter.BoolValue,jsoniter.ArrayValue,jsoniter.ObjectValue:
		return ""
	case jsoniter.StringValue:
		return temp.ToString()
	case jsoniter.NumberValue:
		return strconv.Itoa(temp.ToInt())
	}

	return ""
}


