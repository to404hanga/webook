package domain

import (
	"errors"
	"fmt"

	"github.com/ecodeclub/ekit"
)

type ExtendFields map[string]string

var errKeyNotFound = errors.New("没有找到对应的 key")

// TODO 使用自己的实现替换 ekit.AnyValue
func (e ExtendFields) Get(key string) ekit.AnyValue {
	val, ok := e[key]
	if !ok {
		return ekit.AnyValue{
			Err: fmt.Errorf("%w, key %s", errKeyNotFound, key),
		}
	}
	return ekit.AnyValue{Val: val}
}
