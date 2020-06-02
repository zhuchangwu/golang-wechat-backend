package mail

import (
	"strconv"
)
import "gopkg.in/gomail.v2"

func SendMail(mailTo []string, subject string, body string) error {
	//定义邮箱服务器连接信息，pass = qq邮箱填授权码

	mailConn := map[string]string{
		"user": "646450308@qq.com",
		"pass": "trsoqkbkmgijbccd",
		"host": "smtp.qq.com",
		"port": "465",
	}
	port, _ := strconv.Atoi(mailConn["port"])
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(mailConn["user"], "自动化成绩查询"))
	m.SetHeader("To", mailTo...)    //发送给多个用户
	m.SetHeader("Subject", subject) //设置邮件主题
	 m.SetBody("text/html", body)    //设置邮件正文
	d := gomail.NewDialer(mailConn["host"], port, mailConn["user"], mailConn["pass"])
	err := d.DialAndSend(m)
	return err
}

	/*
	发送邮件
	stuEmail：学生的邮箱
	subject：标题
	body：发送的内容
  */
func DoSendMail(stuEmail , subject, body string) (e error) {
	mailTo := []string{stuEmail}
	err := SendMail(mailTo, subject, body)
	if err != nil {
		e = err
		return e
	}
	return nil
}

//func main() {
//	//定义收件人
//	mailTo := []string{
//		"2693667388@qq.com",
//		"1969509030@qq.com",
//	}
//	//邮件主题为"Hello"
//	subject := "Hi 出成绩了"
//	// 邮件正文
//	body := "请查收您的新成绩"
//
//	err := SendMail(mailTo, subject, body)
//	if err != nil {
//		log.Println(err)
//		fmt.Println("send fail")
//		return
//	}
//	fmt.Println("send successfully")
//}
