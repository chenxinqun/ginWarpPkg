package mysqlx

import (
	"gorm.io/gorm"
)

type pageDataObj struct {
	total       int
	totalPage   int
	pageSize    int
	currentPage int
	currentSize int
	result      *gorm.DB
}

func (p pageDataObj) Total() int {
	return p.total
}
func (p pageDataObj) TotalPage() int {
	return p.totalPage
}

func (p pageDataObj) PageSize() int {
	return p.pageSize
}
func (p pageDataObj) CurrentSize() int {
	return p.currentSize
}
func (p pageDataObj) CurrentPage() int {
	return p.currentPage
}

func (m pageDataObj) Result(val interface{}) error {
	return m.result.Find(val).Error
}
