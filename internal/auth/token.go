package auth

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenManager struct {
	secret   []byte
	duration time.Duration
}

func NewTokenManager(secret []byte, duration time.Duration) *TokenManager {
	return &TokenManager{
		secret:   secret,
		duration: duration,
	}
}
func (tokenManager *TokenManager) Generate(userID uint64) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(tokenManager.duration)

	// RegisteredClaims 保存 JWT 的标准字段。这些内容可以被读取，不要放密码。
	claims := jwt.RegisteredClaims{
		// Subject 表示 token 的主体。将 uint64 用户 ID 转为十进制字符串。
		Subject: strconv.FormatUint(userID, 10),
		// NewNumericDate 将 Go 的时间转成 JWT 使用的时间格式。
		IssuedAt: jwt.NewNumericDate(now), // token 的签发时间
		// 验证 token 时需要检查这个过期时间。
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}

	// 创建 token 对象并选择 HS256 签名算法，此时还没有生成最终字符串。
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用密钥签名并生成 token 字符串；验证时也要用同一个密钥检查签名。
	tokenString, err := token.SignedString(tokenManager.secret)

	// 签名失败时，返回空字符串、零值时间和错误。
	if err != nil {
		return "", time.Time{}, err
	}

	// 签名成功后，返回 token 和它的过期时间。
	return tokenString, claims.ExpiresAt.Time, nil
}

func (tokenManager *TokenManager) Verify(tokenString string) (uint64, error) {
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			return tokenManager.secret, nil
		},
		// 只接受 HS256，不能由 token 自己决定允许什么算法。
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		// 必须包含 exp；JWT 库会检查是否已经过期。
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, errors.New("token无效")
	}

	userID, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil || userID == 0 {
		return 0, errors.New("token 中的用户 ID 无效")
	}
	return userID, nil
}
