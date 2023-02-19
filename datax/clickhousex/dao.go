package clickhousex

import (
	"database/sql"
	"fmt"
	"github.com/chenxinqun/ginWarpPkg/datax/pagex"
	"github.com/chenxinqun/ginWarpPkg/errno"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"math"
	"reflect"
	"strings"
)

// err 常量
const (
	CreateErrStr  = "create error"
	UpdatesErrStr = "updates error"
	DeleteErrStr  = "delete error"
)

// 查询常量
const (
	DescStr  = "DESC"
	AscStr   = "ASC"
	InStr    = "IN"
	NotInStr = "NOT IN"
)

func NewQueryBuilder(model interface{}, mustBeField ...string) *QueryBuilder {
	/*
		model 要传进来一个结构体的指针.
		传进来的model, 一定要有 gorm:"column:id" 这样的tag
		否则的话, 所有的查询条件都无法执行.
		这样做的目的是为了严格校验查询的字段, 防止SQL注入.

		mustBeField 如果传了, 则会校验. 所有的查询条件中, 缺少这个查询条件, 则会抛出异常.
	*/
	filedSet := make(map[string]struct{})
	typeOf := reflect.TypeOf(model).Elem()
	for i := 0; i < typeOf.NumField(); i++ {
		tag := typeOf.Field(i).Tag.Get("gorm")
		if len(tag) > 0 {
			tagList := strings.Split(tag, " ")
			for _, t := range tagList {
				if strings.HasPrefix(t, "column:") {
					field := strings.Split(t, ":")[1]
					field = strings.Split(field, ",")[0]
					field = strings.Split(field, ";")[0]
					if len(field) > 0 {
						filedSet[strings.TrimSpace(field)] = struct{}{}
					}
				}
			}
		}
	}
	if len(filedSet) == 0 {
		panic(errno.NewError("没有传入fields变量, 或者model中没有定义 grom:\"column:filed\" tag"))
	}
	mustMap := make(map[string]bool)
	for _, filed := range mustBeField {
		mustMap[strings.TrimSpace(filed)] = false
	}
	ret := new(QueryBuilder)
	ret.model = model
	ret.filedSet = filedSet
	ret.mustBeField = mustMap
	return ret
}

type QueryBuilder struct {
	limit       int
	offset      int
	mustBeField map[string]bool
	order       []string
	group       []string
	where       []struct {
		prefix string
		value  interface{}
	}
	filedSet map[string]struct{}
	model    interface{}
}

func (qb *QueryBuilder) Transaction(db *gorm.DB, fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) (err error) {
	return db.Transaction(fc, opts...)
}

func (qb *QueryBuilder) Create(db *gorm.DB, model interface{}) (rowsAffected int64, err error) {
	tx := db.Create(model)
	rowsAffected = tx.RowsAffected
	if err = tx.Error; err != nil {
		return rowsAffected, errno.Wrap(err, CreateErrStr)
	}
	return rowsAffected, nil
}

func (qb *QueryBuilder) Updates(db *gorm.DB, m map[string]interface{}) (rowsAffected int64, err error) {
	db = qb.buildModel(db)
	tx := qb.buildWhere(db).Updates(m)
	rowsAffected = tx.RowsAffected
	if err = tx.Error; err != nil {
		return rowsAffected, errno.Wrap(err, UpdatesErrStr)
	}

	return rowsAffected, nil
}

func (qb *QueryBuilder) Delete(db *gorm.DB) (rowsAffected int64, err error) {
	// 至少有一个查询条件才允许删除
	if len(qb.where) > 0 {
		tx := qb.buildWhere(db).Delete(qb.model)
		rowsAffected = tx.RowsAffected
		if err = tx.Error; err != nil {
			return rowsAffected, errno.Wrap(err, DeleteErrStr)
		}
	}

	return rowsAffected, nil
}

func (qb *QueryBuilder) Count(db *gorm.DB) (int64, error) {
	var c int64
	err := qb.buildQuery(db).Count(&c).Error
	if err != nil {
		c = 0
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return c, err
}

func (qb *QueryBuilder) Exist(db *gorm.DB) (bool, error) {
	var c int64
	c, err := qb.Count(db)

	return c > 0, err
}

func (qb *QueryBuilder) Page(db *gorm.DB, val interface{}, param *pagex.Params) (*pagex.Result, error) {
	if param == nil {
		param = pagex.DefaultPageOption()
	}
	if param.PageSize == 0 {
		param.PageSize = pagex.DefaultPageOption().PageSize
	}
	if param.CurrentPage == 0 {
		param.CurrentPage = pagex.DefaultPageOption().CurrentPage
	}

	// 统计总数
	totalSize, err := qb.Count(db)
	if err != nil {
		return nil, err
	}

	// 如果传了all=true, 则设当前页为第一页,分页大小为数据总量. 查询所有(与list等效)
	if param.All {
		param.PageSize = int(totalSize)
		param.CurrentPage = 1
	}
	var (
		pageSize    = param.PageSize
		currentPage = param.CurrentPage
		sortField   = param.SortField
		sortType    = param.SortType
	)
	// 总数为0时
	if totalSize == 0 {
		ret := &pagex.Result{
			Total:       0,
			TotalPage:   0,
			PageSize:    pageSize,
			CurrentPage: currentPage,
			CurrentSize: 0,
		}
		return ret, nil
	}
	// 计算总页码
	p := float64(totalSize) / float64(pageSize)
	totalPage := int(math.Ceil(p))
	// 计算当前页大小
	currentSize := pageSize
	if totalSize == 0 {
		currentSize = 0
	} else {
		if currentPage == totalPage {
			currentSize = int(totalSize - int64(pageSize*(totalPage-1)))
		}
	}
	if len(sortField) > 0 {
		for i, field := range sortField {
			if _, ok := qb.filedSet[field]; ok {
				qb.OrderBy(field, sortType[i])
			}

		}
	}
	qb.Limit(pageSize)
	offset := (currentPage - 1) * pageSize
	qb.Offset(offset)
	pageData := &pageDataObj{
		total:       int(totalSize),
		totalPage:   totalPage,
		pageSize:    pageSize,
		currentPage: currentPage,
		currentSize: currentSize,
		result:      qb.buildQuery(db),
	}
	ret, err := pagex.ConvertResult(pageData, val)
	return ret, err
}

func (qb *QueryBuilder) First(db *gorm.DB, ret interface{}) error {
	err := qb.buildQuery(db).First(ret).Error

	return err
}

func (qb *QueryBuilder) Get(db *gorm.DB, ret interface{}) error {
	return qb.First(db, ret)
}

func (qb *QueryBuilder) List(db *gorm.DB, ret interface{}) error {
	err := qb.buildQuery(db).Find(ret).Error

	return err
}

func (qb *QueryBuilder) SetModel(model interface{}) {
	qb.model = model
}

func (qb *QueryBuilder) BuildQuery(db *gorm.DB) *gorm.DB {
	ret := db
	ret = qb.buildModel(ret)
	ret = qb.buildWhere(ret)
	ret = qb.buildOrder(ret)
	ret = qb.buildGroup(ret)
	ret = qb.buildLimit(ret)
	return ret
}

func (qb *QueryBuilder) buildQuery(db *gorm.DB) *gorm.DB {
	ret := db
	ret = qb.buildModel(ret)
	ret = qb.buildWhere(ret)
	ret = qb.buildOrder(ret)
	ret = qb.buildGroup(ret)
	ret = qb.buildLimit(ret)
	return ret
}

func (qb *QueryBuilder) buildModel(db *gorm.DB) *gorm.DB {
	ret := db.Model(qb.model)
	return ret
}

func (qb *QueryBuilder) buildWhere(db *gorm.DB) *gorm.DB {
	ret := db
	for _, where := range qb.where {
		filed := strings.TrimSpace(strings.Split(where.prefix, " ")[0])
		if _, ok := qb.mustBeField[filed]; ok {
			qb.mustBeField[filed] = true
		}
		ret = ret.Where(where.prefix, where.value)
	}
	qualified := true
	k := ""
	for key, value := range qb.mustBeField {
		if !value {
			qualified = false
			k = key
		}
	}
	if !qualified {
		panic(errno.Errorf("where条件中缺少必传字段 \"%s\"", k))
	}
	return ret
}

func (qb *QueryBuilder) buildOrder(db *gorm.DB) *gorm.DB {
	ret := db
	for _, order := range qb.order {
		ret = ret.Order(order)
	}
	return ret
}

func (qb *QueryBuilder) buildGroup(db *gorm.DB) *gorm.DB {
	ret := db
	for _, group := range qb.group {
		ret = ret.Group(group)
	}
	return ret
}

func (qb *QueryBuilder) buildLimit(db *gorm.DB) *gorm.DB {
	ret := db
	ret = ret.Limit(qb.limit).Offset(qb.offset)
	return ret
}

func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

func (qb *QueryBuilder) Where(p Predicate, field string, value interface{}) *QueryBuilder {
	// 校验字段是否存在, 如果不存在, 则直接返回
	if _, ok := qb.filedSet[field]; !ok {
		return qb
	}
	qb.where = append(qb.where, struct {
		prefix string
		value  interface{}
	}{
		fmt.Sprintf("%v %v ?", field, p),
		value,
	})

	return qb
}

func (qb *QueryBuilder) WhereIn(field string, value interface{}) *QueryBuilder {
	// 校验字段是否存在, 如果不存在, 则直接返回
	if _, ok := qb.filedSet[field]; !ok {
		return qb
	}
	qb.where = append(qb.where, struct {
		prefix string
		value  interface{}
	}{
		fmt.Sprintf("%v %v ?", field, InStr),
		value,
	})

	return qb
}

func (qb *QueryBuilder) WhereNotIn(field string, value interface{}) *QueryBuilder {
	// 校验字段是否存在, 如果不存在, 则直接返回
	if _, ok := qb.filedSet[field]; !ok {
		return qb
	}
	qb.where = append(qb.where, struct {
		prefix string
		value  interface{}
	}{
		fmt.Sprintf("%v %v ?", field, NotInStr),
		value,
	})

	return qb
}

func (qb *QueryBuilder) OrderBy(field string, sortType pagex.SortType) *QueryBuilder {
	// 校验字段是否存在, 如果不存在, 则直接返回
	if _, ok := qb.filedSet[field]; !ok {
		return qb
	}
	order := AscStr
	switch sortType {
	case pagex.SortDesc:
		order = DescStr
	}
	qb.order = append(qb.order, field+" "+order)

	return qb
}

func (qb *QueryBuilder) GroupBy(field string) *QueryBuilder {
	// 校验字段是否存在, 如果不存在, 则直接返回
	if _, ok := qb.filedSet[field]; !ok {
		return qb
	}
	qb.order = append(qb.group, field)
	return qb
}
