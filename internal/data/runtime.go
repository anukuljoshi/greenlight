package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Runtime int32

func (r Runtime) MarshalJSON() ([]byte, error) {
	var jsonValue = fmt.Sprintf("%d mins", r)
	var quotedJSONValue = strconv.Quote(jsonValue)
	return []byte(quotedJSONValue), nil
}

// error to return if Runtime value is not in correct format
var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	var unquotedJSONValue, err = strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// split the string
	var parts = strings.Split(unquotedJSONValue, " ")
	// check if  parts are in correct format
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}
	// parse first part to int
	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}
	*r = Runtime(i)
	return nil
}
