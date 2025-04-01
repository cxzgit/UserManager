package service

import (
	"UserManager/src/mapper"
	"UserManager/src/models"
	"UserManager/src/utils"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

// VerificationInfo 保存验证码及其过期时间

type RegisterService struct {
	Mapper              *mapper.RegisterMapper
	EmailService        *utils.EmailService
	VerificationService *utils.VerificationService
}

func NewRegisterService(rm *mapper.RegisterMapper, emailService *utils.EmailService, vService *utils.VerificationService) *RegisterService {
	return &RegisterService{
		Mapper:              rm,
		EmailService:        emailService,
		VerificationService: vService,
	}
}

// SendVerificationCode 生成验证码，并使用 Redis 存储（有效期5分钟）后发送邮件
func (rs *RegisterService) SendVerificationCode(email string) error {
	code, err := rs.VerificationService.GenerateAndStoreCode(email)
	if err != nil {
		return err
	}
	subject := "您的验证码"
	body := fmt.Sprintf("您的验证码为：%s，请勿泄露于他人!该验证码5分钟内有效!如非本人操作,请忽略此邮件!。", code)

	return rs.EmailService.SendEmail(email, subject, body)
}

// 用户注册
func (rs *RegisterService) RegisterUser(email, inputCode, password, passwordConfirm string) error {
	if password != passwordConfirm {
		return fmt.Errorf("两次密码输入不一致")
	}
	// 验证验证码
	if err := rs.VerificationService.VerifyCode(email, inputCode); err != nil {
		return err
	}
	// 检查邮箱是否已注册
	if user, _ := rs.Mapper.GetUserByEmail(email); user != nil {
		return fmt.Errorf("邮箱已注册")
	}
	// 加密密码
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败：%v", err)
	}
	// 插入新用户
	newUser := &models.User{
		Email:        email,
		PasswordHash: string(hashed),
	}
	if err := rs.Mapper.InsertUser(newUser); err != nil {
		return fmt.Errorf("用户注册失败：%v", err)
	}
	return nil
}
