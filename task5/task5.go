package main

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type User1 struct {
	Username   string `sql:"username"`
	Departname string `sql:"departname"`
	Status     int64  `sql:"status"`
}

type SimpleORM struct {
	Db           *sql.DB
	TableName    string
	Prepare      string
	AllExec      []interface{}
	Sql          string
	WhereParam   string
	LimitParam   string
	OrderParam   string
	OrWhereParam string
	WhereExec    []interface{}
	UpdateParam  string
	UpdateExec   []interface{}
	FieldParam   string
}

// 新建Mysql连接
func NewMysql(Username string, Password string, Address string, Dbname string) (*SimpleORM, error) {
	dsn := Username + ":" + Password + "@tcp(" + Address + ")/" + Dbname + "?charset=utf8&timeout=5s&readTimeout=6s"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	//最大连接数等配置，先占个位
	//db.SetMaxOpenConns(3)
	//db.SetMaxIdleConns(3)

	return &SimpleORM{
		Db:         db,
		FieldParam: "*",
	}, nil
}

// 设置表名
func (e *SimpleORM) Table(name string) *SimpleORM {
	e.TableName = name

	//重置引擎
	e.resetSmallormEngine()
	return e
}

func (e *SimpleORM) resetSmallormEngine() {
	// Reset all the fields of SmallormEngine to their zero values
	e.Prepare = ""
	e.AllExec = nil
	e.Sql = ""
	e.WhereParam = ""
	e.LimitParam = ""
	e.OrderParam = ""
	e.OrWhereParam = ""
	e.WhereExec = nil
	e.UpdateParam = ""
	e.UpdateExec = nil
}

func (e *SimpleORM) GetTable() string {
	return e.TableName
}

// 批量插入
func (e *SimpleORM) BatchInsert(data interface{}) (int64, error) {
	return e.batchInsertData(data, "insert")
}

// 批量替换插入
func (e *SimpleORM) BatchReplace(data interface{}) (int64, error) {
	return e.batchInsertData(data, "replace")
}

// 批量插入
func (e *SimpleORM) batchInsertData(batchData interface{}, insertType string) (int64, error) {

	//反射解析
	getValue := reflect.ValueOf(batchData)

	//切片大小
	l := getValue.Len()

	//字段名
	var fieldName []string

	//占位符
	var placeholderString []string

	//循环判断
	for i := 0; i < l; i++ {
		value := getValue.Index(i) // Value of item
		typed := value.Type()      // Type of item
		if typed.Kind() != reflect.Struct {
			panic("批量插入的子元素必须是结构体类型")
		}

		num := value.NumField()

		//子元素值
		var placeholder []string
		//循环遍历子元素
		for j := 0; j < num; j++ {

			//小写开头，无法反射，跳过
			if !value.Field(j).CanInterface() {
				continue
			}

			//解析tag，找出真实的sql字段名
			sqlTag := typed.Field(j).Tag.Get("sql")
			if sqlTag != "" {
				//跳过自增字段
				if strings.Contains(strings.ToLower(sqlTag), "auto_increment") {
					continue
				} else {
					//字段名只记录第一个的
					if i == 1 {
						fieldName = append(fieldName, strings.Split(sqlTag, ",")[0])
					}
					placeholder = append(placeholder, "?")
				}
			} else {
				//字段名只记录第一个的
				if i == 1 {
					fieldName = append(fieldName, typed.Field(j).Name)
				}
				placeholder = append(placeholder, "?")
			}

			//字段值
			e.AllExec = append(e.AllExec, value.Field(j).Interface())
		}

		//子元素拼接成多个()括号后的值
		placeholderString = append(placeholderString, "("+strings.Join(placeholder, ",")+")")
	}

	//拼接表，字段名，占位符
	e.Prepare = insertType + " into " + e.GetTable() + " (" + strings.Join(fieldName, ",") + ") values " + strings.Join(placeholderString, ",")

	//prepare
	var stmt *sql.Stmt
	var err error
	stmt, err = e.Db.Prepare(e.Prepare)
	if err != nil {
		return 0, e.setErrorInfo(err)
	}

	//执行exec,注意这是stmt.Exec
	result, err := stmt.Exec(e.AllExec...)
	if err != nil {
		return 0, e.setErrorInfo(err)
	}

	//获取自增ID
	id, _ := result.LastInsertId()
	return id, nil
}

func (e *SimpleORM) insertData(data interface{}, insertType string) (int64, error) {
	v := reflect.ValueOf(data)
	t := reflect.TypeOf(data)

	var placeholder []string
	var fieldName []string

	for i := 0; i < t.NumField(); i++ {
		if !v.Field(i).CanInterface() {
			continue
		}
		sqlTag := t.Field(i).Tag.Get("sql")
		if sqlTag != "" {
			if strings.Contains(strings.ToLower(sqlTag), "auto_increment") {
				continue
			} else {
				fieldName = append(fieldName, strings.Split(sqlTag, ",")[0])
				placeholder = append(placeholder, "?")
			}
		} else {
			fieldName = append(fieldName, t.Field(i).Name)
			placeholder = append(placeholder, "?")
		}
		e.AllExec = append(e.AllExec, v.Field(i).Interface())
	}
	e.Prepare = insertType + " into " + e.GetTable() + " (" + strings.Join(fieldName, ",") + ") values " + "(" + strings.Join(placeholder, ",") + ")"

	var stmt *sql.Stmt
	var err error

	stmt, err = e.Db.Prepare(e.Prepare)
	if err != nil {
		return 0, e.setErrorInfo(err)
	}
	result, err := stmt.Exec(e.AllExec...)
	if err != nil {
		return 0, e.setErrorInfo(err)
	}
	id, _ := result.LastInsertId()

	return id, nil
}

// 自定义错误格式
func (e *SimpleORM) setErrorInfo(err error) error {
	_, file, line, _ := runtime.Caller(1)
	return errors.New("File: " + file + ":" + strconv.Itoa(line) + ", " + err.Error())
}

// 插入
func (e *SimpleORM) Insert(data interface{}) (int64, error) {

	//判断是批量还是单个插入
	getValue := reflect.ValueOf(data).Kind()
	if getValue == reflect.Struct {
		return e.insertData(data, "insert")
	} else if getValue == reflect.Slice || getValue == reflect.Array {
		return e.batchInsertData(data, "insert")
	} else {
		return 0, errors.New("插入的数据格式不正确，单个插入格式为: struct，批量插入格式为: []struct")
	}
}

// 替换插入
func (e *SimpleORM) Replace(data interface{}) (int64, error) {
	//判断是批量还是单个插入
	getValue := reflect.ValueOf(data).Kind()
	if getValue == reflect.Struct {
		return e.insertData(data, "replace")
	} else if getValue == reflect.Slice || getValue == reflect.Array {
		return e.batchInsertData(data, "replace")
	} else {
		return 0, errors.New("插入的数据格式不正确，单个插入格式为: struct，批量插入格式为: []struct")
	}
}

// 传入and条件
func (e *SimpleORM) Where(data ...interface{}) *SimpleORM {

	//判断是结构体还是多个字符串
	var dataType int
	if len(data) == 1 {
		dataType = 1
	} else if len(data) == 2 {
		dataType = 2
	} else if len(data) == 3 {
		dataType = 3
	} else {
		panic("参数个数错误")
	}

	//多次调用判断
	if e.WhereParam != "" {
		e.WhereParam += " and ("
	} else {
		e.WhereParam += "("
	}

	//如果是结构体
	if dataType == 1 {
		t := reflect.TypeOf(data[0])
		v := reflect.ValueOf(data[0])

		//字段名
		var fieldNameArray []string

		//循环解析
		for i := 0; i < t.NumField(); i++ {

			//首字母小写，不可反射
			if !v.Field(i).CanInterface() {
				continue
			}

			//解析tag，找出真实的sql字段名
			sqlTag := t.Field(i).Tag.Get("sql")
			if sqlTag != "" {
				fieldNameArray = append(fieldNameArray, strings.Split(sqlTag, ",")[0]+"=?")
			} else {
				fieldNameArray = append(fieldNameArray, t.Field(i).Name+"=?")
			}

			e.WhereExec = append(e.WhereExec, v.Field(i).Interface())
		}

		//拼接
		e.WhereParam += strings.Join(fieldNameArray, " and ") + ") "

	} else if dataType == 2 {
		//直接=的情况
		e.WhereParam += data[0].(string) + "=?) "
		e.WhereExec = append(e.WhereExec, data[1])
	} else if dataType == 3 {
		//3个参数的情况
		//区分是操作符in的情况
		data2 := strings.Trim(strings.ToLower(data[1].(string)), " ")
		if data2 == "in" || data2 == "not in" {
			//判断传入的是切片
			reType := reflect.TypeOf(data[2]).Kind()
			if reType != reflect.Slice && reType != reflect.Array {
				panic("in/not in 操作传入的数据必须是切片或者数组")
			}

			//反射值
			v := reflect.ValueOf(data[2])
			//数组/切片长度
			dataNum := v.Len()
			//占位符
			ps := make([]string, dataNum)
			for i := 0; i < dataNum; i++ {
				ps[i] = "?"
				e.WhereExec = append(e.WhereExec, v.Index(i).Interface())
			}

			//拼接
			e.WhereParam += data[0].(string) + " " + data2 + " (" + strings.Join(ps, ",") + ")) "

		} else {
			e.WhereParam += data[0].(string) + " " + data[1].(string) + " ?) "
			e.WhereExec = append(e.WhereExec, data[2])
		}
	}

	return e
}

// 删除
func (e *SimpleORM) Delete() (int64, error) {

	//拼接delete sql
	e.Prepare = "delete from " + e.GetTable()

	//如果where不为空
	if e.WhereParam != "" || e.OrWhereParam != "" {
		e.Prepare += " where " + e.WhereParam + e.OrWhereParam
	}

	//limit不为空
	if e.LimitParam != "" {
		e.Prepare += "limit " + e.LimitParam
	}

	//第一步：Prepare
	var stmt *sql.Stmt
	var err error
	stmt, err = e.Db.Prepare(e.Prepare)
	if err != nil {
		return 0, err
	}

	e.AllExec = e.WhereExec

	//第二步：执行exec,注意这是stmt.Exec
	result, err := stmt.Exec(e.AllExec...)
	if err != nil {
		return 0, e.setErrorInfo(err)
	}

	//影响的行数
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, e.setErrorInfo(err)
	}

	return rowsAffected, nil
}

// 更新
func (e *SimpleORM) Update(data ...interface{}) (int64, error) {

	//判断是结构体还是多个字符串
	var dataType int
	if len(data) == 1 {
		dataType = 1
	} else if len(data) == 2 {
		dataType = 2
	} else {
		return 0, errors.New("参数个数错误")
	}

	//如果是结构体
	if dataType == 1 {
		t := reflect.TypeOf(data[0])
		v := reflect.ValueOf(data[0])

		var fieldNameArray []string
		for i := 0; i < t.NumField(); i++ {

			//首字母小写，不可反射
			if !v.Field(i).CanInterface() {
				continue
			}

			//解析tag，找出真实的sql字段名
			sqlTag := t.Field(i).Tag.Get("sql")
			if sqlTag != "" {
				fieldNameArray = append(fieldNameArray, strings.Split(sqlTag, ",")[0]+"=?")
			} else {
				fieldNameArray = append(fieldNameArray, t.Field(i).Name+"=?")
			}

			e.UpdateExec = append(e.UpdateExec, v.Field(i).Interface())
		}
		e.UpdateParam += strings.Join(fieldNameArray, ",")

	} else if dataType == 2 {
		//直接=的情况
		e.UpdateParam += data[0].(string) + "=?"
		e.UpdateExec = append(e.UpdateExec, data[1])
	}

	//拼接sql
	e.Prepare = "update " + e.GetTable() + " set " + e.UpdateParam

	//如果where不为空
	if e.WhereParam != "" || e.OrWhereParam != "" {
		e.Prepare += " where " + e.WhereParam + e.OrWhereParam
	}

	//limit不为空
	if e.LimitParam != "" {
		e.Prepare += "limit " + e.LimitParam
	}

	//prepare
	var stmt *sql.Stmt
	var err error
	stmt, err = e.Db.Prepare(e.Prepare)
	if err != nil {
		return 0, e.setErrorInfo(err)
	}

	//合并UpdateExec和WhereExec
	if e.WhereExec != nil {
		e.AllExec = append(e.UpdateExec, e.WhereExec...)
	}

	//执行exec,注意这是stmt.Exec
	result, err := stmt.Exec(e.AllExec...)
	if err != nil {
		return 0, e.setErrorInfo(err)
	}

	//影响的行数
	id, _ := result.RowsAffected()
	return id, nil
}
func main() {
	user2 := User1{
		Username:   "EE",
		Departname: "22",
		Status:     1,
	}
	user3 := User1{

		Username:   "EE1",
		Departname: "23",
		Status:     0,
	}
	e, er := NewMysql("root", "root", "localhost:3306", "record")
	if er != nil {
		fmt.Println(er)
	}

}
