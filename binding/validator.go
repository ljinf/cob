package binding

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
	"sync"
)

type StructValidator interface {
	//结构体验证
	ValidateStruct(interface{}) error
	//返回使用的验证器
	Engine() interface{}
}

var Validator StructValidator = &defaultValidator{}

type defaultValidator struct {
	one      sync.Once
	validate *validator.Validate
}

func (v *defaultValidator) ValidateStruct(obj interface{}) error {
	of := reflect.ValueOf(obj)
	switch of.Kind() {
	case reflect.Ptr:
		return v.ValidateStruct(of.Elem().Interface())
	case reflect.Struct:
		return v.validateStruct(obj)
	case reflect.Slice, reflect.Array:
		errs := make([]string, of.Len())
		for i := 0; i < of.Len(); i++ {
			errs = append(errs, v.validateStruct(of.Index(i).Interface()).Error())
		}
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}

func (v *defaultValidator) Engine() interface{} {
	v.lazyInit()
	return v.validate
}

func (v *defaultValidator) lazyInit() {
	v.one.Do(func() {
		v.validate = validator.New()
	})
}

func (v *defaultValidator) validateStruct(obj interface{}) error {
	v.lazyInit()
	return v.validate.Struct(obj)
}

func validate(obj interface{}) error {
	return Validator.ValidateStruct(obj)
}
