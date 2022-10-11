package dialect

import "reflect"

// SQL 语句中的类型和 Go 语言中的类型是不同的，
// 例如Go 语言中的 int、int8、int16 等类型均对应 SQLite 中的 integer 类型。
// 因此实现 ORM 映射的第一步，需要思考如何将 Go 语言的类型映射为数据库中的类型。

// 同时，不同数据库支持的数据类型也是有差异的，
// 即使功能相同，在 SQL 语句的表达上也可能有差异。
// ORM 框架往往需要兼容多种数据库，因此我们需要将差异的这一部分提取出来，
// 每一种数据库分别实现，实现最大程度的复用和解耦。这部分代码称之为 dialect。

var dialectsMap = map[string]Dialect{}

type Dialect interface {
	DataTypeOf(typ reflect.Value) string                    //用于将 Go 语言的类型转换为该数据库的数据类型。
	TableExistSQL(tableName string) (string, []interface{}) // 返回某个表是否存在的 SQL 语句，参数是表名(table)。
}

func RegisterDialect(name string, dialect Dialect) {
	dialectsMap[name] = dialect
}

func GetDialect(name string) (dialect Dialect, ok bool) {
	dialect, ok = dialectsMap[name]
	return
}
