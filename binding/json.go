package binding

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
)

type jsonBinding struct {
	DisallowUnknownFields bool
	IsInvalid             bool
}

func (j *jsonBinding) Name() string {
	return "json"
}

func (j *jsonBinding) Bind(r *http.Request, obj interface{}) error {
	//post 传参保存到body
	body := r.Body
	if body == nil {
		return errors.New("invalid request")
	}
	decoder := json.NewDecoder(body)
	if j.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}

	var err error
	if j.IsInvalid {
		err = validateParam(obj, decoder)
	} else {
		err = decoder.Decode(obj)
	}
	if err != nil {
		return err
	}
	//第三方 validator
	return validate(obj)
}

func validateParam(obj interface{}, decoder *json.Decoder) error {
	//解析为map ,根据map key 进行对比
	val := reflect.ValueOf(obj)
	if val.Kind() != reflect.Ptr {
		return errors.New("not ptr type")
	}
	elem := val.Elem().Interface()
	of := reflect.ValueOf(elem)
	switch of.Kind() {
	case reflect.Struct:
		return checkParam(of, obj, decoder)
	case reflect.Slice, reflect.Array:
		t := of.Type().Elem()
		if t.Kind() == reflect.Struct {
			return checkParamSlice(of, obj, decoder)
		}
	default:
		decoder.Decode(obj)
	}
	return nil
}

func checkParamSlice(of reflect.Value, obj interface{}, decoder *json.Decoder) error {
	//结构体才能解析为map
	mapVal := make([]map[string]interface{}, 0)
	err := decoder.Decode(&mapVal)
	if err != nil {
		return err
	}
	for i := 0; i < of.NumField(); i++ {
		field := of.Type().Field(i)
		name := field.Name
		tag := field.Tag.Get("json")
		if tag != "" {
			name = tag
		}
		for _, v := range mapVal {
			value := v[name]
			required := field.Tag.Get("cob")
			if value == nil && required == "required" {
				return errors.New(fmt.Sprintf("field [%s] is not exist", tag))
			}
		}
	}
	marshal, err := json.Marshal(mapVal)
	if err != nil {
		return err
	}
	return json.Unmarshal(marshal, obj)
}

func checkParam(of reflect.Value, obj interface{}, decoder *json.Decoder) error {
	//结构体才能解析为map
	mapVal := make(map[string]interface{})
	err := decoder.Decode(&mapVal)
	if err != nil {
		return err
	}
	for i := 0; i < of.NumField(); i++ {
		field := of.Type().Field(i)
		name := field.Name
		tag := field.Tag.Get("json")
		if tag != "" {
			name = tag
		}
		value := mapVal[name]
		required := field.Tag.Get("cob")
		if value == nil && required == "required" {
			return errors.New(fmt.Sprintf("field [%s] is not exist", tag))
		}
	}
	marshal, err := json.Marshal(mapVal)
	if err != nil {
		return err
	}
	return json.Unmarshal(marshal, obj)
}
