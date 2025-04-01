package utils

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/redis/go-redis/v9"
	"math/rand"
	"net/smtp"
	"time"
)

type VerificationService struct {
	RedisClient *redis.Client
	Ctx         context.Context
}
type EmailService struct {
	From     string
	Password string
	Server   string
	Port     string
}

// NewEmailService 初始化邮件服务
func NewEmailService(email, password string) *EmailService {
	return &EmailService{
		From:     email,
		Password: password,
		Server:   "smtp.qq.com",
		Port:     "465", // TLS 端口
	}
}

// SendEmail 发送邮件
func (es *EmailService) SendEmail(to, subject, body string) error {
	msg := []byte("From: " + es.From + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body)

	// 创建 TLS 连接
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         es.Server,
	}

	conn, err := tls.Dial("tcp", es.Server+":"+es.Port, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS 连接失败: %v", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, es.Server)
	if err != nil {
		return fmt.Errorf("创建 SMTP 客户端失败: %v", err)
	}
	defer client.Close()

	// 认证
	auth := smtp.PlainAuth("", es.From, es.Password, es.Server)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP 认证失败: %v", err)
	}

	// 发送邮件
	if err := client.Mail(es.From); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	return w.Close()
}

// NewVerificationService 初始化验证码服务
func NewVerificationService(redisClient *redis.Client) *VerificationService {
	return &VerificationService{
		RedisClient: redisClient,
		Ctx:         context.Background(),
	}
}

// GenerateAndStoreCode 生成验证码并存入 Redis
func (vs *VerificationService) GenerateAndStoreCode(email string) (string, error) {
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	expiry := 5 * time.Minute
	redisKey := fmt.Sprintf("verification:%s", email)

	err := vs.RedisClient.Set(vs.Ctx, redisKey, code, expiry).Err()
	if err != nil {
		return "", fmt.Errorf("验证码存储失败")
	}
	return code, nil
}

// VerifyCode 验证用户输入的验证码
func (vs *VerificationService) VerifyCode(email, inputCode string) error {
	redisKey := fmt.Sprintf("verification:%s", email)
	storedCode, err := vs.RedisClient.Get(vs.Ctx, redisKey).Result()

	if err == redis.Nil {
		return fmt.Errorf("验证码已过期或不存在")
	} else if err != nil {
		return fmt.Errorf("验证码验证失败")
	}

	if inputCode != storedCode {
		return fmt.Errorf("验证码错误")
	}

	// 删除验证码，防止重复使用
	vs.RedisClient.Del(vs.Ctx, redisKey)
	return nil
}
