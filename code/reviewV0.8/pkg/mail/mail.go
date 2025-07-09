package mail

import (
	"crypto/tls"

	"gopkg.in/gomail.v2"
)

type MailSetting struct {
	Host     string
	Port     int
	Username string
	Password string
	To       []string
}

func SendMail(mailConn MailSetting, body string, attachment ...string) error {
	m := gomail.NewMessage()
	// m.SetHeader("From", m.FormatAddress(mailConn.User, mailConn.Alias)) //设置邮件别名
	m.SetHeader("From", mailConn.Username) //设置邮件别名
	m.SetHeader("To", mailConn.To...)      //发送给多个用户
	m.SetHeader("Subject", "需要人工处理的订单")    //设置邮件主题
	m.SetBody("text/plain", body)          //设置邮件正文
	if len(attachment) > 0 {
		for _, v := range attachment {
			m.Attach(v) // 附件文件，可以是文件，照片，视频等等
		}
	}

	d := gomail.NewDialer(mailConn.Host, mailConn.Port, mailConn.Username, mailConn.Password)
	// 关闭SSL协议认证
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	return d.DialAndSend(m)
}
