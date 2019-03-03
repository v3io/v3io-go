/*
Copyright 2018 The v3io Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v3io

import (
	"strconv"

	"github.com/v3io/v3io-go/pkg/errors"
)

type Item map[string]interface{}

func (i Item) GetField(name string) interface{} {
	return i[name]
}

func (i Item) GetFieldInt(name string) (int, error) {
	switch typedField := i[name].(type) {
	case int:
		return typedField, nil
	case float64:
		return int(typedField), nil
	case string:
		return strconv.Atoi(typedField)
	default:
		return 0, v3ioerrors.ErrInvalidTypeConversion
	}
}

func (i Item) GetFieldString(name string) (string, error) {
	switch typedField := i[name].(type) {
	case int:
		return strconv.Itoa(typedField), nil
	case float64:
		return strconv.FormatFloat(typedField, 'E', -1, 64), nil
	case string:
		return typedField, nil
	default:
		return "", v3ioerrors.ErrInvalidTypeConversion
	}
}
