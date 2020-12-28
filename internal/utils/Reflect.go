package utils

import (
	"errors"
	"reflect"

	e "github.com/sam1677/ytdl/internal/ytdlerrors"
)

type reflectTypeHelper struct {
	typ  reflect.Type
	kind reflect.Kind
}

func newReflectTypeHelperFromType(t reflect.Type) *reflectTypeHelper {
	return &reflectTypeHelper{
		typ:  t,
		kind: t.Kind(),
	}
}

// if rh's kind is not in kinds, returns error
func (rh *reflectTypeHelper) assertKinds(kinds []reflect.Kind) error {
	for _, k := range kinds {
		if rh.kind == k {
			return nil
		}
	}
	return e.DbgErr(errors.New("target's kind is not matched with given kinds"))
}

type reflectHelper struct {
	val    reflect.Value
	valPtr reflect.Value
	*reflectTypeHelper
	arraySliceItemType reflect.Type
}

func newReflectHelper(target interface{}) *reflectHelper {
	val := reflect.ValueOf(target)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	rh := &reflectHelper{
		val:               val,
		reflectTypeHelper: newReflectTypeHelperFromType(val.Type()),
	}

	if val.CanAddr() {
		rh.valPtr = val.Addr()
	}

	err := rh.assertKinds([]reflect.Kind{
		reflect.Array,
		reflect.Slice,
	})
	if err == nil { // if rh's kind is array or slice
		rh.arraySliceItemType = rh.val.Index(0).Type()
	}

	return rh
}

func MapValues(mapp reflect.Value) []interface{} {
	iter := mapp.MapRange()
	values := []interface{}{}
	for iter.Next() {
		values = append(values, iter.Value().Interface())
	}
	return values
}
