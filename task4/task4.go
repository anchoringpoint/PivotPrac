package main
import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type route struct {
	id   int
	route string
	origin string
	destination string
}
type alias struct {
	id   int
	name string
	alias string
}
// 定义一个全局对象db
var db *sql.DB

func main() {
	err := initDB() // 调用输出化数据库的函数
	if err != nil {
		fmt.Printf("init db failed,err:%v\n", err)
		return
	}
	fmt.Println("连接数据库成功！")

}

func route_insertRow(route string,origin string,destination string){
	//插入数据
	sqlStr := "insert into route(route,origin,destination) values (?,?,?)"
	ret, err := db.Exec(sqlStr, route,origin,destination)
	if err != nil {
		fmt.Printf("insert failed, err:%v\n", err)
		return
	}
	_,err = ret.LastInsertId() // 新插入数据的id
	if err != nil {
		fmt.Printf("get lastinsert ID failed, err:%v\n", err)
		return
	}
}
func route_queryRow(route_name string)route{
	sqlStr := "select id,route,origin,destination from route where route = ?"
	var r route

	err := db.QueryRow(sqlStr, route_name).Scan(&r.id, &r.route, &r.origin, &r.destination)
	if err != nil {
		fmt.Printf("scan failed, err:%v\n", err)
		return r
	}
	return r
}
func route_updateRow(route_name string,origin string,destination string){
	sqlStr := "update route set origin = ? ,destination = ? where route = ?"
	ret, err := db.Exec(sqlStr, origin,destination,route_name)
	if err != nil {
		fmt.Printf("update failed, err:%v\n", err)
		return
	}
	_, err = ret.RowsAffected() // 操作影响的行数
	if err != nil {
		fmt.Printf("get RowsAffected failed, err:%v\n", err)
		return
	}
}
func route_deleteRow(route_name string){
	sqlStr := "delete from route where route = ?"
	ret, err := db.Exec(sqlStr, route_name)
	if err != nil {
		fmt.Printf("delete failed, err:%v\n", err)
		return
	}
	_, err = ret.RowsAffected() // 操作影响的行数
	if err != nil {
		fmt.Printf("get RowsAffected failed, err:%v\n", err)
		return
	}
}
func alias_insertRow(alias string,name string){
	//插入数据
	sqlStr := "insert into alias(alias,name) values (?,?)"
	ret, err := db.Exec(sqlStr, alias,name)
	if err != nil {
		fmt.Printf("insert failed, err:%v\n", err)
		return
	}
	_,err = ret.LastInsertId() // 新插入数据的id
	if err != nil {
		fmt.Printf("get lastinsert ID failed, err:%v\n", err)
		return
	}
}
func alias_queryRow(name string)alias{
	sqlStr := "select id,name,alias from alias where alias = ?"
	var a alias

	err := db.QueryRow(sqlStr, name).Scan(&a.id, &a.name, &a.alias)
	if err != nil {
		a.id=-1
		return a
	}
	return a
}
func alias_updateRow(alias string,name string){
	sqlStr := "update alias set name = ? where alias = ?"
	ret, err := db.Exec(sqlStr, name,alias)
	if err != nil {
		fmt.Printf("update failed, err:%v\n", err)
		return
	}
	_, err = ret.RowsAffected() // 操作影响的行数
	if err != nil {
		fmt.Printf("get RowsAffected failed, err:%v\n", err)
		return
	}
}
func alias_deleteRow(alias string){
	sqlStr := "delete from alias where alias = ?"
	ret, err := db.Exec(sqlStr, alias)
	if err != nil {
		fmt.Printf("delete failed, err:%v\n", err)
		return
	}
	_, err = ret.RowsAffected() // 操作影响的行数
	if err != nil {
		fmt.Printf("get RowsAffected failed, err:%v\n", err)
		return
	}
}

// 定义一个初始化数据库的函数
func initDB() (err error) {
	// DSN:Data Source Name
	dsn := "root:root@tcp(localhost:3306)/record?charset=utf8mb4&parseTime=True"

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	// 尝试与数据库建立连接（校验dsn是否正确）
	err = db.Ping()
	if err != nil {
		return err
	}
	return nil
}

