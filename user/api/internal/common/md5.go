package common

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
)

// EncryptPassword 密码加密（纯 MD5，32 位小写）
// 直接对原始密码进行 MD5 哈希，无盐值
func EncryptPassword(password string) string {
	h := md5.New()
	h.Write([]byte(password)) // 直接写入原始密码字节流
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyPassword 密码校验（无盐值版本）
// rawPassword: 用户输入的原始密码
// encryptedPassword: 数据库存储的 MD5 加密密码
func VerifyPassword(rawPassword, encryptedPassword string) bool {
	// 对原始密码执行相同的 MD5 加密，与存储值对比
	rawEncrypted := EncryptPassword(rawPassword)
	fmt.Println(rawEncrypted)
	fmt.Println(encryptedPassword)
	// 大小写不敏感对比（兼容可能的存储格式差异）
	return strings.EqualFold(rawEncrypted, encryptedPassword)
}

// 可选扩展：EncryptPasswordUpper 生成 32 位大写 MD5 密码
func EncryptPasswordUpper(password string) string {
	return strings.ToUpper(EncryptPassword(password))
}
