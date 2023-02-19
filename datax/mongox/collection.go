package mongox

import (
	"context"
	"errors"
	page "github.com/chenxinqun/ginWarpPkg/datax/pagex"
	"math"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Collection is a handle to a MongoDB collection. It is safe for concurrent use by multiple goroutines.
type Collection interface {
	Clone(opts ...*options.CollectionOptions) (*mongo.Collection, error)
	Name() string
	Database() *mongo.Database
	BulkWrite(ctx context.Context, models []mongo.WriteModel,
		opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error)
	// InsertOne 插入一条
	InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	// InsertMany 插入多条
	InsertMany(ctx context.Context, documents []interface{},
		opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error)
	// DeleteOne 删除一条
	DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
	// DeleteMany 删除多条
	DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
	// UpdateOne 更新一条
	UpdateOne(ctx context.Context, filter interface{}, update interface{},
		opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	// UpdateMany 更新多条
	UpdateMany(ctx context.Context, filter interface{}, update interface{},
		opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	// ReplaceOne 替换一条
	ReplaceOne(ctx context.Context, filter interface{}, replacement interface{},
		opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error)
	// Aggregate 管道聚合操作
	Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error)
	// CountDocuments 统计文档数量
	CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error)
	EstimatedDocumentCount(ctx context.Context, opts ...*options.EstimatedDocumentCountOptions) (int64, error)
	Distinct(ctx context.Context, fieldName string, filter interface{},
		opts ...*options.DistinctOptions) ([]interface{}, error)
	// Find 自定义查询
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error)

	// 	coll := mongox.Default().DefaultDatabase().Collection("TestReplicate")
	//	pageData, err := coll.FindPage(context.Background(), bson.M{}, //
	//		pagex.CurrentPage(2),                      // 分页
	//		pagex.PageAddSort("rand", pagex.SortAsc), // 排序字段
	//	)
	//	if err != nil {
	//		panic(err)
	//	}
	//	replicates := make([]TestReplicate, 0, pageData.CurrentSize())
	//	pageResult, err := pagex.ConvertResult(pageData, &replicates)
	//	if err != nil {
	//		panic(err)
	//	}
	//	log.Printf("%v", pageResult)

	// FindPage 分页查询
	FindPage(ctx context.Context, filter interface{}, param *page.Params) (page.Page, error)
	// FindOne 查询一条
	FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult
	// FindOneAndDelete 查询一条并删除
	FindOneAndDelete(ctx context.Context, filter interface{}, opts ...*options.FindOneAndDeleteOptions) *mongo.SingleResult
	// FindOneAndReplace 查询一条并替换
	FindOneAndReplace(ctx context.Context, filter interface{}, replacement interface{},
		opts ...*options.FindOneAndReplaceOptions) *mongo.SingleResult
	// FindOneAndUpdate 查询一条并更新
	FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{},
		opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult
	Watch(ctx context.Context, pipeline interface{}, opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error)
	Indexes() mongo.IndexView
	Drop(ctx context.Context) error
}

func newDummyCollection(db *dynamicDataBase, collectionName string, opts ...*options.CollectionOptions) Collection {
	d := dummyCollection{
		db:             db,
		collectionName: collectionName,
		opts:           opts,
	}

	return d
}

//  see more at /go.mongodb.org/mongo-driver@v1.7.2/mongo/collection.go
type dummyCollection struct {
	db             *dynamicDataBase
	collectionName string
	opts           []*options.CollectionOptions // database options
}

func (d dummyCollection) obtainCollection(ctx context.Context) (*mongo.Collection, error) {
	db, err := d.db.obtainDatabase(ctx) // todo 这两个步骤都需要cache
	if err != nil {
		return nil, err
	}
	// collection := db.Collection(d.collectionName)
	// return collection, nil
	c := db.Collection(d.collectionName, d.opts...)
	return c, nil
}

var CloneError = errors.New("dummyCollection not support clone")

func (d dummyCollection) Clone(opts ...*options.CollectionOptions) (*mongo.Collection, error) {
	err := CloneError
	return nil, err
}

func (d dummyCollection) Name() string {
	return d.collectionName
}

func (d dummyCollection) Database() *mongo.Database {
	// panic(errors.New("dummyCollection not support Database()")).
	return d.db.database
}

func (d dummyCollection) BulkWrite(ctx context.Context, models []mongo.WriteModel,
	opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	collection, err := d.obtainCollection(ctx)
	if err != nil {
		return nil, err
	}
	return collection.BulkWrite(ctx, models, opts...)
}

func (d dummyCollection) InsertOne(ctx context.Context, document interface{},
	opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
	collection, err := d.obtainCollection(ctx)
	if err != nil {
		return nil, err
	}
	return collection.InsertOne(ctx, document, opts...)
}

func (d dummyCollection) InsertMany(ctx context.Context, documents []interface{},
	opts ...*options.InsertManyOptions) (*mongo.InsertManyResult, error) {
	collection, err := d.obtainCollection(ctx)
	if err != nil {
		return nil, err
	}
	return collection.InsertMany(ctx, documents, opts...)
}

func (d dummyCollection) DeleteOne(ctx context.Context, filter interface{},
	opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	collection, err := d.obtainCollection(ctx)
	if err != nil {
		return nil, err
	}
	return collection.DeleteOne(ctx, filter, opts...)
}

func (d dummyCollection) DeleteMany(ctx context.Context, filter interface{},
	opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
	collection, err := d.obtainCollection(ctx)
	if err != nil {
		return nil, err
	}
	return collection.DeleteMany(ctx, filter, opts...)
}

func (d dummyCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{},
	opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	collection, err := d.obtainCollection(ctx)
	if err != nil {
		return nil, err
	}
	return collection.UpdateOne(ctx, filter, update, opts...)
}

func (d dummyCollection) UpdateMany(ctx context.Context, filter interface{}, update interface{},
	opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	collection, err := d.obtainCollection(ctx)
	if err != nil {
		return nil, err
	}
	return collection.UpdateMany(ctx, filter, update, opts...)
}

func (d dummyCollection) ReplaceOne(ctx context.Context, filter interface{}, replacement interface{},
	opts ...*options.ReplaceOptions) (*mongo.UpdateResult, error) {
	collection, err := d.obtainCollection(ctx)
	if err != nil {
		return nil, err
	}
	return collection.ReplaceOne(ctx, filter, replacement, opts...)
}

func (d dummyCollection) Aggregate(ctx context.Context, pipeline interface{},
	opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	collection, err := d.obtainCollection(ctx)
	if err != nil {
		return nil, err
	}
	return collection.Aggregate(ctx, pipeline, opts...)
}

func (d dummyCollection) CountDocuments(ctx context.Context, filter interface{},
	opts ...*options.CountOptions) (int64, error) {
	collection, err := d.obtainCollection(ctx)
	if err != nil {
		return 0, err
	}
	return collection.CountDocuments(ctx, filter, opts...)
}

func (d dummyCollection) EstimatedDocumentCount(ctx context.Context,
	opts ...*options.EstimatedDocumentCountOptions) (int64, error) {
	collection, err := d.obtainCollection(ctx)
	if err != nil {
		return 0, err
	}
	return collection.EstimatedDocumentCount(ctx, opts...)
}

func (d dummyCollection) Distinct(ctx context.Context, fieldName string, filter interface{},
	opts ...*options.DistinctOptions) ([]interface{}, error) {
	collection, err := d.obtainCollection(ctx)
	if err != nil {
		return nil, err
	}
	return collection.Distinct(ctx, fieldName, filter, opts...)
}

func (d dummyCollection) Find(ctx context.Context, filter interface{},
	opts ...*options.FindOptions) (*mongo.Cursor, error) {
	collection, err := d.obtainCollection(ctx)
	if err != nil {
		return nil, err
	}
	if filter == nil {
		filter = bson.M{}
	}
	return collection.Find(ctx, filter, opts...)
}

func (d dummyCollection) FindPage(ctx context.Context, filter interface{}, param *page.Params) (page.Page, error) {
	if filter == nil {
		filter = bson.M{}
	}
	counts, err := d.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}
	if param == nil {
		param = page.DefaultPageOption()
	}
	if param.PageSize == 0 {
		param.PageSize = page.DefaultPageOption().PageSize
	}
	if param.CurrentPage == 0 {
		param.CurrentPage = page.DefaultPageOption().CurrentPage
	}
	var (
		pageSize    = param.PageSize
		currentPage = param.CurrentPage
		sortField   = param.SortField
		sortType    = param.SortType
	)
	findOpts := options.Find()
	// set pagex
	if pageSize > 0 && currentPage > 0 {
		findOpts.SetSkip(int64((currentPage - 1) * pageSize))
		findOpts.SetLimit(int64(pageSize))
	}
	// set sort
	if len(sortField) > 0 {
		sort := bson.D{}
		for i, field := range sortField {
			sType := sortType[i]
			var s = 1 // default to use ASC
			switch sType {
			case page.SortNormal, page.SortAsc:
				s = 1
			case page.SortDesc:
				s = -1
			}
			sort = append(sort, bson.E{
				Key:   field,
				Value: s,
			})
		}
		findOpts.SetSort(sort)
	}
	// execute find
	result, err := d.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	if e := result.Err(); e != nil {
		return nil, e
	}
	//
	totalPage := int(math.Ceil(float64(counts / int64(pageSize))))

	// actual pagex size
	pd := &mongoPageData{
		totalPage:   totalPage,
		total:       int(counts),
		pageSize:    pageSize,
		currentPage: currentPage,
		currentSize: result.RemainingBatchLength(),
		result:      result,
	}
	return pd, nil
}

type mongoPageData struct {
	total       int
	totalPage   int
	pageSize    int
	currentPage int
	currentSize int
	result      *mongo.Cursor
}

func (p mongoPageData) Total() int {
	return p.total
}
func (p mongoPageData) TotalPage() int {
	return p.totalPage
}

func (p mongoPageData) PageSize() int {
	return p.pageSize
}
func (p mongoPageData) CurrentSize() int {
	return p.currentPage
}
func (p mongoPageData) CurrentPage() int {
	return p.currentSize
}
func (p mongoPageData) Result(val interface{}) error {
	return p.result.All(context.Background(), val)
}

func (d dummyCollection) FindOne(ctx context.Context, filter interface{},
	opts ...*options.FindOneOptions) *mongo.SingleResult {
	collection, err := d.obtainCollection(ctx)
	if err != nil {
		panic(err)
	}
	return collection.FindOne(ctx, filter, opts...)
}

func (d dummyCollection) FindOneAndDelete(ctx context.Context, filter interface{},
	opts ...*options.FindOneAndDeleteOptions) *mongo.SingleResult {
	collection, err := d.obtainCollection(ctx)
	if err != nil {
		panic(err)
	}
	return collection.FindOneAndDelete(ctx, filter, opts...)
}

func (d dummyCollection) FindOneAndReplace(ctx context.Context, filter interface{}, replacement interface{},
	opts ...*options.FindOneAndReplaceOptions) *mongo.SingleResult {
	collection, err := d.obtainCollection(ctx)
	if err != nil {
		panic(err)
	}
	return collection.FindOneAndReplace(ctx, filter, replacement, opts...)
}

func (d dummyCollection) FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{},
	opts ...*options.FindOneAndUpdateOptions) *mongo.SingleResult {
	collection, err := d.obtainCollection(ctx)
	if err != nil {
		panic(err)
	}
	return collection.FindOneAndUpdate(ctx, filter, update, opts...)
}

func (d dummyCollection) Watch(ctx context.Context, pipeline interface{},
	opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error) {
	collection, err := d.obtainCollection(ctx)
	if err != nil {
		return nil, err
	}
	return collection.Watch(ctx, pipeline, opts...)
}

func (d dummyCollection) Indexes() mongo.IndexView {
	collection, err := d.obtainCollection(context.TODO())
	if err != nil {
		return collection.Indexes()
	}
	return collection.Indexes()
}

func (d dummyCollection) Drop(ctx context.Context) error {
	collection, err := d.obtainCollection(ctx)
	if err != nil {
		return err
	}
	return collection.Drop(ctx)
}
