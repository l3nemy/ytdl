package ytdlerrors

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
)

type moduleErr error

//YtdlErr Describes YoutubeDownloader Error
type YtdlErr moduleErr

//CipherErr Describes Cipher Error
type CipherErr moduleErr

//RegexpErr Describes Regexp Error
type RegexpErr moduleErr

var (
	//ErrURLIsEmpty URL is Empty
	ErrURLIsEmpty YtdlErr = errors.New("URL is empty. (Check if Format's URL is Ciphered)")
	//ErrJSstrIsEmpty JSstr is empty
	ErrJSstrIsEmpty CipherErr = errors.New("Js string is empty")
	//ErrRegexpNotMatched RegexpNotMatched
	ErrRegexpNotMatched RegexpErr = errors.New("Regexp Search Error: Not Matched")
	//ErrFuncListIsTooShort len(functionList) < 3
	ErrFuncListIsTooShort RegexpErr = errors.New("Regexp Match Error: function list length is less than 3")
	//ErrNoMatchOnFunction function regexp not matched
	ErrNoMatchOnFunction RegexpErr = errors.New("Regexp Match Error: No Match on function")
)

//DbgMode activates debug mode
var DbgMode = false

//DbgErr adds caller string to error
func DbgErr(err error) error {
	_, file, line, ok := runtime.Caller(1)

	if ok && DbgMode {
		return fmt.Errorf("%v\n\t(From %s:%d)", err, file, line)
	}
	return err
}

//IsYtdlErr Checks if err is YtdlErr
func IsYtdlErr(err error) bool {
	return IsImplements(err, reflect.TypeOf(new(YtdlErr)).Elem())
}

//IsCipherErr Checks if err is CipherErr
func IsCipherErr(err error) bool {
	return IsImplements(err, reflect.TypeOf(new(CipherErr)).Elem())
}

//IsRegexpErr Checks if err is RegexpErr
func IsRegexpErr(err error) bool {
	return IsImplements(err, reflect.TypeOf(new(RegexpErr)).Elem())
}

//IsImplements Checks if inter implements typ
func IsImplements(inter interface{}, typ reflect.Type) bool {
	return reflect.TypeOf(inter).Implements(typ)
}
