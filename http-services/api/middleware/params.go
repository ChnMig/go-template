package middleware

import (
	"http-services/api/response"
	serviceLog "http-services/utils/log"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap"
)

// CheckParam binds parameters with Gin's default binder and writes an error response on failure.
func CheckParam(params interface{}, context *gin.Context) bool {
	return checkParamWithBinderAndMessage(
		params, context, binding.Default(context.Request.Method, context.ContentType()), "",
	)
}

// CheckParamWithMessage binds with the default binder and uses message on failure.
func CheckParamWithMessage(params interface{}, context *gin.Context, message string) bool {
	return checkParamWithBinderAndMessage(
		params, context, binding.Default(context.Request.Method, context.ContentType()), message,
	)
}

// CheckJSONParam binds a JSON request and writes an error response on failure.
func CheckJSONParam(params interface{}, context *gin.Context) bool {
	return checkParamWithBinderAndMessage(params, context, binding.JSON, "")
}

// CheckJSONParamWithMessage binds JSON and uses message on failure.
func CheckJSONParamWithMessage(params interface{}, context *gin.Context, message string) bool {
	return checkParamWithBinderAndMessage(params, context, binding.JSON, message)
}

// CheckQueryParam binds query parameters and writes an error response on failure.
func CheckQueryParam(params interface{}, context *gin.Context) bool {
	return checkParamWithBinderAndMessage(params, context, binding.Query, "")
}

// CheckQueryParamWithMessage binds query parameters and uses message on failure.
func CheckQueryParamWithMessage(params interface{}, context *gin.Context, message string) bool {
	return checkParamWithBinderAndMessage(params, context, binding.Query, message)
}

// BindParam binds parameters with Gin's default binder.
func BindParam(params interface{}, context *gin.Context) error {
	return bindParamWithBinder(params, context, binding.Default(context.Request.Method, context.ContentType()))
}

// BindJSONParam binds JSON parameters.
func BindJSONParam(params interface{}, context *gin.Context) error {
	return bindParamWithBinder(params, context, binding.JSON)
}

// BindQueryParam binds query parameters.
func BindQueryParam(params interface{}, context *gin.Context) error {
	return bindParamWithBinder(params, context, binding.Query)
}

func checkParamWithBinderAndMessage(
	params interface{}, context *gin.Context, binder binding.Binding, message string,
) bool {
	if err := bindParamWithBinder(params, context, binder); err != nil {
		zap.L().Error("operation failed", zap.Error(err))
		if message == "" {
			message = err.Error()
		}
		response.ReturnError(context, response.INVALID_ARGUMENT, message)
		return false
	}
	return true
}

func bindParamWithBinder(params interface{}, context *gin.Context, binder binding.Binding) error {
	if err := context.ShouldBindWith(params, binder); err != nil {
		return err
	}
	context.Set(serviceLog.BoundParamsKey, params)
	return nil
}
