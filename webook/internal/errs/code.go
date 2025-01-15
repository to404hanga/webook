package errs

// User 相关
const (
	UserInvalidInput      = 401001 // 统一的用户模块的输入错误
	UserInvalidOrPassword = 401002 // 用户名或密码错误
	UserDuplicateEmail    = 401003 // 用户邮箱冲突
	UserSmsSendTooMany    = 401004 // 用户验证码发送太频繁

	UserInternalServerError = 501001 // 统一的用户模块的系统错误
)

// Article 相关
const (
	ArticleInvalidInput = 402001 // 统一的文章模块的输入错误

	ArticleInternalServerError = 502001 // 统一的文章模块的系统错误
)

// Wechat 相关
const (
	WechatInvalidInput    = 403001 // 统一的微信模块的输入错误
	WechatAuthorizeFailed = 403002 // 微信授权失败

	WechatInternalServerError = 503001 // 统一的微信模块的系统错误
)
