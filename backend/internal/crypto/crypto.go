// Package crypto 提供 LLM API Key 等敏感字段的静态加密与展示脱敏。
//
// 加密格式：enc:v1:<base64(nonce || ciphertext)>，使用 AES-256-GCM。
// 密钥派生自环境变量 AIOPS_SECRET_KEY（SHA-256）；未配置时使用内置开发默认
// 密钥并输出警告日志，生产环境务必通过环境变量注入独立密钥。
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"log"
	"os"
	"strings"
	"sync"
)

const (
	// Prefix 密文前缀，用于区分加密值与历史明文
	Prefix = "enc:v1:"

	defaultSecret = "aiops-dev-secret-do-not-use-in-prod"
)

var (
	keyOnce sync.Once
	key     []byte
)

func secretKey() []byte {
	keyOnce.Do(func() {
		s := os.Getenv("AIOPS_SECRET_KEY")
		if s == "" {
			log.Println("[crypto] 警告: 未设置 AIOPS_SECRET_KEY，使用内置开发默认密钥；生产环境请通过环境变量注入独立密钥")
			s = defaultSecret
		}
		sum := sha256.Sum256([]byte(s))
		key = sum[:]
	})
	return key
}

// Encrypt 加密明文，输出带 enc:v1: 前缀的密文字符串
func Encrypt(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	block, err := aes.NewCipher(secretKey())
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	ct := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return Prefix + base64.StdEncoding.EncodeToString(ct), nil
}

// Decrypt 解密 enc:v1: 密文；输入非加密格式（历史明文）时原样返回
func Decrypt(s string) (string, error) {
	if s == "" || !strings.HasPrefix(s, Prefix) {
		return s, nil
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(s, Prefix))
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(secretKey())
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", errors.New("密文长度不足")
	}
	nonce, ct := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", errors.New("API Key 解密失败，请检查 AIOPS_SECRET_KEY 是否与加密时一致")
	}
	return string(plain), nil
}

// Mask 返回脱敏展示值：前 7 位 + **** + 后 4 位；短值仅保留前 2 位
func Mask(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return key[:2] + "****"
	}
	return key[:7] + "****" + key[len(key)-4:]
}
