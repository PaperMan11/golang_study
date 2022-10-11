package schema

import (
	"GeeORM/dialect"
	"go/ast"
	"reflect"
)

/*
Dialect 实现了一些特定的 SQL 语句的转换，
接下来我们将要实现 ORM 框架中最为核心的转换——对象(object)和表(table)的转换。
给定一个任意的对象，转换为关系型数据库中的表结构。

在数据库中创建一张表需要哪些要素呢？
    - 表名(table name) —— 结构体名(struct name)
    - 字段名和字段类型 —— 成员变量和类型。
    - 额外的约束条件(例如非空、主键等) —— 成员变量的Tag（Go 语言通过 Tag 实现，Java、Python 等语言通过注解实现）
*/

// Field 表示数据库的一列
type Field struct {
	Name string // 字段名
	Type string // 数据类型（具体数据库的 非go语言的）
	Tag  string // 约束等
}

// Schema 表示数据库的一个表
type Schema struct {
	Model      interface{} // 被映射的对象
	Name       string
	Fields     []*Field          // eg. Name string primary key;...
	FieldNames []string          // 所有的列名
	fieldMap   map[string]*Field // 列名与Field之间的映射，后续无需遍历
}

func (schema *Schema) GetField(name string) *Field {
	return schema.fieldMap[name]
}

func Parse(dest interface{}, d dialect.Dialect) *Schema {
	// reflect.Indirect
	// 间接返回v所指向的值。如果v是一个空指针，间接返回一个零值。如果v不是指针，则Indirect返回v。
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	schema := &Schema{
		Model:    dest,
		Name:     modelType.Name(),
		fieldMap: make(map[string]*Field),
	}

	for i := 0; i < modelType.NumField(); i++ {
		p := modelType.Field(i)
		// p.Anonymous 是否为嵌入字段
		// isexports 用于报告名称是否以大写字母开头
		if !p.Anonymous && ast.IsExported(p.Name) {
			field := &Field{
				Name: p.Name,
				Type: d.DataTypeOf(reflect.Indirect(reflect.New(p.Type))),
			}
			if v, ok := p.Tag.Lookup("geeorm"); ok {
				field.Tag = v
			}
			schema.Fields = append(schema.Fields, field)
			schema.FieldNames = append(schema.FieldNames, p.Name)
			schema.fieldMap[p.Name] = field
		}
	}
	return schema
}

// u1 := &User{Name: "Tom", Age: 18}
// u2 := &User{Name: "Sam", Age: 25}
// RecordValues u1, u2 转换为("Tom", 18), ("Same", 25) 这样的格式
func (schema *Schema) RecordValues(dest interface{}) []interface{} {
	destValue := reflect.Indirect(reflect.ValueOf(dest))
	var fieldValues []interface{}
	for _, field := range schema.Fields {
		// FieldByName 返回带有给定名称的结构字段。如果没有找到字段，则返回零值。
		fieldValues = append(fieldValues, destValue.FieldByName(field.Name).Interface())
	}
	return fieldValues
}
