package db

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"os"
	"time"

	"database/sql"
	"gorm.io/gorm"
)

// 数据库实例
var (
	GormDB *gorm.DB
)

// Init 初始化
func init() {
	//从配置文件读取Mysql配置信息
	//var MysqlConnectData = config.Conf.Mysql
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second,   // 慢 SQL 阈值
			LogLevel:      logger.Silent, // Log level
			Colorful:      false,         // 禁用彩色打印
		},
	)
	var mysqlDSN="%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true&loc=Local"
	var DBUser="root"
	var DBPwd="root"
	var DBHost="127.0.0.1"
	var DBPort=3306
	var DBName="demo"
	//dbConfig := fmt.Sprintf(MysqlConnectData.MysqlDSN, MysqlConnectData.DBUser, MysqlConnectData.DBPwd, MysqlConnectData.DBHost, MysqlConnectData.DBPort, MysqlConnectData.DBName)
	dbConfig := fmt.Sprintf(mysqlDSN, DBUser, DBPwd,DBHost,DBPort, DBName)
	sqlDB, _ := sql.Open("mysql", dbConfig)
	GormDB, _ = gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: newLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名，启用该选项后，`User` 表将是`user`
		},
	})
	GormDB = GormDB.Debug()
	DB, _ := GormDB.DB()
	DB.SetMaxIdleConns(10)                   //最大空闲连接数
	DB.SetMaxOpenConns(30)                   //最大连接数
	DB.SetConnMaxLifetime(time.Second * 300) //设置连接空闲超时
}
