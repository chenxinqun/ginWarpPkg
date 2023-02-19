package mongox

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// var (
//	dynamicDBKey = fmt.Sprintf("dynamic_%s", uuid.New().String())
//)

type DataBase interface {
	// Name returns the name of the database.
	Name() string

	// Collection 获取文档集合
	Collection(name string, opts ...*options.CollectionOptions) Collection
}

type dynamicDataBase struct {
	cli      *mongo.Client
	database *mongo.Database
	dbname   string
	opts     []*options.DatabaseOptions
}

func NewDataBase(cli *mongo.Client, dbname string) DataBase {
	database := cli.Database(dbname)
	db := &dynamicDataBase{
		database: database,
		cli:      cli,
		dbname:   dbname,
	}

	return db
}

func (d dynamicDataBase) Client() *mongo.Client {
	return d.cli
}

func (d dynamicDataBase) obtainDatabase(ctx context.Context) (db *mongo.Database, err error) {
	return d.cli.Database(d.dbname, d.opts...), nil
}

func (d dynamicDataBase) Name() string {
	// return "dynamicDataBase"
	return d.dbname
}

func (d *dynamicDataBase) Collection(name string, opts ...*options.CollectionOptions) Collection {
	collection := newDummyCollection(d, name, opts...)
	return collection
}

func (d dynamicDataBase) Aggregate(ctx context.Context, pipeline interface{},
	opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	return d.database.Aggregate(ctx, pipeline, opts...)
}

func (d dynamicDataBase) RunCommand(ctx context.Context, runCommand interface{},
	opts ...*options.RunCmdOptions) *mongo.SingleResult {
	return d.database.RunCommand(ctx, runCommand, opts...)
}

func (d dynamicDataBase) RunCommandCursor(ctx context.Context, runCommand interface{},
	opts ...*options.RunCmdOptions) (*mongo.Cursor, error) {
	return d.database.RunCommandCursor(ctx, runCommand, opts...)
}

func (d dynamicDataBase) Drop(ctx context.Context) error {
	return d.database.Drop(ctx)
}

func (d dynamicDataBase) ListCollections(ctx context.Context, filter interface{},
	opts ...*options.ListCollectionsOptions) (*mongo.Cursor, error) {
	return d.database.ListCollections(ctx, filter, opts...)
}

func (d dynamicDataBase) ListCollectionNames(ctx context.Context, filter interface{},
	opts ...*options.ListCollectionsOptions) ([]string, error) {
	return d.database.ListCollectionNames(ctx, filter, opts...)
}

func (d dynamicDataBase) ReadConcern() *readconcern.ReadConcern {
	return d.database.ReadConcern()
}

func (d dynamicDataBase) ReadPreference() *readpref.ReadPref {
	return d.database.ReadPreference()
}

func (d dynamicDataBase) WriteConcern() *writeconcern.WriteConcern {
	return d.database.WriteConcern()
}

func (d dynamicDataBase) Watch(ctx context.Context, pipeline interface{},
	opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error) {
	return d.database.Watch(ctx, pipeline, opts...)
}
