package geeorm

import (
	"GeeORM/dialect"
	"GeeORM/log"
	"GeeORM/session"
	"database/sql"
)

// 用户交互

type Engine struct {
	db      *sql.DB
	dialect dialect.Dialect
}

// @driver: 数据库驱动
// @source: 数据库名称
func NewEngine(driver, source string) (e *Engine, err error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		log.Error(err)
		return
	}
	// 确保数据库连接正常
	if err = db.Ping(); err != nil {
		log.Error(err)
		return
	}
	// 确保方言(dialect)存在
	dial, ok := dialect.GetDialect(driver)
	if !ok {
		log.Errorf("dialect %s Not Found", driver)
		return
	}
	e = &Engine{db: db, dialect: dial}
	log.Info("Connect database success!")
	return
}

func (engine *Engine) Close() {
	if err := engine.db.Close(); err != nil {
		log.Error("Failed to close database!")
	}
	log.Info("Close database success!")
}

func (engine *Engine) NewSession() *session.Session {
	return session.New(engine.db, engine.dialect)
}

type TxFunc func(*session.Session) (result interface{}, err error)

// 事务一键式使用的接口
func (engine *Engine) Transaction(f TxFunc) (result interface{}, err error) {
	s := engine.NewSession()
	if err := s.Begin(); err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = s.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			_ = s.Rollback() // err is non-nil; don't change it
		} else {
			defer func() {
				if err != nil {
					_ = s.Rollback()
				}
			}()
			err = s.Commit() // err is nil; if Commit returns error update err
		}
	}()
	return f(s)
}
