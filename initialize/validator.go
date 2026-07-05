package initialize

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	customValidator "olixops/pkg/validator"
)

func InitValidator() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {

		// 注册所有自定义验证规则
		v.RegisterValidation("username", customValidator.ValidateUsername)
		v.RegisterValidation("password_strong", customValidator.ValidatePasswordStrong)
		v.RegisterValidation("phone", customValidator.ValidatePhone)

		// 注册结构体级别的校验
		// v.RegisterStructValidation(ValidateUser, User{})
	}
}
