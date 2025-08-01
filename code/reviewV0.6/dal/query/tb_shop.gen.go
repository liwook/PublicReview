// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package query

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"gorm.io/plugin/dbresolver"

	"review/dal/model"
)

func newTbShop(db *gorm.DB, opts ...gen.DOOption) tbShop {
	_tbShop := tbShop{}

	_tbShop.tbShopDo.UseDB(db, opts...)
	_tbShop.tbShopDo.UseModel(&model.TbShop{})

	tableName := _tbShop.tbShopDo.TableName()
	_tbShop.ALL = field.NewAsterisk(tableName)
	_tbShop.ID = field.NewUint64(tableName, "id")
	_tbShop.Name = field.NewString(tableName, "name")
	_tbShop.TypeID = field.NewUint64(tableName, "type_id")
	_tbShop.Images = field.NewString(tableName, "images")
	_tbShop.Area = field.NewString(tableName, "area")
	_tbShop.Address = field.NewString(tableName, "address")
	_tbShop.X = field.NewFloat64(tableName, "x")
	_tbShop.Y = field.NewFloat64(tableName, "y")
	_tbShop.AvgPrice = field.NewUint64(tableName, "avg_price")
	_tbShop.Sold = field.NewUint32(tableName, "sold")
	_tbShop.Comments = field.NewUint32(tableName, "comments")
	_tbShop.Score = field.NewUint32(tableName, "score")
	_tbShop.OpenHours = field.NewString(tableName, "open_hours")
	_tbShop.CreateTime = field.NewTime(tableName, "create_time")
	_tbShop.UpdateTime = field.NewTime(tableName, "update_time")

	_tbShop.fillFieldMap()

	return _tbShop
}

type tbShop struct {
	tbShopDo

	ALL        field.Asterisk
	ID         field.Uint64  // 主键
	Name       field.String  // 商铺名称
	TypeID     field.Uint64  // 商铺类型的id
	Images     field.String  // 商铺图片，多个图片以','隔开
	Area       field.String  // 商圈，例如陆家嘴
	Address    field.String  // 地址
	X          field.Float64 // 经度
	Y          field.Float64 // 维度
	AvgPrice   field.Uint64  // 均价，取整数
	Sold       field.Uint32  // 销量
	Comments   field.Uint32  // 评论数量
	Score      field.Uint32  // 评分，1~5分，乘10保存，避免小数
	OpenHours  field.String  // 营业时间，例如 10:00-22:00
	CreateTime field.Time    // 创建时间
	UpdateTime field.Time    // 更新时间

	fieldMap map[string]field.Expr
}

func (t tbShop) Table(newTableName string) *tbShop {
	t.tbShopDo.UseTable(newTableName)
	return t.updateTableName(newTableName)
}

func (t tbShop) As(alias string) *tbShop {
	t.tbShopDo.DO = *(t.tbShopDo.As(alias).(*gen.DO))
	return t.updateTableName(alias)
}

func (t *tbShop) updateTableName(table string) *tbShop {
	t.ALL = field.NewAsterisk(table)
	t.ID = field.NewUint64(table, "id")
	t.Name = field.NewString(table, "name")
	t.TypeID = field.NewUint64(table, "type_id")
	t.Images = field.NewString(table, "images")
	t.Area = field.NewString(table, "area")
	t.Address = field.NewString(table, "address")
	t.X = field.NewFloat64(table, "x")
	t.Y = field.NewFloat64(table, "y")
	t.AvgPrice = field.NewUint64(table, "avg_price")
	t.Sold = field.NewUint32(table, "sold")
	t.Comments = field.NewUint32(table, "comments")
	t.Score = field.NewUint32(table, "score")
	t.OpenHours = field.NewString(table, "open_hours")
	t.CreateTime = field.NewTime(table, "create_time")
	t.UpdateTime = field.NewTime(table, "update_time")

	t.fillFieldMap()

	return t
}

func (t *tbShop) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := t.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (t *tbShop) fillFieldMap() {
	t.fieldMap = make(map[string]field.Expr, 15)
	t.fieldMap["id"] = t.ID
	t.fieldMap["name"] = t.Name
	t.fieldMap["type_id"] = t.TypeID
	t.fieldMap["images"] = t.Images
	t.fieldMap["area"] = t.Area
	t.fieldMap["address"] = t.Address
	t.fieldMap["x"] = t.X
	t.fieldMap["y"] = t.Y
	t.fieldMap["avg_price"] = t.AvgPrice
	t.fieldMap["sold"] = t.Sold
	t.fieldMap["comments"] = t.Comments
	t.fieldMap["score"] = t.Score
	t.fieldMap["open_hours"] = t.OpenHours
	t.fieldMap["create_time"] = t.CreateTime
	t.fieldMap["update_time"] = t.UpdateTime
}

func (t tbShop) clone(db *gorm.DB) tbShop {
	t.tbShopDo.ReplaceConnPool(db.Statement.ConnPool)
	return t
}

func (t tbShop) replaceDB(db *gorm.DB) tbShop {
	t.tbShopDo.ReplaceDB(db)
	return t
}

type tbShopDo struct{ gen.DO }

type ITbShopDo interface {
	gen.SubQuery
	Debug() ITbShopDo
	WithContext(ctx context.Context) ITbShopDo
	WithResult(fc func(tx gen.Dao)) gen.ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() ITbShopDo
	WriteDB() ITbShopDo
	As(alias string) gen.Dao
	Session(config *gorm.Session) ITbShopDo
	Columns(cols ...field.Expr) gen.Columns
	Clauses(conds ...clause.Expression) ITbShopDo
	Not(conds ...gen.Condition) ITbShopDo
	Or(conds ...gen.Condition) ITbShopDo
	Select(conds ...field.Expr) ITbShopDo
	Where(conds ...gen.Condition) ITbShopDo
	Order(conds ...field.Expr) ITbShopDo
	Distinct(cols ...field.Expr) ITbShopDo
	Omit(cols ...field.Expr) ITbShopDo
	Join(table schema.Tabler, on ...field.Expr) ITbShopDo
	LeftJoin(table schema.Tabler, on ...field.Expr) ITbShopDo
	RightJoin(table schema.Tabler, on ...field.Expr) ITbShopDo
	Group(cols ...field.Expr) ITbShopDo
	Having(conds ...gen.Condition) ITbShopDo
	Limit(limit int) ITbShopDo
	Offset(offset int) ITbShopDo
	Count() (count int64, err error)
	Scopes(funcs ...func(gen.Dao) gen.Dao) ITbShopDo
	Unscoped() ITbShopDo
	Create(values ...*model.TbShop) error
	CreateInBatches(values []*model.TbShop, batchSize int) error
	Save(values ...*model.TbShop) error
	First() (*model.TbShop, error)
	Take() (*model.TbShop, error)
	Last() (*model.TbShop, error)
	Find() ([]*model.TbShop, error)
	FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.TbShop, err error)
	FindInBatches(result *[]*model.TbShop, batchSize int, fc func(tx gen.Dao, batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*model.TbShop) (info gen.ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	Updates(value interface{}) (info gen.ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	UpdateColumns(value interface{}) (info gen.ResultInfo, err error)
	UpdateFrom(q gen.SubQuery) gen.Dao
	Attrs(attrs ...field.AssignExpr) ITbShopDo
	Assign(attrs ...field.AssignExpr) ITbShopDo
	Joins(fields ...field.RelationField) ITbShopDo
	Preload(fields ...field.RelationField) ITbShopDo
	FirstOrInit() (*model.TbShop, error)
	FirstOrCreate() (*model.TbShop, error)
	FindByPage(offset int, limit int) (result []*model.TbShop, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Rows() (*sql.Rows, error)
	Row() *sql.Row
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) ITbShopDo
	UnderlyingDB() *gorm.DB
	schema.Tabler
}

func (t tbShopDo) Debug() ITbShopDo {
	return t.withDO(t.DO.Debug())
}

func (t tbShopDo) WithContext(ctx context.Context) ITbShopDo {
	return t.withDO(t.DO.WithContext(ctx))
}

func (t tbShopDo) ReadDB() ITbShopDo {
	return t.Clauses(dbresolver.Read)
}

func (t tbShopDo) WriteDB() ITbShopDo {
	return t.Clauses(dbresolver.Write)
}

func (t tbShopDo) Session(config *gorm.Session) ITbShopDo {
	return t.withDO(t.DO.Session(config))
}

func (t tbShopDo) Clauses(conds ...clause.Expression) ITbShopDo {
	return t.withDO(t.DO.Clauses(conds...))
}

func (t tbShopDo) Returning(value interface{}, columns ...string) ITbShopDo {
	return t.withDO(t.DO.Returning(value, columns...))
}

func (t tbShopDo) Not(conds ...gen.Condition) ITbShopDo {
	return t.withDO(t.DO.Not(conds...))
}

func (t tbShopDo) Or(conds ...gen.Condition) ITbShopDo {
	return t.withDO(t.DO.Or(conds...))
}

func (t tbShopDo) Select(conds ...field.Expr) ITbShopDo {
	return t.withDO(t.DO.Select(conds...))
}

func (t tbShopDo) Where(conds ...gen.Condition) ITbShopDo {
	return t.withDO(t.DO.Where(conds...))
}

func (t tbShopDo) Order(conds ...field.Expr) ITbShopDo {
	return t.withDO(t.DO.Order(conds...))
}

func (t tbShopDo) Distinct(cols ...field.Expr) ITbShopDo {
	return t.withDO(t.DO.Distinct(cols...))
}

func (t tbShopDo) Omit(cols ...field.Expr) ITbShopDo {
	return t.withDO(t.DO.Omit(cols...))
}

func (t tbShopDo) Join(table schema.Tabler, on ...field.Expr) ITbShopDo {
	return t.withDO(t.DO.Join(table, on...))
}

func (t tbShopDo) LeftJoin(table schema.Tabler, on ...field.Expr) ITbShopDo {
	return t.withDO(t.DO.LeftJoin(table, on...))
}

func (t tbShopDo) RightJoin(table schema.Tabler, on ...field.Expr) ITbShopDo {
	return t.withDO(t.DO.RightJoin(table, on...))
}

func (t tbShopDo) Group(cols ...field.Expr) ITbShopDo {
	return t.withDO(t.DO.Group(cols...))
}

func (t tbShopDo) Having(conds ...gen.Condition) ITbShopDo {
	return t.withDO(t.DO.Having(conds...))
}

func (t tbShopDo) Limit(limit int) ITbShopDo {
	return t.withDO(t.DO.Limit(limit))
}

func (t tbShopDo) Offset(offset int) ITbShopDo {
	return t.withDO(t.DO.Offset(offset))
}

func (t tbShopDo) Scopes(funcs ...func(gen.Dao) gen.Dao) ITbShopDo {
	return t.withDO(t.DO.Scopes(funcs...))
}

func (t tbShopDo) Unscoped() ITbShopDo {
	return t.withDO(t.DO.Unscoped())
}

func (t tbShopDo) Create(values ...*model.TbShop) error {
	if len(values) == 0 {
		return nil
	}
	return t.DO.Create(values)
}

func (t tbShopDo) CreateInBatches(values []*model.TbShop, batchSize int) error {
	return t.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (t tbShopDo) Save(values ...*model.TbShop) error {
	if len(values) == 0 {
		return nil
	}
	return t.DO.Save(values)
}

func (t tbShopDo) First() (*model.TbShop, error) {
	if result, err := t.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*model.TbShop), nil
	}
}

func (t tbShopDo) Take() (*model.TbShop, error) {
	if result, err := t.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*model.TbShop), nil
	}
}

func (t tbShopDo) Last() (*model.TbShop, error) {
	if result, err := t.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*model.TbShop), nil
	}
}

func (t tbShopDo) Find() ([]*model.TbShop, error) {
	result, err := t.DO.Find()
	return result.([]*model.TbShop), err
}

func (t tbShopDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.TbShop, err error) {
	buf := make([]*model.TbShop, 0, batchSize)
	err = t.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (t tbShopDo) FindInBatches(result *[]*model.TbShop, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return t.DO.FindInBatches(result, batchSize, fc)
}

func (t tbShopDo) Attrs(attrs ...field.AssignExpr) ITbShopDo {
	return t.withDO(t.DO.Attrs(attrs...))
}

func (t tbShopDo) Assign(attrs ...field.AssignExpr) ITbShopDo {
	return t.withDO(t.DO.Assign(attrs...))
}

func (t tbShopDo) Joins(fields ...field.RelationField) ITbShopDo {
	for _, _f := range fields {
		t = *t.withDO(t.DO.Joins(_f))
	}
	return &t
}

func (t tbShopDo) Preload(fields ...field.RelationField) ITbShopDo {
	for _, _f := range fields {
		t = *t.withDO(t.DO.Preload(_f))
	}
	return &t
}

func (t tbShopDo) FirstOrInit() (*model.TbShop, error) {
	if result, err := t.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*model.TbShop), nil
	}
}

func (t tbShopDo) FirstOrCreate() (*model.TbShop, error) {
	if result, err := t.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*model.TbShop), nil
	}
}

func (t tbShopDo) FindByPage(offset int, limit int) (result []*model.TbShop, count int64, err error) {
	result, err = t.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = t.Offset(-1).Limit(-1).Count()
	return
}

func (t tbShopDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = t.Count()
	if err != nil {
		return
	}

	err = t.Offset(offset).Limit(limit).Scan(result)
	return
}

func (t tbShopDo) Scan(result interface{}) (err error) {
	return t.DO.Scan(result)
}

func (t tbShopDo) Delete(models ...*model.TbShop) (result gen.ResultInfo, err error) {
	return t.DO.Delete(models)
}

func (t *tbShopDo) withDO(do gen.Dao) *tbShopDo {
	t.DO = *do.(*gen.DO)
	return t
}
