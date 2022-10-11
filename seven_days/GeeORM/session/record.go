package session

import (
	"GeeORM/clause"
	"errors"
	"reflect"
)

// 1）多次调用 clause.Set() 构造好每一个子句。
// 2）调用一次 clause.Build() 按照传入的顺序构造出最终的 SQL 语句。
// 构造完成后，调用 Raw().Exec() 方法执行。

// Insert 需要将已经存在的对象的每一个字段的值平铺开来 struct -> []interface{}
func (s *Session) Insert(values ...interface{}) (int64, error) {
	recordValues := make([]interface{}, 0)
	for _, value := range values {
		s.CallMethod(BeforeInsert, value) // 每行都操作
		table := s.Model(value).RefTable()
		s.clause.Set(clause.INSERT, table.Name, table.FieldNames)
		recordValues = append(recordValues, table.RecordValues(value))
	}

	s.clause.Set(clause.VALUES, recordValues...)
	sql, vars := s.clause.Build(clause.INSERT, clause.VALUES)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	s.CallMethod(AfterInsert, nil)
	return result.RowsAffected()
}

// Find 功能的难点和 Insert 恰好反了过来。
// Insert 需要将已经存在的对象的每一个字段的值平铺开来，
// 而 Find 则是需要根据平铺开的字段的值构造出对象。
// 同样，也需要用到反射(reflect)。
func (s *Session) Find(values interface{}) error {
	s.CallMethod(BeforeQuery, nil)
	// 1）destSlice.Type().Elem() 获取切片的单个元素的类型 destType，
	// 	使用 reflect.New() 方法创建一个 destType 的实例，作为 Model() 的入参，映射出表结构 RefTable()。
	destSlice := reflect.Indirect(reflect.ValueOf(values)) // []
	destType := destSlice.Type().Elem()                    // main.User
	table := s.Model(reflect.New(destType).Elem().Interface()).RefTable()
	// 2）根据表结构，使用 clause 构造出 SELECT 语句，查询到所有符合条件的记录 rows。
	s.clause.Set(clause.SELECT, table.Name, table.FieldNames)
	sql, vars := s.clause.Build(clause.SELECT, clause.WHERE, clause.ORDERBY, clause.LIMIT)
	rows, err := s.Raw(sql, vars...).QueryRows()
	if err != nil {
		return err
	}
	// 3）遍历每一行记录，利用反射创建 destType 的实例 dest，将 dest 的所有字段平铺开，构造切片 values1。
	for rows.Next() {
		dest := reflect.New(destType).Elem()
		var values1 []interface{}
		for _, name := range table.FieldNames {
			values1 = append(values1, dest.FieldByName(name).Addr().Interface())
		}
		// 4）调用 rows.Scan() 将该行记录每一列的值依次赋值给 values1 中的每一个字段。
		if err := rows.Scan(values1...); err != nil {
			return err
		}
		s.CallMethod(AfterQuery, dest.Addr().Interface()) // AfterQuery 钩子可以操作每一行记录
		// 5）将 dest 添加到切片 destSlice 中。循环直到所有的记录都添加到切片 destSlice 中。
		destSlice.Set(reflect.Append(destSlice, dest))
	}
	return rows.Close()
}

// support map[string]interface{}
// also support kv list: "Name", "tom", "Age", 18, ...
func (s *Session) Update(kv ...interface{}) (int64, error) {
	s.CallMethod(BeforeUpdate, nil)
	m, ok := kv[0].(map[string]interface{})
	if !ok {
		m = make(map[string]interface{})
		for i := 0; i < len(kv); i += 2 {
			m[kv[i].(string)] = kv[i+1]
		}
	}
	s.clause.Set(clause.UPDATE, s.RefTable().Name, m)
	sql, vars := s.clause.Build(clause.UPDATE, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	s.CallMethod(AfterUpdate, nil)
	return result.RowsAffected()
}

func (s *Session) Delete() (int64, error) {
	s.CallMethod(BeforeDelete, nil)
	s.clause.Set(clause.DELETE, s.RefTable().Name)
	sql, vars := s.clause.Build(clause.DELETE, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	s.CallMethod(AfterDelete, nil)
	return result.RowsAffected()
}

func (s *Session) Count() (int64, error) {
	s.clause.Set(clause.COUNT, s.RefTable().Name)
	sql, vars := s.clause.Build(clause.COUNT, clause.WHERE)
	row := s.Raw(sql, vars...).QueryRow()
	var tmp int64
	if err := row.Scan(&tmp); err != nil {
		return 0, err
	}
	return tmp, nil
}

// First 结合链式调用返回一条信息
// eg: 	u := &User{}
// 		_ = s.OrderBy("Age DESC").First(u)
func (s *Session) First(value interface{}) error {
	dest := reflect.Indirect(reflect.ValueOf(value))
	destSlice := reflect.New(reflect.SliceOf(dest.Type())).Elem()
	if err := s.Limit(1).Find(destSlice.Addr().Interface()); err != nil {
		return err
	}
	if destSlice.Len() == 0 {
		return errors.New("NOT FOUND")
	}
	dest.Set(destSlice.Index(0))
	return nil
}

/*-------------------------------------chain链式调用-------------------------------------*/
// s := geeorm.NewEngine("sqlite3", "gee.db").NewSession()
// var users []User
// s.Where("Age > 18").Limit(3).Find(&users)
// WHERE、LIMIT、ORDER BY 等查询条件语句非常适合链式调用

func (s *Session) Limit(num int) *Session {
	s.clause.Set(clause.LIMIT, num)
	return s
}

func (s *Session) Where(desc string, args ...interface{}) *Session {
	var vars []interface{}
	s.clause.Set(clause.WHERE, append(append(vars, desc), args...)...)
	return s
}

func (s *Session) OrderBy(desc string) *Session {
	s.clause.Set(clause.ORDERBY, desc)
	return s
}
