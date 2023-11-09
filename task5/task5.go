package main
import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type User1 struct {
    Username   string `sql:"username"`
    Departname string `sql:"departname"`
    Status     int64  `sql:"status"`
}
 
 
user2 := User1{
    Username:   "EE",
    Departname: "22", 
    Status:     1,
}
 
 
type SmallormEngine struct {
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
   TransStatus  int
   Tx           *sql.Tx
   GroupParam   string
   HavingParam  string
}
//新建Mysql连接
func NewMysql(Username string, Password string, Address string, Dbname string) (*SmallormEngine, error) {
    dsn := Username + ":" + Password + "@tcp(" + Address + ")/" + Dbname + "?charset=utf8&timeout=5s&readTimeout=6s"
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }
 
 
    //最大连接数等配置，先占个位
   //db.SetMaxOpenConns(3)
   //db.SetMaxIdleConns(3)
 
 
    return &SmallormEngine{
        Db:         db,
        FieldParam: "*",
    }, nil
}
//设置表名
func (e *SmallormEngine) Table(name string) *SmallormEngine {
   e.TableName = name
 
 
   //重置引擎
   e.resetSmallormEngine()
   return e
}
 
 
//获取表名
func (e *SmallormEngine) GetTable() string {
   return e.TableName
}

//批量插入
func (e *SmallormEngine) BatchInsert(data interface{}) (int64, error) {
    return e.batchInsertData(data, "insert")
}
 
 
//批量替换插入
func (e *SmallormEngine) BatchReplace(data interface{}) (int64, error) {
    return e.batchInsertData(data, "replace")
}
  
 
//批量插入
func (e *SmallormEngine) batchInsertData(batchData interface{}, insertType string) (int64, error) {
 
 
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
 
 
 
 
//自定义错误格式
func (e *SmallormEngine) setErrorInfo(err error) error {
  _, file, line, _ := runtime.Caller(1)
  return errors.New("File: " + file + ":" + strconv.Itoa(line) + ", " + err.Error())
}