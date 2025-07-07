package notification

import (
	"log"
	"testing"
)

func TestMailNotification_Send(t *testing.T) {
	// 配置 SMTP（使用 Zoho SMTP 服务器）
	config := MailConfig{
		Host:     "smtp.zoho.com",
		Port:     465, // Zoho SMTP 端口
		Username: "heath-cetide@zohomail.com",
		Password: "CTct288513832##",
		From:     "heath-cetide@zohomail.com",
	}

	// 创建邮件通知实例
	mail := NewMailNotification(config)

	// 测试发送邮件
	err := mail.Send("19511899044@163.com", "Test Subject", "This is a test email.")
	if err != nil {
		t.Errorf("Failed to send email: %v", err)
	} else {
		t.Log("Email sent successfully!")
	}
}

func TestMailNotification_Welcome(t *testing.T) {
	config := MailConfig{
		Host:     "smtp.zoho.com",
		Port:     587, // Zoho SMTP 端口
		Username: "heath-cetide@zohomail.com",
		Password: "CTct288513832##",
		From:     "heath-cetide@zohomail.com",
	}

	mailer := NewMailNotification(config)

	err := mailer.SendWelcomeEmail(
		"19511899044@163.com", // 收件人
		"小明",                  // 用户名
		"https://yourapp.com/verify?token=abc123", // 验证链接
	)
	if err != nil {
		log.Fatalf("邮件发送失败: %v", err)
	} else {
		log.Println("🎉 邮件发送成功！")
	}
}
