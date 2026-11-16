package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// CheckStrictJSONParam 拒绝未知字段及尾随 JSON，不修改 Gin 全局解码设置。
func CheckStrictJSONParam(params any, c *gin.Context) bool {
	return checkParamWithBinderAndMessage(params, c, strictJSONBinding{}, "")
}

// CheckStrictJSONParamWithMessage 在严格绑定失败时返回指定提示，日志保留原始错误。
func CheckStrictJSONParamWithMessage(params any, c *gin.Context, message string) bool {
	return checkParamWithBinderAndMessage(params, c, strictJSONBinding{}, message)
}

// BindStrictJSONParam 只返回绑定错误，由调用方决定响应和日志。
func BindStrictJSONParam(params any, c *gin.Context) error {
	return bindParamWithBinder(params, c, strictJSONBinding{})
}

type strictJSONBinding struct{}

func (strictJSONBinding) Name() string { return "json" }

func (strictJSONBinding) Bind(request *http.Request, target any) error {
	if request == nil || request.Body == nil {
		return errors.New("JSON 请求体不能为空")
	}
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		if err != nil {
			return fmt.Errorf("JSON 请求体末尾内容无法解析: %w", err)
		}
		return errors.New("请求体只能包含一个 JSON 值")
	}
	return binding.Validator.ValidateStruct(target)
}
