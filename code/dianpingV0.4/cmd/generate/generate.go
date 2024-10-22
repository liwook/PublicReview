package main

import (
	"dianping/internal/config"
	"dianping/internal/db"

	"github.com/spf13/pflag"
	"gorm.io/gen"
)

var options config.MysqlSetting

// func init() {
// 	err := config.ReadConfigFile()
// 	if err != nil {
// 		panic(err)
// 	}

//		err = config.ReadSection("mysql", &options)
//		if err != nil {
//			panic(err)
//		}
//	}
func initConfig(path string) {
	err := config.ReadConfigFile(path)
	if err != nil {
		panic(err)
	}

	err = config.ReadSection("mysql", &options)
	if err != nil {
		panic(err)
	}
}

// 这里使用的是gorm.io/gen@v0.3.16，最新版本的某个函数不会使用，gen.FieldGORMTag不知该如何使用。是为了可以自动更新创建时间和更改时间
func main() {
	configPath := pflag.StringP("config", "c", "../configs/config.yaml", "config file path")
	pflag.Parse()
	initConfig(*configPath)

	db, err := db.NewMySQL(&options)
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
	dataMap := map[string]func(detailType string) (dataType string){
		"tinyint":  func(detailType string) (dataType string) { return "int8" },
		"smallint": func(detailType string) (dataType string) { return "int16" },
		"bigint":   func(detailType string) (dataType string) { return "int64" },
		"int":      func(detailType string) (dataType string) { return "int64" },
	}

	g.WithDataTypeMap(dataMap)

	autoUpdateTimeField := gen.FieldGORMTag("modified_on", "column:modified_on;type:int unsigned;autoUpdateTime")
	autoCreateTimeField := gen.FieldGORMTag("created_on", "column:created_on;type:int unsigned;autoCreateTime")
	// softDeleteField := gen.FieldType("deleted_on", "soft_delete.DeletedAt")
	// 模型自定义选项组
	fieldOpts := []gen.ModelOpt{autoCreateTimeField, autoUpdateTimeField}

	allModel := g.GenerateAllTable(fieldOpts...)
	g.ApplyBasic(allModel...)
	g.Execute()
}
