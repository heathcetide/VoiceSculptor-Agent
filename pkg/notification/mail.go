package notification

import (
	voiceSculptor "VoiceSculptor"
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
)

// MailConfig 邮件配置
type MailConfig struct {
	Host     string `json:"host"`     // SMTP 服务器地址
	Port     int64  `json:"port"`     // SMTP 服务器端口
	Username string `json:"username"` // SMTP 用户名
	Password string `json:"password"` // SMTP 密码
	From     string `json:"from"`     // 发件人邮箱
}

// MailNotification 邮件通知
type MailNotification struct {
	Config MailConfig
}

// NewMailNotification 创建邮件通知实例
func NewMailNotification(config MailConfig) *MailNotification {
	return &MailNotification{Config: config}
}

// Send 发送邮件
func (m *MailNotification) Send(to, subject, body string) error {
	// 邮件内容
	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body)

	// SMTP 认证
	auth := smtp.PlainAuth("", m.Config.Username, m.Config.Password, m.Config.Host)

	// 配置 TLS
	tlsConfig := &tls.Config{
		ServerName:         m.Config.Host, // 服务器名称
		InsecureSkipVerify: false,         // 不跳过证书验证
	}

	// 连接 SMTP 服务器
	addr := fmt.Sprintf("%s:%d", m.Config.Host, m.Config.Port)
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to dial SMTP server: %v", err)
	}
	defer conn.Close()

	// 创建 SMTP 客户端
	client, err := smtp.NewClient(conn, m.Config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %v", err)
	}
	defer client.Close()

	// 认证
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %v", err)
	}

	// 设置发件人和收件人
	if err = client.Mail(m.Config.From); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %v", err)
	}

	// 发送邮件内容
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to prepare data: %v", err)
	}
	defer w.Close()

	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("failed to write email content: %v", err)
	}

	return nil
}

func (m *MailNotification) SendHTML(to, subject, htmlBody string) error {
	msg := "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/html; charset=\"UTF-8\"\r\n"
	msg += fmt.Sprintf("From: %s\r\n", m.Config.From)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "\r\n" + htmlBody

	addr := fmt.Sprintf("%s:%d", m.Config.Host, m.Config.Port)

	auth := smtp.PlainAuth("", m.Config.Username, m.Config.Password, m.Config.Host)

	// smtp.SendMail 不支持 465（SSL），只能发给 STARTTLS 服务，或使用第三方库
	return smtp.SendMail(addr, auth, m.Config.From, []string{to}, []byte(msg))
}

// SendHTML sends an HTML email using the embedded welcome template
func (m *MailNotification) SendWelcomeEmail(to string, username string, verifyURL string) error {
	// Parse the embedded template
	tmpl, err := template.New("welcome").Parse(voiceSculptor.WelcomeHTML)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	data := struct {
		Username  string
		VerifyURL string
	}{
		Username:  username,
		VerifyURL: verifyURL,
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to render email body: %w", err)
	}

	// Build MIME email message
	msg := "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/html; charset=\"UTF-8\"\r\n"
	msg += fmt.Sprintf("From: %s\r\n", m.Config.From)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", "Welcome to Join VoiceSculptor！")
	msg += "\r\n" + body.String()

	// Zoho SMTP uses SSL (port 465), but net/smtp only supports STARTTLS (usually port 587)
	addr := fmt.Sprintf("%s:%d", m.Config.Host, m.Config.Port)
	auth := smtp.PlainAuth("", m.Config.Username, m.Config.Password, m.Config.Host)

	return smtp.SendMail(addr, auth, m.Config.From, []string{to}, []byte(msg))
}

func (m *MailNotification) SendVerificationCode(to, code string) error {
	tmpl, err := template.New("verification").Parse(voiceSculptor.VerificationHTML)
	if err != nil {
		return fmt.Errorf("failed to parse verification template: %w", err)
	}
	data := struct {
		Code string
	}{
		Code: code,
	}
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to render verification email: %w", err)
	}

	msg := "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/html; charset=\"UTF-8\"\r\n"
	msg += fmt.Sprintf("From: %s\r\n", m.Config.From)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", "Your VoiceSculptor Veritifaction Code")
	msg += "\r\n" + body.String()

	addr := fmt.Sprintf("%s:%d", m.Config.Host, m.Config.Port)
	auth := smtp.PlainAuth("", m.Config.Username, m.Config.Password, m.Config.Host)

	return smtp.SendMail(addr, auth, m.Config.From, []string{to}, []byte(msg))
}
