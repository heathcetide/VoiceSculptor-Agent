package util

import (
	"errors"
	"fmt"
	"net/http"
)

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e Error) StatusCode() int {
	return e.Code
}

func (e Error) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

var ErrUnauthorized = &Error{Code: http.StatusUnauthorized, Message: "unauthorized"}
var ErrAttachmentNotExist = &Error{Code: http.StatusNotFound, Message: "attachment not exist"}
var ErrNotAttachmentOwner = &Error{Code: http.StatusForbidden, Message: "not attachment owner"}

// 身份认证 & 注册相关错误

var ErrQuotaExceeded = errors.New("额度不足") // 用户使用额度已用完

var ErrLLMCallFailed = errors.New("调用语言模型失败") // 调用语言模型失败

var ErrEmptyPassword = errors.New("empty password") // 密码为空，通常用于注册或登录校验失败

var ErrEmptyEmail = errors.New("empty email") // 邮箱为空，通常用于注册、登录、找回密码等操作

var ErrSameEmail = errors.New("same email") // 新旧邮箱相同，用户尝试更换邮箱时触发

var ErrEmailExists = errors.New("email exists, please use another email") // 邮箱已存在，尝试注册或更新为已被注册的邮箱

var ErrUserNotExists = errors.New("user not exists") // 用户不存在，常用于登录、查询或操作不存在的用户时

var ErrForbidden = errors.New("forbidden access") // 拒绝访问，用户虽已登录但无权限访问目标资源

var ErrUserNotAllowLogin = errors.New("user not allow login") // 用户被禁止登录，可能是被管理员封禁

var ErrUserNotAllowSignup = errors.New("user not allow signup") // 用户被禁止注册，系统配置或策略限制注册行为

var ErrNotActivated = errors.New("user not activated") // 用户账户未激活，通常用于邮箱激活未完成

var ErrTokenRequired = errors.New("token required") // 缺少必要的令牌，例如访问受保护资源时

var ErrInvalidToken = errors.New("invalid token") // 令牌格式非法或不符合规范

var ErrBadToken = errors.New("bad token") // 令牌已被篡改、伪造或无效

var ErrTokenExpired = errors.New("token expired") // 令牌已过期

var ErrEmailRequired = errors.New("email required") // 邮箱字段必须提供但未提供

// 通用资源/数据处理相关错误

var ErrNotFound = errors.New("not found") // 请求的数据或资源未找到

var ErrNotChanged = errors.New("not changed") // 数据未发生变化，例如更新请求中没有实际变更字段

var ErrInvalidView = errors.New("with invalid view") // 请求使用了无效的视图标识或参数

// 权限与逻辑控制相关错误

var ErrOnlySuperUser = errors.New("only super user can do this") // 仅限超级用户执行的操作

var ErrInvalidPrimaryKey = errors.New("invalid primary key") // 主键非法，可能为格式错误或缺失

// Common errors
var (
	// Tools related errors
	ErrInvalidToolListFormat = errors.New("invalid tool list response format")
	ErrInvalidToolFormat     = errors.New("invalid tool format")
	ErrToolNotFound          = errors.New("tool not found")
	ErrInvalidToolParams     = errors.New("invalid tool parameters")

	// JSON-RPC related errors
	ErrParseJSONRPC           = errors.New("failed to parse JSON-RPC message")
	ErrInvalidJSONRPCFormat   = errors.New("invalid JSON-RPC format")
	ErrInvalidJSONRPCResponse = errors.New("invalid JSON-RPC response")
	ErrInvalidJSONRPCRequest  = errors.New("invalid JSON-RPC request")
	ErrInvalidJSONRPCParams   = errors.New("invalid JSON-RPC parameters")

	// Resource related errors
	ErrInvalidResourceFormat = errors.New("invalid resource format")
	ErrResourceNotFound      = errors.New("resource not found")

	// Prompt related errors
	ErrInvalidPromptFormat = errors.New("invalid prompt format")
	ErrPromptNotFound      = errors.New("prompt not found")

	// Tool manager errors
	ErrEmptyToolName         = errors.New("tool name cannot be empty")
	ErrToolAlreadyRegistered = errors.New("tool already registered")
	ErrToolExecutionFailed   = errors.New("tool execution failed")

	// Resource manager errors
	ErrEmptyResourceURI = errors.New("resource URI cannot be empty")

	// Prompt manager errors
	ErrEmptyPromptName = errors.New("prompt name cannot be empty")

	// Lifecycle manager errors
	ErrSessionAlreadyInitialized = errors.New("session already initialized")
	ErrSessionNotInitialized     = errors.New("session not initialized")

	// Parameter errors
	ErrInvalidParams = errors.New("invalid parameters")
	ErrMissingParams = errors.New("missing required parameters")

	// Client errors
	ErrAlreadyInitialized = errors.New("client already initialized")
	ErrNotInitialized     = errors.New("client not initialized")
	ErrInvalidServerURL   = errors.New("invalid server URL")
)
