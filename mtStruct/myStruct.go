package mtStruct

// 用户
type User struct {
	Id          int `json:"id"`
	StuNumber   string `json:"学号"`
	StuPassword string `json:"密码"`
	StuName     string `json:"姓名"`
	SchoolName  string `json:"学校名"`
	StuEmail  string `json:"邮箱"`
	Status  string `json:"是否激活0为正常，1为未激活"`
	StatusCode  string `json:"激活码，用户输入后方可激活，默认是学号的后四位"`
	StuMaxScoreRecord int `json:"最大的成绩序号"`
}

// 成绩单
type Score struct {
	Id                int `json:"id"`
	StuIdentify       string `json:"身份信息"` // 学生对唯一标示：学校名_学号
	SerialNumber      int `json:"课程序号"` // 这个序号会递增，比较是否有新的课程时，使用这个序号大小进行比较
	Date              string `json:"开课学期"`
	CNumber           string `json:"课程编号"`
	CName             string `json:"课程名称"`
	Score             string `json:"成绩"`
	Credit            string `json:"学分"`
	Time              string `json:"总学时"`
	AchievementPoint  string `json:"绩点"`
	ExaminationMethod string `json:"考试方式"`
	Attribute         string `json:"课程属性"`
	Property          string `json:"课程性质"`
}