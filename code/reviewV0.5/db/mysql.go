package db

import (
	"fmt"
	"review/config"
	"review/dal/query"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DBEngine *gorm.DB

func NewMySQL(mysqlCfg *config.MysqlSetting) (*gorm.DB, error) {
	dsn := fmt.Sprintf(`%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true&loc=Local`,
		mysqlCfg.UserName,
		mysqlCfg.Password,
		mysqlCfg.Host,
		mysqlCfg.DbName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(mysqlCfg.MaxOpenConns) //设置数据库连接池最大连接数
	sqlDB.SetMaxIdleConns(mysqlCfg.MaxIdleConns) //连接池最大允许的空闲连接数，如果没有sql任务需要执行的连接数大于MaxIdleConns，超过的连接会被连接池关闭

	query.SetDefault(db) //设置了才能使用query包
	return db, nil
}
