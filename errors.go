package csort

import "errors"

// 错误定义
var (
	ErrInvalidScore   = errors.New("invalid score format")
	ErrKeyNotFound    = errors.New("key not found")
	ErrMemberNotFound = errors.New("member not found")
)
