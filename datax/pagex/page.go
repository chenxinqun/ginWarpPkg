package pagex

import "github.com/chenxinqun/ginWarpPkg/convert"

type SortType byte

const (
	SortNormal SortType = iota // 正常排序,不做特殊处理,按照db默认排序规则, use default for db's rule
	SortAsc                    // 升序, Ascending
	SortDesc                   // 降序, Descending
)

type Option func(opts *Params)

type Params struct {
	All         bool       `json:"all" form:"all"`                   // 如果all=true, 则查询所有
	PageSize    int        `json:"page_size" form:"page_size"`       // 分页大小
	CurrentPage int        `json:"current_page" form:"current_page"` // 分页索引
	SortField   []string   `json:"sort_field" form:"sort_field"`     // 排序字段
	SortType    []SortType `json:"sort_type" form:"sort_type"`       // 字段的排序类型
}

func DefaultPageOption() *Params {
	return &Params{
		PageSize:    10,
		CurrentPage: 1,
	}
}

// ConvertParams 将拥有相同字段的结构体, 转换为page.Params类型
func ConvertParams(src interface{}) (*Params, error) {
	/*
		src 传入一个结构体, 不要传入指针
	*/
	param := new(Params)
	err := convert.StructToStruct(src, param)
	return param, err
}

type Page interface {
	// Total 总记录数
	Total() int

	// TotalPage 总页数
	TotalPage() int

	// PageSize 要求的分页大小
	PageSize() int

	// CurrentSize 当前页实际大小
	CurrentSize() int

	// CurrentPage 当前页的页码
	CurrentPage() int

	// Result 将分页结果反序列化至 val , val需要是一个切片指针
	Result(val interface{}) error
}

// ConvertResult 转换分页对象并将分页结果反序列化至val中
func ConvertResult(page Page, val interface{}) (*Result, error) {
	if page == nil {
		panic("ConvertResult failed,use nil page")
	}
	if page.CurrentSize() > 0 {
		if e := page.Result(val); e != nil {
			return nil, e
		}
	}
	result := &Result{
		Total:       page.Total(),
		TotalPage:   page.TotalPage(),
		PageSize:    page.PageSize(),
		CurrentPage: page.CurrentPage(),
		CurrentSize: page.CurrentSize(),
		List:        val,
	}
	return result, nil
}

type Result struct {
	Total       int         `json:"total" form:"total"`               // 记录总数
	TotalPage   int         `json:"total_page" form:"total_page"`     // 总页数
	PageSize    int         `json:"page_size" form:"page_size"`       // 要求的分页大小
	CurrentPage int         `json:"current_page" form:"current_page"` // 当前页码
	CurrentSize int         `json:"current_size" form:"current_size"` // 当前页实际大小
	List        interface{} `json:"list" form:"list"`                 //  结果集
}
