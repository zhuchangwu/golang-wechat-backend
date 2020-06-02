package task

import (
	"encoding/base64"
	"fmt"
	"weixin-backend/mail"
	"weixin-backend/mtStruct"
	"weixin-backend/qlu_no_safecode"
	"weixin-backend/sqlModel"
	"weixin-backend/util"
)

// 定时任务
func CornTask() {
	// 处理错误
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("出了错：", err)
			mail.DoSendMail("admin email", "出了错",err.(error).Error())
		}
	}()
	// 查询出所有的已经被激活的用户
	user, _ := sqlModel.FindAllUser()
	// 遍历所有的用户
	for i := 0; i < len(user); i++ {
		// 对密码解码
		decodeString, _ := base64.StdEncoding.DecodeString(user[i].StuPassword)
		re2, _ := util.AesDecrypt(decodeString, util.AesKey)
		password := string(re2)

		// 拿着每隔用户的信息去模拟登陆，获取到用户到成绩单
		scores, err := qlu_no_safecode.CronLogin(user[i].StuNumber, password, user[i].SchoolName+"_"+user[i].StuNumber)
		if err != nil {
			// 当前学生出错后，跳过它
			fmt.Printf("error : %v", err)
			continue
		}
		// 将网上成绩单最大的序号和用户表中的stu_max_score_record字段比较
		// 两个字段相等，说明没有当前同学的成绩单没有任何更新
		if user[i].StuMaxScoreRecord == len(scores)-1 {
			fmt.Println("当前同学的成绩为最新的状态")
			continue
		}
		// 网上的最大序号比stu_max_score_record多出来的部分，就是需要更新的部分
		if user[i].StuMaxScoreRecord < len(scores)-1 {
			// 将多出的部分写入mysql
			// 如果stu_max_score_record还是默认值，那么将本次获取的成绩全部保存进mysql
			if user[i].StuMaxScoreRecord == -1 {
				fmt.Println("当前同学第一次注册，给他发送一封邮件")
				// 保存进数据库
				// sqlModel.InsertScores(scores)
				// 发送邮件
				html := TracerseObjToHtml(scores)
				//mail.DoSendMail(user[i].StuEmail,"","系统已经")
				mail.DoSendMail(user[i].StuEmail, "同学，如下是你现有的成绩，后续成绩更新后我们同样会在第一时间通过此邮箱通知你！！！", html)
			} else {
				// 将新更新的成绩取出写入数据库
				scores := scores[user[i].StuMaxScoreRecord+1 : len(scores)]
				//sqlModel.InsertScores(scores)
				// 发送邮件
				html := TracerseObjToHtml(scores)
				mail.DoSendMail(user[i].StuEmail, "同学，出新成绩了！！！", html)
				fmt.Println("同学，出新成绩了")
			}
			// 更新user中的stu_max_score_record
			sqlModel.UpdateUserMaxScoreRecord(user[i].StuNumber, len(scores)-1)
		}
	}
}

// 用户激活后，执行这个任务
// 用户第一次登陆，拉去它的信息，然后发送第一个邮件
func FirstTask(username string) {
	// 查询用户的邮箱
	email, password, school_name := sqlModel.FindUserInfoByUsername(username)

	// 拿着每隔用户的信息去模拟登陆，获取到用户到成绩单
	scores, _ := qlu_no_safecode.CronLogin(username, password, school_name+"_"+username)

	// 如果stu_max_score_record还是默认值，那么将本次获取的成绩全部保存进mysql
	// 保存进数据库
	sqlModel.InsertScores(scores)

	// 发送邮件
	html := TracerseObjToHtml(scores)
	mail.DoSendMail(email, "同学，如下是你现有的成绩，后续成绩更新后我们同样会在第一时间通过此邮箱通知你！！！", html)
	// 更新user中的stu_max_score_record
	sqlModel.UpdateUserMaxScoreRecord(username, len(scores)-1)
	fmt.Println("第一次任务执行完了 ")
}

// 将obj转换成html
func TracerseObjToHtml(scores []mtStruct.Score) (html string) {
	str := "<table>" +
		"<tr>" +
		"<th>课程名</th>" +
		"<th>成绩</th>" +
		"<th>绩点</th>" +
		"<th>属性</th>" +
		"<th>开课学期</th>" +
		"</tr>"

	for i := 0; i < len(scores); i++ {
		// 如何课程名超过了个字
		cname := scores[i].CName
		if len(cname) > 27 {
			cname = cname[0:27] + "..."
		}
		str += "<tr>" +
			"<td>" + cname + "</td>" +
			"<td>" + scores[i].Score + "</td>" +
			"<td>" + scores[i].AchievementPoint + "</td>" +
			"<td>" + scores[i].Attribute + "</td>" +
			"<td>" + scores[i].Date + "</td>" +
			"</tr>"
	}
	str += "</table>"
	html = str
	return
}
