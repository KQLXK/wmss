package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 统一响应结构体
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// 失败响应
func failResponse(w http.ResponseWriter, code int, msg string) {
	res := Response{Code: code, Msg: msg, Data: nil}
	httpStatus := mapCodeToHttpStatus(code)
	w.WriteHeader(httpStatus)
	httpx.OkJson(w, res)
}

// 业务码映射 HTTP 状态码
func mapCodeToHttpStatus(code int) int {
	switch code {
	case 401:
		return http.StatusUnauthorized
	case 403:
		return http.StatusForbidden
	case 500:
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}

// 自定义 Claims 结构
type CustomClaims struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// JWT 认证中间件
func JwtAuthMiddleware(secret string) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				failResponse(w, 401, "未携带 Authorization Token，请先登录")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				failResponse(w, 401, "Token 格式错误，请使用 Bearer + Token 格式（示例：Bearer eyJhbGciOiJIUzI1Ni...）")
				return
			}
			tokenStr := parts[1]

			// 解析 JWT Token
			token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})

			if err != nil {
				handleJWTError(w, r, err, tokenStr)
				return
			}

			if !token.Valid {
				failResponse(w, 401, "Token 无效")
				return
			}

			// 从 token 中提取 claims
			claims, ok := token.Claims.(*CustomClaims)
			if !ok {
				failResponse(w, 401, "Token claims 解析失败")
				return
			}

			// 存入上下文（供后续接口使用）
			ctx := context.WithValue(r.Context(), "user_id", claims.UserId)
			ctx = context.WithValue(ctx, "username", claims.Username)
			r = r.WithContext(ctx)

			next(w, r)
		}
	}
}

// 处理 JWT 错误
func handleJWTError(w http.ResponseWriter, r *http.Request, err error, tokenStr string) {
	var msg string

	if err == jwt.ErrTokenExpired {
		msg = "Token 已过期，请重新登录获取有效 Token"
	} else if err == jwt.ErrTokenMalformed {
		msg = "Token 格式错误"
	} else if err == jwt.ErrTokenNotValidYet {
		msg = "Token 尚未生效"
	} else if err == jwt.ErrSignatureInvalid {
		msg = "Token 签名验证失败，请检查密钥是否正确"
	} else {
		msg = "授权失败：" + err.Error()
	}

	failResponse(w, 401, msg)

	logx.Errorw("JWT authorization failed",
		logx.Field("error", err.Error()),
		logx.Field("token", tokenStr),
		logx.Field("path", r.URL.Path),
		logx.Field("method", r.Method))
}
