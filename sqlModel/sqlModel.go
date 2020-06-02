package sqlModel

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"weixin-backend/util"

	// todo 我得整明白这为什么不用，还导入？
	_ "github.com/go-sql-driver/mysql"
	"weixin-backend/mtStruct"
)

var dbdb *sql.DB

func init() {
	db, err := sql.Open("mysql", "root:qwer1010..@tcp(139.9.92.235:3306)/spider")
	if err != nil {
		fmt.Printf("error : %v", err)
		return
	}
	db.SetMaxOpenConns(500)
	db.SetMaxIdleConns(200)
	dbdb = db
	//err = db.Ping()
	if err != nil {
		fmt.Printf("error : %v", err)
		return
	}
}

/** 学生注册
	stu_number : 学号
 	password: 教务系统的密码
	stuName： 学生的姓名
 	schoolName：学校名
	重复注册报：Duplicate entry '201708120035' for key 'users_stu_number_uindex'
*/
func ResgisterUser(stu_number, password, stuName, schoolName, stuEmail string) (e error) {
	sqlStr := "insert into users(id,stu_number,stu_password,stu_name,school_name,stu_email,status,status_code,stu_max_score_record) values(?,?,?,?,?,?,?,?,?)"
	// 预留出验证码
	status_code := stu_number[(len(stu_number) - 4):len(stu_number)]
	// 加密密码
	re1, _ := util.AesEncrypt([]byte(password), util.AesKey)
	password =base64.StdEncoding.EncodeToString(re1)
	result, err := dbdb.Exec(sqlStr, nil, stu_number, password, stuName, schoolName, stuEmail, 1, status_code, -1)
	if err != nil {
		fmt.Printf("error : %v", err)
		e = err
		return
	}
	id, err := result.LastInsertId()
	if err != nil {
		fmt.Printf("error : %v", err)
		e = err
		return
	}
	if id < 0 {
		return fmt.Errorf("数据库挂了，用户注册失败")
	}
	return nil
}

func ActiveAccount(username, safe_code string) (e error) {
	// 查看是否存在这条记录
	sqlStr := "select id from users where stu_number = ? and status_code = ? and status = 1"
	row := dbdb.QueryRow(sqlStr, username, safe_code)
	id := -10
	row.Scan(&id)
	if id == -10 {
		return fmt.Errorf("同学，先注册才能激活，或者不能重复激活")
	}
	// 激活
	sqlStr = "update users set status = 0 where stu_number = ? and status_code = ?"
	_, err := dbdb.Exec(sqlStr, username, safe_code)
	if err != nil {
		return err
	}
	return nil
}

// 查询用户的学校名
func FindUserSchoolName(username, password string) string {
	// 对用户传过来铭文加密，得到结果和数据库中的密码一致
	re1, _ := util.AesEncrypt([]byte(password), util.AesKey)
	password =base64.StdEncoding.EncodeToString(re1)

	// 查看用户是否存在， 如果存在的话开启goroutinue去删除信息，告诉用
	sqlStr := "select school_name from users where stu_number = ? and stu_password = ?"
	schoolName := ""
	res := dbdb.QueryRow(sqlStr, username, password)
	res.Scan(&schoolName)
	return schoolName
}

// 删除用户的信息
func UnActiveUser(username, password, schoolName string) {
	// 对用户传过来铭文加密，得到结果和数据库中的密码一致
	re1, _ := util.AesEncrypt([]byte(password), util.AesKey)
	password =base64.StdEncoding.EncodeToString(re1)
	// 删除用户信息
	sqlStr := "delete from users where stu_number = ? and stu_password = ?; "
	dbdb.Exec(sqlStr, username, password)

	// 删除成绩
	sqlStr = "delete from score where stu_identify = ?; "
	dbdb.Exec(sqlStr, schoolName+"_"+username)
}

// 查询出所有的激活过的user
func FindAllUser() (users []mtStruct.User, e error) {
	us := make([]mtStruct.User, 0)
	rows, err := dbdb.Query("select id,stu_number,stu_password,stu_name, school_name,stu_max_score_record,stu_email from users where status =?;", 0)
	if err != nil {
		e = err
		return nil, err
	}
	for rows.Next() {
		var u mtStruct.User
		err = rows.Scan(&u.Id, &u.StuNumber, &u.StuPassword, &u.StuName, &u.SchoolName, &u.StuMaxScoreRecord, &u.StuEmail)
		if err != nil {
			fmt.Printf("error : %v", err)
			return nil, err
		}
		us = append(us, u)
	}
	return us, err
}

// 批量插入user的score
func InsertScores(scores []mtStruct.Score) error {
	sqlStr := "insert into score(id,stu_identify,cdate,cnumber,cname,score,credit,ctime,achievement_point,examination_method,attribute,property,serial_number) values(?,?,?,?,?,?,?,?,?,?,?,?,?)"
	// 循环单条插入
	for i := 0; i <= len(scores)-1; i++ {
		_, err := dbdb.Exec(sqlStr, nil, scores[i].StuIdentify, scores[i].Date, scores[i].CNumber, scores[i].CName, scores[i].Score, scores[i].Credit, scores[i].Time, scores[i].AchievementPoint, scores[i].ExaminationMethod, scores[i].Attribute, scores[i].Property, scores[i].SerialNumber)
		if err != nil {
			fmt.Printf("error : %v", err)
			return err
		}
	}
	return nil
}

// 更新用户最大到成绩值
func UpdateUserMaxScoreRecord(stuNumber string, maxValue int) {
	sqlStr := "update users set stu_max_score_record = ? where stu_number =?"
	_, err := dbdb.Exec(sqlStr, maxValue, stuNumber)
	if err != nil {
		fmt.Printf("error : %v", err)
		return
	}
}

//  根据用户名查询用户信息
func FindUserInfoByUsername(username string) (email,password,school_name string) {
	sqlStr:="select stu_email,stu_password,school_name from users where stu_number = ?"
	row:=dbdb.QueryRow(sqlStr,username)
	row.Scan(&email,&password,&school_name)
	// 对密码进行解密
	decodeString, _ := base64.StdEncoding.DecodeString(password)
	re2, _ := util.AesDecrypt(decodeString, util.AesKey)
	password = string(re2)
	return
}

// 根据学生的唯一id，查询出该学生所有的成绩信息
func FindAllScoresByStuIndentify(stuIdentify string) ([]mtStruct.Score, error) {
	scores := make([]mtStruct.Score, 0)
	sqlStr := "select cname,serial_number from score where stu_identify=?"
	rows, err := dbdb.Query(sqlStr, stuIdentify)
	if err != nil {
		fmt.Printf("error : %v", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var s mtStruct.Score
		err = rows.Scan(&s.CName, &s.SerialNumber)
		if err != nil {
			fmt.Printf("error : %v", err)
			return nil, err
		}
		scores = append(scores, s)
	}
	return scores, nil
}
