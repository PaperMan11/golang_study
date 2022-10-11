package repository

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

func InitDB() {
	host := viper.GetString("mysql.host")
	port := viper.GetString("mysql.port")
	database := viper.GetString("mysql.database")
	username := viper.GetString("mysql.username")
	password := viper.GetString("mysql.password")
	charset := viper.GetString("mysql.charset")
	// dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	// dsn := strings.Join([]string{username, ":", password, "@tpc(", host, ":", port, ")/", database, "?charset=" + charset + "&parseTime=true"}, "")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local", username, password, host, port, database, charset)
	if err := Database(dsn); err != nil {
		panic(err)
	}
}

func Database(dsn string) error {
	var ormLogger logger.Interface
	if gin.Mode() == "debug" {
		ormLogger = logger.Default.LogMode(logger.Info)
	} else {
		ormLogger = logger.Default
	}
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,  //禁用datatime的精度，mysql 5.6之前不支持
		DontSupportRenameIndex:    true,  // 重命名索引的时候采用删除并新建的方式。mysql 5.7 之前不支持重命名
		DontSupportRenameColumn:   true,  // 用change重命名列，mysql 8 之前的数据库不支持重命名列
		SkipInitializeWithVersion: false, // 根据版本自动配置
	}), &gorm.Config{
		Logger: ormLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 单数命名table
		},
	})
	if err != nil {
		return err
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(20)  // 设置空闲连接池
	sqlDB.SetMaxOpenConns(100) // 最大连接数
	sqlDB.SetConnMaxLifetime(time.Second * 30)
	DB = db
	migration()
	return err
}
