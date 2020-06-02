package wx

import (
	"crypto/sha1"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/clbanning/mxj"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"weixin-backend/mail"
	"weixin-backend/qlu_no_safecode"
	"weixin-backend/sqlModel"
	"weixin-backend/task"
)

/*
	对微信发送过来的请求的封装， 每一次微信的亲戚都会携带如下的面参数
	参照第二步，按要求对请求进行参数进行验证： https://developers.weixin.qq.com/doc/offiaccount/Basic_Information/Access_Overview.html
*/
type weixinQuery struct {
	Signature    string `json:"signature"`
	Timestamp    string `json:"timestamp"`
	Nonce        string `json:"nonce"`
	EncryptType  string `json:"encrypt_type"`
	MsgSignature string `json:"msg_signature"`
	Echostr      string `json:"echostr"` // 只有在第一次请求时才存在
}

/*
	自定义的WX结构体
*/
type WeixinClient struct {
	Token          string
	Query          weixinQuery
	Message        map[string]interface{}
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Methods        map[string]func() bool
}

func NewClient(r *http.Request, w http.ResponseWriter, token string) (*WeixinClient, error) {
	// 实例话微信结构体指针
	weixinClient := new(WeixinClient)
	// 初始化
	weixinClient.Token = token
	weixinClient.Request = r
	weixinClient.ResponseWriter = w

	weixinClient.initWeixinQuery()

	// 验证签名
	if weixinClient.Query.Signature != weixinClient.signature() {
		return nil, errors.New("Invalid Signature.")
	}

	return weixinClient, nil
}

// 使用go的http包，将wx发送过来的请求中的参数获取出来
func (this *WeixinClient) initWeixinQuery() {

	var q weixinQuery
	// 随机字符串，用于签名使用
	q.Nonce = this.Request.URL.Query().Get("nonce")
	// 返回给微信的字符串
	q.Echostr = this.Request.URL.Query().Get("echostr")
	q.Signature = this.Request.URL.Query().Get("signature")
	q.Timestamp = this.Request.URL.Query().Get("timestamp")
	// 加密解密的方式
	q.EncryptType = this.Request.URL.Query().Get("encrypt_type")
	// 加密解密的消息体的签名
	q.MsgSignature = this.Request.URL.Query().Get("msg_signature")

	this.Query = q
}

// 生成签名： https://developers.weixin.qq.com/doc/offiaccount/Basic_Information/Access_Overview.html
// 参照如上的链接， 将微信发送过来的token，时间，随机串，排序生成
func (this *WeixinClient) signature() string {

	strs := sort.StringSlice{this.Token, this.Query.Timestamp, this.Query.Nonce}
	sort.Strings(strs)
	str := ""
	for _, s := range strs {
		str += s
	}
	// 使用sha1进行加密
	h := sha1.New()
	h.Write([]byte(str))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// 解析post数据，并将解析出来的数据封装进上面自定义的map中
func (this *WeixinClient) initMessage() error {
	// 读取请求体
	body, err := ioutil.ReadAll(this.Request.Body)

	if err != nil {
		return err
	}
	// 使用第三方包解析xml格式的请求体
	m, err := mxj.NewMapXml(body)

	if err != nil {
		return err
	}

	if _, ok := m["xml"]; !ok {
		return errors.New("Invalid Message.")
	}

	message, ok := m["xml"].(map[string]interface{})

	if !ok {
		return errors.New("Invalid Field `xml` Type.")
	}

	this.Message = message

	log.Println(this.Message)

	return nil
}

func (this *WeixinClient) text() {

	inMsg, ok := this.Message["Content"].(string)

	if !ok {
		return
	}

	var reply TextMessage

	reply.InitBaseData(this, "text")

	// 回复给用户的信息
	var resurtStr string

	inMsg = fmt.Sprint(inMsg)
	// 解析收到的消息
	if inMsg == "自动化成绩提示" {
		resurtStr = `				同学你好，欢迎开通自动化成绩提示，开通后，机器人将每隔30分钟自动化查询一次您的成绩单，当老师更新成绩后, 系统将在第一时间通过邮件的方式提示您。保证你在第一时间得到考试结果。（目前仅支持 齐鲁工业大学）
​	    
     开通请回复：成绩订阅&学号&登陆教务系统的密码&昵称&QQ邮箱&学校名

​	    例如： 成绩订阅&201708120000&qwer00&张三&646450308@qq.com&齐鲁工业大学 

		  订阅成功后，将收到系统推送的邮件。

		  取消请回复：取消订阅&学号&密码

​		  免责声明：本服务免费，且将对您的信息加密后妥善存储，郑重承诺不会盗用，转卖您的任何信息，但是出现其他不可抗拒的意外，本服务盖不负责。`
	} else if strings.HasPrefix(inMsg, "成绩订阅") {
		// 检验格式
		split := strings.Split(inMsg, "&")
		// 如果长度不为6，说明想开通提示，但是格式填写错误了
		if len(split) != 6 {
			resurtStr = "您输入的格式有误，请检查后重新输入"
			goto GOTOTAG
		}
		// 校验qq邮箱格式是否正确
		if !strings.HasSuffix(split[4], "@qq.com") {
			resurtStr = "请输入正确格式的qq邮箱"
			goto GOTOTAG
		}
		/*
			index:= 0 ,value:= 成绩提示
			index:= 1 ,value:= 学号
			index:= 2 ,value:= 登陆教务系统的密码
			index:= 3 ,value:= 昵称
			index:= 4 ,value:= QQ邮箱
			index:= 5 ,value:= 学校名
		*/
		// 使用账号密码尝试登陆，校验账号密码的正确性
		username := split[1]
		password := split[2]
		nickname := split[3]
		qqemail := split[4]
		schoolname := split[5]
		err := qlu_no_safecode.MoniLogin(strings.Trim(username, " "), strings.Trim(password, " "))
		if err != nil {
			resurtStr = fmt.Sprint(err)
			goto GOTOTAG
		}
		// 登陆没问题，就注册，并校验是否已经注册过了
		err = sqlModel.ResgisterUser(username, password, nickname, schoolname, qqemail)
		if strings.HasPrefix(fmt.Sprint(err), "Error 1062: Duplicate") {
			resurtStr = "同学，不能重复注册哦"
			goto GOTOTAG
		}
		if err != nil {
			resurtStr = fmt.Sprint("注册失败了，请联系管理员 ", err)
			goto GOTOTAG
		}

		// 回复用户注册成功了
		resurtStr = "你好 " + nickname + `，你已经完成初步订阅，稍后请查收QQ邮件中的验证码。
回复：激活&学号&验证码 
将正式享受自动化服务`

		// 发送验证邮件
		go mail.DoSendMail(qqemail, "自动化成绩推送服务验证", username[(len(username)-4):len(username)])

	} else if strings.HasPrefix(inMsg, "取消订阅") {
		// 删除用户在数据库中全部信息 取消订阅&学号&密码`
		// 校验格式
		msgs := strings.Split(inMsg, "&")
		if len(msgs) != 3 {
			resurtStr = `如需取消订阅，请输入正确的格式：
取消订阅&学号&密码`
			goto GOTOTAG
		}
		username := msgs[1]
		password := msgs[2]
		// 删除用户所有的信息
		schoolName := sqlModel.FindUserSchoolName(username, password)
		if schoolName == "" {
			resurtStr = "系统没有您的信息，不能注销，如有问题，请联系管理员。"
			goto GOTOTAG
		}
		// 异步删除用户信息
		go sqlModel.UnActiveUser(username,password,schoolName)
		// 回复用户
		resurtStr = "谢谢使用，您已成功取消订阅，系统将不再保留您的任何个人信息，祝您生活愉快！"
	} else if strings.HasPrefix(inMsg, "激活") {
		// 校验格式
		msgs := strings.Split(inMsg, "&")
		if len(msgs) != 3 {
			resurtStr = `如需激活，请输入正确的激活格式：
激活&学号&验证码`
			goto GOTOTAG
		}
		username := msgs[1]
		safe_code := msgs[2]
		// 激活校验
		err := sqlModel.ActiveAccount(username, safe_code)
		if err != nil {
			// 激活失败
			resurtStr = "激活失败：" + fmt.Sprint(err)
			goto GOTOTAG
		}
		// 发送第一封邮件
		go task.FirstTask(username)

		resurtStr = `激活成功，欢迎体验自动化服务
稍后您将收到系统向您第一封成绩推送的邮件，后续成绩有更新，我们将在第一时间向你推送`
	} else {
		resurtStr = `系统暂不理解你的请求。
为您提供自动化成绩查询服务：
（目前仅支持齐鲁工业大学）
开启请回复：自动化成绩提示
取消请回复：取消订阅&学号&密码`
	}
GOTOTAG:
	fmt.Println("-------")
	// 将消息转码，序列化准备发送回客户端
	reply.Content = value2CDATA(resurtStr)
	replyXml, err := xml.Marshal(reply)
	if err != nil {
		log.Println(err)
		this.ResponseWriter.WriteHeader(403)
		return
	}
	// 写回
	this.ResponseWriter.Header().Set("Content-Type", "text/xml")
	this.ResponseWriter.Write(replyXml)
}

func (this *WeixinClient) Run() {

	err := this.initMessage()

	if err != nil {

		log.Println(err)
		this.ResponseWriter.WriteHeader(403)
		return
	}

	MsgType, ok := this.Message["MsgType"].(string)

	if !ok {
		this.ResponseWriter.WriteHeader(403)
		return
	}

	switch MsgType {
	case "text":
		this.text()
		break
	default:
		break
	}
	return
}
