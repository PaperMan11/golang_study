package session

import (
	"GeeORM/clause"
	"GeeORM/dialect"
	"GeeORM/log"
	"GeeORM/schema"
	"database/sql"
	"strings"
)

type Session struct {
	db *sql.DB
	/*
		session struct 是会在会话中复用的，如果使用 string 类型，
		string 是只读不可变的，每次修改其实都要重新申请一个内存空间，
		都是一个新的 string，而 string.Builder 底层使用 []byte 实现。
	*/
	sql     strings.Builder // SQL 语句
	sqlVars []interface{}   // SQL 语句中占位符的对应值
	clause  clause.Clause
	/*
		当 tx 不为空时，则使用 tx 执行 SQL 语句，
		否则使用 db 执行 SQL 语句。这样既兼容了原有的执行方式，
		又提供了对事务的支持。
	*/
	tx *sql.Tx // 支持事务

	dialect  dialect.Dialect
	refTable *schema.Schema
}

// CommonDB db的最小函数集
type CommonDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

var _ CommonDB = (*sql.DB)(nil)
var _ CommonDB = (*sql.Tx)(nil)

func New(db *sql.DB, dialect dialect.Dialect) *Session {
	return &Session{
		db:      db,
		dialect: dialect,
	}
}

// Clear 清空变量，实现一次会话执行多次 sql 语句
func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
	s.clause = clause.Clause{}
}

func (s *Session) DB() CommonDB {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}

func (s *Session) Raw(sql string, values ...interface{}) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVars = append(s.sqlVars, values...)
	return s
}

// Exec 还原成原始的 sql 去执行
func (s *Session) Exec() (result sql.Result, err error) {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	if result, err = s.DB().Exec(s.sql.String(), s.sqlVars...); err != nil {
		log.Error(err)
	}
	return
}

// QueryRow 从 db 获取一条记录
func (s *Session) QueryRow() *sql.Row {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	return s.DB().QueryRow(s.sql.String(), s.sqlVars...)
}

// QueryRows 从 db 获取记录列表
func (s *Session) QueryRows() (rows *sql.Rows, err error) {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	if rows, err = s.DB().Query(s.sql.String(), s.sqlVars...); err != nil {
		log.Error(err)
	}
	return
}
