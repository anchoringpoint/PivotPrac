package simpleorm

import (
	"database/sql"
	"errors"
	"reflect"
	"strconv"
	"strings"
	_"github.com/go-sql-driver/mysql"
)
type Route struct {
	ID   int	`sql:"id,auto_increment"`
	Route string `sql:"route"`
	Origin string `sql:"origin"`
	Destination string	`sql:"destination"`
}
type Alias struct {
	ID   int `sql:"id,auto_increment"`
	Name string	`sql:"name"`
	Alias string	`sql:"alias"`
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

// NewMysql 函数用于新建一个Mysql连接
// 输入：
//       Username - Mysql用户名
//       Password - Mysql密码
//       Address - Mysql服务器地址
//       Dbname - 数据库名称
// 返回：
//       *SimpleORM - 新建的Mysql连接
//       error - 错误信息，如果没有错误则为nil
func NewMysql(Username string, Password string, Address string, Dbname string) (*SimpleORM, error) {
	dsn := Username + ":" + Password + "@tcp(" + Address + ")/" + Dbname + "?charset=utf8&timeout=5s&readTimeout=6s"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	//最大连接数等配置
	//db.SetMaxOpenConns(3)
	//db.SetMaxIdleConns(3)

	return &SimpleORM{
		Db:         db,
		FieldParam: "*",
	}, nil
}

// Table 函数用于选择要操作的表
// 输入：
//       name - 表名
// 返回：
//       *SimpleORM - 选定表的SimpleORM
func (e *SimpleORM) Table(name string) *SimpleORM {
	e.TableName = name

	//重置对表操作相关的所有参数
	e.resetSimpleORMParam()
	return e
}

// resetSimpleORMParam 函数用于重置对表操作相关的所有参数
// 返回：
//       *SimpleORM - 重置后的SimpleORM
func (e *SimpleORM) resetSimpleORMParam() {
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
	e.FieldParam = "*"//默认查询所有字段
}

// GetTable 函数用于获取当前操作的表名
// 返回：
//       string - 当前操作的表名
func (e *SimpleORM) GetTable() string {
	return e.TableName
}

// BatchInsert 函数用于批量插入数据
// 输入：
//       data - 要插入的数据，必须是结构体切片或数组
// 返回：
//       int64 - 插入的最后一条数据的自增ID
//       error - 错误信息，如果没有错误则为nil
func (e *SimpleORM) BatchInsert(data interface{}) (int64, error) {
	return e.batchInsertData(data, "insert")
}

// BatchReplace 函数用于批量替换数据
// 输入：
//       data - 要替换的数据，必须是结构体切片或数组
// 返回：
//       int64 - 替换的最后一条数据的自增ID
//       error - 错误信息，如果没有错误则为nil
func (e *SimpleORM) BatchReplace(data interface{}) (int64, error) {
	return e.batchInsertData(data, "replace")
}

// batchInsertData 函数用于批量插入或替换数据
// 输入：
//	   batchData - 要插入或替换的数据，必须是结构体切片或数组
//     insertType - 插入类型，"insert"为插入，"replace"为替换
// 返回：
//       int64 - 插入或替换的最后一条数据的自增ID
//       error - 错误信息，如果没有错误则为nil
func (e *SimpleORM) batchInsertData(batchData interface{}, insertType string) (int64, error) {

	//反射获取切片的值
	getValue := reflect.ValueOf(batchData)

	//切片大小
	l := getValue.Len()

	//字段名
	var fieldName []string

	//占位符组
	var placeholderString []string


	for i := 0; i < l; i++ {
		value := getValue.Index(i)
		typed := value.Type() 
		if typed.Kind() != reflect.Struct {
			panic("批量插入的子元素必须是结构体类型")
		}

		num := value.NumField()

		//占位符
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
					//字段名只需记录一遍
					if i == 1 {
						fieldName = append(fieldName, strings.Split(sqlTag, ",")[0])
					}
					placeholder = append(placeholder, "?")
				}
			} else {
				//字段名只需记录一遍
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

	//拼接插入类型，表，字段名，占位符
	e.Prepare = insertType + " into " + e.GetTable() + " (" + strings.Join(fieldName, ",") + ") values " + strings.Join(placeholderString, ",")

	//prepare
	var stmt *sql.Stmt
	var err error
	stmt, err = e.Db.Prepare(e.Prepare)
	if err != nil {
		return 0, e.setErrorInfo(err)
	}

	//执行exec
	result, err := stmt.Exec(e.AllExec...)
	if err != nil {
		return 0, e.setErrorInfo(err)
	}

	//获取自增ID
	id, _ := result.LastInsertId()
	return id, nil
}

// InsertData 函数用于插入数据
// 输入：
//       data - 要插入的数据，必须是结构体
//       insertType - 插入类型，"insert"为插入，"replace"为替换
// 返回：
//       int64 - 插入的最后一条数据的自增ID
//       error - 错误信息，如果没有错误则为nil
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

// setErrorInfo 函数用于设置错误信息
// 输入：
//       err - 错误信息
// 返回：
//       error - 错误信息
func (e *SimpleORM) setErrorInfo(err error) error {
	// _, file, line, _ := runtime.Caller(1)
	// return errors.New("File: " + file + ":" + strconv.Itoa(line) + ", " + err.Error())
	return err
}


 // Insert 函数用于插入数据
// 输入：
//       data - 要插入的数据，必须是结构体或结构体切片或数组
// 返回：
//       int64 - 插入的最后一条数据的自增ID
//       error - 错误信息，如果没有错误则为nil
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


 // Replace 函数用于替换插入数据
// 输入：
//       data - 要替换插入的数据，必须是结构体或结构体切片或数组
// 返回：
//       int64 - 替换插入的最后一条数据的自增ID
//       error - 错误信息，如果没有错误则为nil
func (e *SimpleORM) Replace(data interface{}) (int64, error) {
	//判断是批量还是单个插入
	getValue := reflect.ValueOf(data).Kind()
	if getValue == reflect.Struct {
		return e.insertData(data, "replace")
	} else if getValue == reflect.Slice || getValue == reflect.Array {
		return e.batchInsertData(data, "replace")
	} else {
		return 0, errors.New("插入的数据格式不正确，单个插入格式为: struct,批量插入格式为: []struct")
	}
}


 // Where 函数用于设置where条件，支持多次调用，多次调用之间是and关系
// 输入：
//       data - 要替换插入的数据，必须是1.传入结构体2.传入键,值3.传入键,操作符,值
// 返回：
//       *SimpleORM - 设置where条件后的SimpleORM
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

// OrWhere 函数用于设置or where条件，支持多次调用，多次调用之间是or关系
// 输入：
//       data - 要替换插入的数据，必须是1.传入结构体2.传入键,值3.传入键,操作符,值
// 返回：
//       *SimpleORM - 设置or where条件后的SimpleORM
func (e *SimpleORM) OrWhere(data ...interface{}) *SimpleORM {
	
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
	if e.OrWhereParam != "" {
		e.OrWhereParam += " or ("
	} else {
		e.OrWhereParam += "("
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
		e.OrWhereParam += strings.Join(fieldNameArray, " and ") + ") "

	} else if dataType == 2 {
		//直接=的情况
		e.OrWhereParam += data[0].(string) + "=?) "
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
			e.OrWhereParam += data[0].(string) + " " + data2 + " (" + strings.Join(ps, ",") + ")) "

		} else {
			e.OrWhereParam += data[0].(string) + " " + data[1].(string) + " ?) "
			e.WhereExec = append(e.WhereExec, data[2])
		}
	}

	return e
}


 // Delete 函数用于删除数据(需先调用where方法，否则会删除全表)
// 返回：
//       int64 - 删除的行数
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


 // Update 函数用于更新数据(需先调用where方法，否则会更新全表)
// 输入：
//       data - 要更新的数据，必须是1.传入结构体2.传入键,值
// 返回：
//       int64 - 更新的行数
//       error - 错误信息，如果没有错误则为nil
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

// Select 函数用于查询数据(可先调用where方法或Field方法)
// 返回：
//       []map[string]string - 查询的结果，每一行是一个map，map的key是字段名，map的value是字段值
//       error - 错误信息，如果没有错误则为nil
func (e *SimpleORM) Select() ([]map[string]string, error) {
 
 
  //拼接
  e.Prepare = "select "+e.FieldParam +" from " + e.GetTable()
 
 
  //如果where不为空
  if e.WhereParam != "" || e.OrWhereParam != "" {
    e.Prepare += " where " + e.WhereParam + e.OrWhereParam
  }
 
 
  e.AllExec = e.WhereExec
 
  //query
  rows, err := e.Db.Query(e.Prepare, e.AllExec...)
  if err != nil {
    return nil, e.setErrorInfo(err)
  }
 
 
  //读出查询出的列字段名
  column, err := rows.Columns()
  if err != nil {
    return nil, e.setErrorInfo(err)
  }
 
 
  //values是单行的每列的值，切片第一层表示各列，第二层表示数据
  values := make([][]byte, len(column))
 
 
  //用len(column)作为当次查询的长度
  scans := make([]interface{}, len(column))

 
  for i := range values {
    scans[i] = &values[i]
  }
 
 
  results := make([]map[string]string, 0)
  for rows.Next() {
    if err := rows.Scan(scans...); err != nil {
      //query.Scan查询出来一行的数据都放在values里
      return nil, e.setErrorInfo(err)
    }
 
 
    //每行数据
    row := make(map[string]string) 
 
 
    //利用values数据生成每行的字典化的数据
    for k, v := range values {
      key := column[k]
      row[key] = string(v)
    }
 
 
    //添加到map切片中
    results = append(results, row)
  }
 
  return results, nil
}

// SelectOne 函数用于查询单个数据(可先调用where方法或Field方法)
// 返回：
//       map[string]string - 查询的结果，map的key是字段名，map的value是字段值
//       error - 错误信息，如果没有错误则为nil
func (e *SimpleORM) SelectOne() (map[string]string, error) {
  
  //limit 1 单个查询
  results, err := e.Limit(1).Select()
  if err != nil {
    return nil, e.setErrorInfo(err)
  }
 
 
  //判断是否为空
  if len(results) == 0 {
    return nil, nil
  } else {
    return results[0], nil
  }
}

// Field 函数用于设置查询字段
// 输入：
//       fields - 要查询的字段名，多个字段名用逗号分隔
// 返回：
//       *SimpleORM - 设置查询字段后的SimpleORM
func (e *SimpleORM)Field(fields ...string) *SimpleORM {
  e.FieldParam = strings.Join(fields, ",")
  return e
}


//Limit 函数用于设置limit条件,以供分页
// 输入：
//       limit - limit条件，1个参数为limit的值，2个参数为offset,limit的值
// 返回：
//       *SimpleORM - 设置limit条件后的SimpleORM
func (e *SimpleORM) Limit(limit ...int64) *SimpleORM {
  if len(limit) == 1 {
    e.LimitParam = strconv.Itoa(int(limit[0]))
  } else if len(limit) == 2 {
    e.LimitParam = strconv.Itoa(int(limit[0])) + "," + strconv.Itoa(int(limit[1]))
  } else {
    panic("参数个数错误")
  }
  return e
}

