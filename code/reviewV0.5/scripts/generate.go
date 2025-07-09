package main

import (
	"review/config"
	"review/db"

	"github.com/spf13/pflag"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func initDb() (*gorm.DB, error) {
	configPath := pflag.StringP("config", "c", "configs/config.yaml", "config file path")
	pflag.Parse()
	err := config.ReadConfigFile(*configPath)
	if err != nil {
		panic(err)
	}

	var options config.MysqlSetting
	err = config.ReadSection("mysql", &options)
	if err != nil {
		panic(err)
	}

	return db.NewMySQL(&options)
}

func main() {
	db, err := initDb()
	if err != nil {
		panic(err)
	}

	g := gen.NewGenerator(gen.Config{
		// 相对执行`go run`时的路径, 会自动创建目录
		OutPath: "./dal/query",
		// ModelPkgPath:      "./dal/model",不写是最好，不然就出现目录：dal/dal/model
		Mode:              gen.WithDefaultQuery | gen.WithoutContext | gen.WithQueryInterface,
		FieldNullable:     false,
		FieldCoverable:    false,
		FieldSignable:     true,
		FieldWithIndexTag: false,
		FieldWithTypeTag:  true,
	})
	g.UseDB(db)

	allModel := g.GenerateAllTable()

	g.ApplyBasic(allModel...)
	g.Execute()
}
