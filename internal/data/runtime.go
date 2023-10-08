package data

import (
	"fmt"
	"strconv"
)

type Runtime int32

func (r Runtime) MarshalJSON() ([]byte, error) {
	var jsonValue = fmt.Sprintf("%d mins", r)
	var quotedJSONValue = strconv.Quote(jsonValue)
	return []byte(quotedJSONValue), nil
}
