package qlu_no_safecode

import (
	_ "database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/opesun/goquery"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
	"weixin-backend/js"
	"weixin-backend/mtStruct"
	"weixin-backend/util"
)

// 教务系统地址
var qluUrl = "http://jwxt.qlu.edu.cn/jsxsd/"

// 真正发送登陆请求的url
var LoginUrl = "http://jwxt.qlu.edu.cn/jsxsd/xk/LoginToXk"

// 访问登陆也，获取cookie
func GetCookieFromLoginhtml(url string) (cookie string, e error) {
	res, err := http.Get(url)
	if err != nil {
		e = err
	}

	// 获取cookie
	cookie = res.Header.Get("Set-Cookie")
	cookie = util.GetOneValueByPrefixAndSurfix("JSESSIONID=", "; Path=/", cookie)
	res.Body.Close()
	return
}

// 模拟js，拼接出加密参数
func getSalt(username, password string) (salt string, e error) {
	// 使用js对uri进行编码
	username, err := js.EncodeInp(username)
	if err != nil {
		e = err
	}
	password, err = js.EncodeInp(password)
	if err != nil {
		e = err
	}
	salt, err = js.EncodeUri(username + "%%%" + password)
	if e != nil {
		fmt.Printf("error : %v", err)
		return
	}
	return
}

// 模拟登陆
func login(salt, cookie string) (html string) {
	// todo 不能写成这样， 自己手动拼接
	// 构造请求中的from数据
	// data := url.Values{}
	// data.Set("encoded", salt)
	// data.Encode()

	req, err := http.NewRequest("POST", LoginUrl, strings.NewReader("encoded="+salt))
	// 添加请求头
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.163 Safari/537.36")
	req.Header.Add("Cookie", "JSESSIONID="+cookie)
	// todo 这里可以说一下那几种提交的方法
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	// 将请求体中的数据parse进from中
	// err = req.ParseForm()
	//if err != nil {
	//	fmt.Printf("url.Values中的值解析进From失败 : %v", err)
	//	return
	//}

	//发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("无密版qlu教务系统登陆请求失败 : %v", err)
		return
	}
	// todo 根据状态码判断下一步如何操作，如果状态码是302，表示操作成功
	fmt.Println("resp.Status:  ", resp.Status)

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error : %v", err)
		return
	}
	// 返回个人主页的html
	html = string(b)
	// 手动关闭
	resp.Body.Close()
	return
}

// 从mysql中加载用数据
func getUserInfoFromDB() (username, password string) {
	// todo 模拟从mysql中取出用户信息
	username = "201708120034"
	password = "xxx"
	return
}

// 获取当前学生所有成绩
func getAllScore(stuIdentify, cookie string) ([]mtStruct.Score, error) {
	// 发送查询成绩的请求
	u := "http://jwxt.qlu.edu.cn/jsxsd/kscj/cjcx_list"
	req, err := http.NewRequest("POST", u, strings.NewReader("kksj=&kcxz=&kcmc=&xsfs=all"))
	if err != nil {
		fmt.Printf("error : %v", err)
		return nil, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.163 Safari/537.36")
	req.Header.Add("Cookie", "JSESSIONID="+cookie)
	req.Header.Add("Referer", "http://jwxt.qlu.edu.cn/jsxsd/kscj/cjcx_query?Ves632DSdyV=NEW_XSD_XJCJ")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*;q=0.8,application/signed-exchange;v=b3;q=0.9")

	client := &http.Client{}
	resp, err := client.Do(req)
	time.Sleep(time.Second*4)
	if err != nil {
		fmt.Printf("error : %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error : %v", err)
		return nil, err
	}

	html := string(bytes)

	// 截取出table
	nodes, err := goquery.ParseString(html)
	if err != nil {
		fmt.Printf("error : %v", err)
		return nil, err
	}
	// 找到所有的成绩
	nodes = nodes.Find("tr")
	// 创建成绩切片：
	Scores := make([]mtStruct.Score, 0)
	for i := 2; i < nodes.Length(); i++ {
		// 获取到完整br
		record := nodes.Eq(i).Html()
		// 净化：去除空格，但保留换行
		html = util.DeleteExtraSpace(record)
		// 安装换行切割
		strs := strings.Split(html, "\n")
		// 去除注释行
		newStrs := make([]string, 0)
		for i := 0; i < len(strs); i++ {
			// 如果是注释行的话，舍弃
			if strings.HasPrefix(strs[i], "<--") || strs[i] == "" {
				continue
			}
			newStrs = append(newStrs, strs[i])
		}
		// strs[1] 课程序列号
		serialNumber := util.GetOneValueByPrefixAndSurfix("<td>", "</td>", newStrs[0])
		sn, _ := strconv.Atoi(serialNumber)

		// strs[2] 开课学期
		date := util.GetOneValueByPrefixAndSurfix("<td>", "</td>", newStrs[1])
		// strs[3] 课程编号
		cNumber := util.GetOneValueByPrefixAndSurfix(`<td align="left">`, "</td>", newStrs[2])
		// strs[4] 课程名称
		cName := util.GetOneValueByPrefixAndSurfix(`<td align="left">`, "</td>", newStrs[3])
		nos, _ := goquery.ParseString(newStrs[5])
		// strs[5] 成绩
		score := nos.Text()
		// strs[6] 学分
		credit := util.GetOneValueByPrefixAndSurfix("<td>", "</td>", newStrs[6])
		// strs[7] 总学时
		time := util.GetOneValueByPrefixAndSurfix("<td>", "</td>", newStrs[7])
		// strs[9] 绩点
		achievementPoint := util.GetOneValueByPrefixAndSurfix("<td>", "</td>", newStrs[9])
		// strs[10] 考核方式
		examinationMethod := util.GetOneValueByPrefixAndSurfix("<td>", "</td>", newStrs[10])
		// strs[11] 课程属性
		attribute := util.GetOneValueByPrefixAndSurfix("<td>", "</td>", newStrs[11])
		// strs[12] 课程性质
		property := util.GetOneValueByPrefixAndSurfix("<td>", "</td>", newStrs[12])

		s := mtStruct.Score{
			StuIdentify:       stuIdentify,
			SerialNumber:      sn,
			Date:              date,
			CNumber:           cNumber,
			CName:             cName,
			Score:             score,
			Credit:            credit,
			Time:              time,
			AchievementPoint:  achievementPoint,
			ExaminationMethod: examinationMethod,
			Attribute:         attribute,
			Property:          property,
		}
		Scores = append(Scores, s)
	}
	return Scores, nil
}

/*
	 模拟登陆，齐鲁工业大学的教务系统：
     判断用户给定的账户密码是否正确
	 err 返回nil，表示登陆成功了
*/
func MoniLogin(username, password string) (e error) {
	// 获取登陆页面
	cookie, err := GetCookieFromLoginhtml(qluUrl)
	if err != nil {
		fmt.Printf("获取登陆页面错误 : %v", err)
		return err
	}

	// 获取盐信息
	salt, err := getSalt(username, password)
	if err != nil {
		fmt.Printf("获取盐信息出错 : %v", err)
		return err
	}

	// 构建请求头
	req, err := http.NewRequest("POST", LoginUrl, strings.NewReader("encoded="+salt))
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.163 Safari/537.36")
	req.Header.Add("Cookie", "JSESSIONID="+cookie)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")

	//发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		e = fmt.Errorf("发送模拟登陆的请求失败了，请联系管理员 %v", err)
		return e
	}
	// 无论登陆成功还是失败，返回的都是200状态码，所以这里不根据状态吗判断
	fmt.Println("resp.Status:  ", resp.Status)

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		e = fmt.Errorf("模拟登陆时，读取响应体失败，请联系管理员 %v", err)
		return e
	}
	html := string(b)
	nodes, err := goquery.ParseString(html)
	if e != nil {
		e = fmt.Errorf("模拟登陆时，响应体转换成html失败，请联系管理员 %v", err)
		return e
	}
	text := nodes.Find("title").Text()
	// 返回个人主页的html
	fmt.Println(text)
	if text != "学生个人中心" {
		e = fmt.Errorf("用户名或者密码错误，请确认后重新输入")
		return e
	}
	resp.Body.Close()
	return nil
}

// 新用户第一次登陆，执行如下的逻辑
/**
作用：用户第一次登陆进来时，读取用户的成绩并存储起来
username：教务系统的账号
password：教务系统的密码
stuIedntify：单条成绩所属用户唯一标识: "齐鲁工业大学_201708120034"
*/
func CronLogin(username, password, stuIedntify string)([]mtStruct.Score, error) {
	// 获取登陆页面
	cookie, err := GetCookieFromLoginhtml(qluUrl)
	if err != nil {
		fmt.Printf("获取登陆页面错误 : %v", err)
		return nil,err
	}
	time.Sleep(time.Second*4)
	// 获取盐信息
	salt, err := getSalt(username, password)
	if err != nil {
		fmt.Printf("获取盐信息出错 : %v", err)
		return nil,err
	}

	// 发送登陆请求，获取个人信息页面
	_ = login(salt, cookie)
	time.Sleep(time.Second*4)

	// 获取所有的成绩
	scores, err := getAllScore(stuIedntify, cookie)
	time.Sleep(time.Second*4)

	if err != nil {
		fmt.Printf("error : %v", err)
		return nil,err
	}
	return scores,nil
}




