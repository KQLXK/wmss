// internal/middleware/simplerecovery.go
package middleware

import (
	"errors"
	"net/http"
	"os"
	"runtime/debug"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 适配器：将http.Handler转换为rest.Middleware
func SimpleRecovery() rest.Middleware {
	recoveryHandler := createRecoveryHandler()

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			recoveryHandler(next).ServeHTTP(w, r)
		}
	}
}

func createRecoveryHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// 简化错误日志
					logx.WithContext(r.Context()).Errorf("[Recovery] panic recovered: %v", err)

					// 通过环境变量判断环境
					if os.Getenv("GO_ENV") == "dev" || os.Getenv("ENV") == "development" {
						logx.WithContext(r.Context()).Errorf("Stack: %s", debug.Stack())
					} else {
						// 生产环境只记录简要信息
						logx.WithContext(r.Context()).Errorf("Brief stack: %s", getBriefStack())
					}

					// 返回简洁的错误响应
					httpx.Error(w, errors.New("internal server error"))
					return
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// 获取简化的堆栈信息
func getBriefStack() string {
	stack := string(debug.Stack())
	// 只取前几行堆栈信息
	lines := strings.Split(stack, "\n")
	if len(lines) > 8 {
		return strings.Join(lines[:8], "\n") + "\n..."
	}
	return stack
}
