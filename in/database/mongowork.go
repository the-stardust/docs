package database

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	clients     map[string]*mongo.Client
	clientsLock sync.Mutex
)

type MongoIndex struct {
	Version int            `json:"v"`
	Key     map[string]int `json:"key"`
	Name    string         `json:"name"`
}

func init() {
	clients = make(map[string]*mongo.Client)
}

func GetModelName(modelFullName string, mw *MongoWork) string {
	var modelName = ""
	if mw.collection == "" {
		fullName := strings.Split(modelFullName, ".")
		modelName = fullName[len(fullName)-1]
	} else {
		modelName = mw.collection
	}
	return modelName
}

type MongoWork struct {
	uri        string
	user       string
	password   string
	database   string
	db         *mongo.Database
	Db         *mongo.Database
	filter     primitive.M
	skip       int64
	limit      int64
	fields     interface{}
	upsert     bool
	sort       []string
	inSession  bool
	ctx        context.Context
	client     *mongo.Client
	session    *mongo.Session
	collection string
}

func NewMongoWork(uri, user, password, database string) *MongoWork {
	mwn := new(MongoWork)
	mwn.uri = uri
	mwn.user = user
	mwn.password = password
	mwn.database = database
	mwn.filter = bson.M{}
	mwn.skip = 0
	mwn.limit = 0
	mwn.sort = make([]string, 0)
	var err error
	key := fmt.Sprintf("%s%s%s", uri, user, password)
	if client, hasKey := clients[key]; hasKey {
		mwn.ctx = context.Background()
		err = client.Ping(mwn.ctx, nil)
		if err == nil {
			mwn.db = client.Database(database)
			mwn.Db = mwn.db
			mwn.client = client
			return mwn
		}
	}

	mwn.client, err = mwn.Init()
	if err != nil {
		return nil
	}
	clientsLock.Lock()
	clients[key] = mwn.client
	clientsLock.Unlock()
	mwn.db = mwn.client.Database(database)
	mwn.Db = mwn.db
	return mwn
}

func (sf *MongoWork) SetDatabase(database string) {
	sf.database = database

	sf.db = sf.client.Database(database)
}

func (sf *MongoWork) Init() (*mongo.Client, error) {
	sf.ctx = context.Background()
	if sf.user != "" && sf.password != "" {
		sf.uri = fmt.Sprintf("mongodb://%s:%s@%s/admin?authSource=%s", sf.user, sf.password, sf.uri, sf.database)
	} else {
		sf.uri = fmt.Sprintf("mongodb://%s", sf.uri)
	}
	config := options.Client().ApplyURI(sf.uri)
	config.SetMaxPoolSize(100)
	config.SetMaxConnIdleTime(30 * time.Second)
	// if sf.user != "" && sf.password != "" {
	// 	config.Auth = &options.Credential{
	// 		Username: sf.user,
	// 		Password: sf.password,
	// 	}
	// }

	return mongo.Connect(sf.ctx, config)
}
func (sf *MongoWork) Collection(name string) *MongoWork {
	sf.collection = name
	return sf
}
func (sf *MongoWork) Where(filter primitive.M) *MongoWork {
	sf.filter = filter
	return sf
}

func (sf *MongoWork) Skip(n int64) *MongoWork {
	sf.skip = n
	return sf
}

func (sf *MongoWork) Limit(n int64) *MongoWork {
	sf.limit = n
	return sf
}

func (sf *MongoWork) Sort(fields ...string) *MongoWork {
	sf.sort = fields
	return sf
}
func (sf *MongoWork) Fields(f interface{}) *MongoWork {
	sf.fields = f
	return sf
}
func (sf *MongoWork) Upsert(f bool) *MongoWork {
	sf.upsert = f
	return sf
}
func (sf *MongoWork) getFindOptions() *options.FindOptions {
	opt := options.FindOptions{}
	isAllowDiskUse := true
	opt.AllowDiskUse = &isAllowDiskUse
	if sf.skip > 0 {
		opt.SetSkip(sf.skip)
	}
	if sf.limit > 0 {
		opt.SetLimit(sf.limit)
	}
	if len(sf.sort) > 0 {
		sort := bson.D{}
		for i := 0; i < len(sf.sort); i++ {
			if sf.sort[i][0:1] == "-" {
				sort = append(sort, bson.E{Key: sf.sort[i][1:], Value: -1})
			} else if sf.sort[i][0:1] == "+" {
				sort = append(sort, bson.E{Key: sf.sort[i][1:], Value: 1})
			} else {
				sort = append(sort, bson.E{Key: sf.sort[i], Value: 1})
			}
		}
		opt.SetSort(sort)
	}
	if sf.fields != nil {
		opt.Projection = sf.fields
	}
	return &opt
}
func (sf *MongoWork) getAggregateOptions() *options.AggregateOptions {
	opt := options.AggregateOptions{}
	isAllowDiskUse := true
	opt.AllowDiskUse = &isAllowDiskUse
	return &opt
}
func (sf *MongoWork) getFindOneOptions() *options.FindOneOptions {
	opt := options.FindOneOptions{}
	if sf.skip > 0 {
		opt.SetSkip(sf.skip)
	}
	if len(sf.sort) > 0 {
		sort := bson.D{}
		for i := 0; i < len(sf.sort); i++ {
			if sf.sort[i][0:1] == "-" {
				sort = append(sort, bson.E{Key: sf.sort[i][1:], Value: -1})
			} else if sf.sort[i][0:1] == "+" {
				sort = append(sort, bson.E{Key: sf.sort[i][1:], Value: 1})
			} else {
				sort = append(sort, bson.E{Key: sf.sort[i], Value: 1})
			}
		}
		opt.SetSort(sort)
	}
	if sf.fields != nil {
		opt.Projection = sf.fields
	}
	return &opt
}
func (sf *MongoWork) getFindOneAndUpdateOptions() *options.FindOneAndUpdateOptions {
	opt := options.FindOneAndUpdateOptions{}
	opt.SetUpsert(sf.upsert)
	opt.SetUpsert(true)
	if len(sf.sort) > 0 {
		sort := bson.D{}
		for i := 0; i < len(sf.sort); i++ {
			if sf.sort[i][0:1] == "-" {
				sort = append(sort, bson.E{Key: sf.sort[i][1:], Value: -1})
			} else if sf.sort[i][0:1] == "+" {
				sort = append(sort, bson.E{Key: sf.sort[i][1:], Value: 1})
			} else {
				sort = append(sort, bson.E{Key: sf.sort[i], Value: 1})
			}
		}
		opt.SetSort(sort)
	}
	if sf.fields != nil {
		opt.Projection = sf.fields
	}
	return &opt
}
func (sf *MongoWork) Find(doc interface{}) error {
	if reflect.TypeOf(doc).Kind().String() == "ptr" {
		modelName := GetModelName(reflect.TypeOf(doc).String(), sf)
		cur, err := sf.db.Collection(modelName).Find(sf.ctx, sf.filter, sf.getFindOptions())
		if err != nil {
			return err
		}
		defer func() {
			err = cur.Close(sf.ctx)
		}()
		return cur.All(sf.ctx, doc)
		// err = cur.All(sf.ctx, doc)
		// if err != nil {
		// 	return err
		// }
		// if reflect.ValueOf(doc).Elem().Len() == 0 {
		// 	return fmt.Errorf("mongo: no documents in result")
		// }
		// return nil
	}
	return fmt.Errorf("need ptr")
}

//	func (sf *MongoWork) FindOneAndUpdate(doc interface{}, update interface{}) error {
//		if reflect.TypeOf(doc).Kind().String() == "ptr" {
//			modelName := GetModelName(reflect.TypeOf(doc).String(), sf)
//			opt := sf.getFindOneAndUpdateOptions()
//			opt.SetUpsert(true)
//			opt.SetArrayFilters(options.ArrayFilters{
//				Filters: []interface{}{bson.M{"elem.manager_id": "9"}},
//			})
//			err := sf.db.Collection(modelName).FindOneAndUpdate(sf.ctx, sf.filter, update, opt).Decode(doc)
//			if err != nil {
//				return err
//			}
//			return nil
//		}
//		return fmt.Errorf("need ptr")
//	}
func (sf *MongoWork) Take(doc interface{}) error {
	if reflect.TypeOf(doc).Kind().String() == "ptr" {
		modelName := GetModelName(reflect.TypeOf(doc).String(), sf)
		err := sf.db.Collection(modelName).FindOne(sf.ctx, sf.filter, sf.getFindOneOptions()).Decode(doc)
		if err != nil {
			return err
		}

		defer func() {
		}()

		return nil
	}
	return fmt.Errorf("need ptr")
}

func (sf *MongoWork) Count() (int64, error) {
	opt := options.CountOptions{}
	if sf.skip > 0 {
		opt.SetSkip(sf.skip)
	}
	if sf.limit > 0 {
		opt.SetLimit(sf.limit)
	}
	defer func() {
	}()
	return sf.db.Collection(sf.collection).CountDocuments(sf.ctx, sf.filter, &opt)
}

func (sf *MongoWork) AggregateCount(filter primitive.A) (int64, error) {
	filter = append(filter, bson.M{"$count": "count"})

	type CountStruct struct {
		Count int64 `bson:"count"`
	}
	var count CountStruct
	err := sf.AggregateOne(filter, &count)
	if err != nil {
		return 0, err
	}
	return count.Count, nil
}

func (sf *MongoWork) aggregate(filter primitive.A, doc interface{}) (*mongo.Cursor, error) {
	if reflect.TypeOf(doc).Kind().String() == "ptr" {
		if len(sf.sort) > 0 {
			sort := bson.M{}
			for i := 0; i < len(sf.sort); i++ {
				if sf.sort[i][0:1] == "-" {
					sort[sf.sort[i][1:]] = -1
				} else if sf.sort[i][0:1] == "+" {
					sort[sf.sort[i][1:]] = 1
				} else {
					sort[sf.sort[i]] = 1
				}
			}
			filter = append(filter, bson.M{"$sort": sort})
		}

		if sf.skip > 0 {
			filter = append(filter, bson.M{"$skip": sf.skip})
		}
		if sf.limit > 0 {
			filter = append(filter, bson.M{"$limit": sf.limit})
		}
		return sf.db.Collection(sf.collection).Aggregate(sf.ctx, filter, sf.getAggregateOptions())
	}
	return nil, fmt.Errorf("need ptr")
}

func (sf *MongoWork) Aggregate(filter primitive.A, doc interface{}) error {
	cur, err := sf.aggregate(filter, doc)
	if err != nil {
		return err
	}
	defer func() {
		err = cur.Close(sf.ctx)
	}()
	err = cur.All(sf.ctx, doc)

	return err
}

func (sf *MongoWork) AggregateOne(filter primitive.A, doc interface{}) error {
	cur, err := sf.aggregate(filter, doc)
	if err != nil {
		return err
	}
	defer func() {
		err = cur.Close(sf.ctx)
	}()
	if cur.Next(sf.ctx) {
		err = cur.Decode(doc)

		if err != nil {
			return err
		}

		return nil
	}
	return fmt.Errorf("mongo: no documents in result")
}

func (sf *MongoWork) TakeField(model interface{}, field string) interface{} {
	typeOf := reflect.TypeOf(model)
	object := reflect.New(typeOf)
	instance := object.Interface()

	err := sf.Take(instance)
	if err == nil {
		fields := strings.Split(field, ".")
		objectInstance := object.Interface()
		for i := 0; i < len(fields); i++ {
			fieldName := fields[i]
			objectInstance = sf.getFieldValue(fieldName, objectInstance)
			if objectInstance != nil {
				if i == len(fields)-1 {
					return objectInstance
				}
			}
		}
	}
	return nil
}

func (sf *MongoWork) getFieldValue(fieldName string, object interface{}) interface{} {
	valueOf := reflect.ValueOf(object)
	typeOf := reflect.TypeOf(object)
	if typeOf.Kind() == reflect.Ptr {
		valueOf = valueOf.Elem()
		typeOf = reflect.TypeOf(valueOf.Interface())
	}

	for j := 0; j < typeOf.NumField(); j++ {
		if typeOf.Field(j).Tag.Get("bson") == fieldName {
			return valueOf.Field(j).Interface()
		}
	}
	return nil
}

func (sf *MongoWork) Save(doc interface{}) error {
	if reflect.TypeOf(doc).Kind().String() == "ptr" {
		if len(sf.filter) == 0 {
			id := reflect.ValueOf(doc).Elem().FieldByName("Id")
			if id.IsValid() {
				sf.filter = bson.M{"_id": id.Interface()}
			}
		}
		vof := reflect.ValueOf(doc)
		// for i := 0; i < vof.NumMethod(); i++ {
		// 	vof.Method(i).Call([]reflect.Value{})
		// }
		defaultUpdateAt := vof.MethodByName("DefaultUpdateAt")
		if defaultUpdateAt.Kind() != reflect.Invalid {
			defaultUpdateAt.Call([]reflect.Value{})
		}
		defaultCreateAt := vof.MethodByName("DefaultCreateAt")
		if defaultCreateAt.Kind() != reflect.Invalid {
			defaultCreateAt.Call([]reflect.Value{})
		}
		defaultId := vof.MethodByName("DefaultId")
		if defaultId.Kind() != reflect.Invalid {
			defaultId.Call([]reflect.Value{})
		}
		modelName := GetModelName(reflect.TypeOf(doc).String(), sf)
		var err error
		_, err = sf.db.Collection(modelName).ReplaceOne(sf.ctx, sf.filter, doc)

		defer func() {
		}()

		return err
	}
	return fmt.Errorf("need ptr")
}
func (sf *MongoWork) StaticSave(doc interface{}) error {
	if reflect.TypeOf(doc).Kind().String() == "ptr" {
		if len(sf.filter) == 0 {
			id := reflect.ValueOf(doc).Elem().FieldByName("Id")
			if id.IsValid() {
				sf.filter = bson.M{"_id": id.Interface()}
			}
		}
		vof := reflect.ValueOf(doc)
		defaultCreateAt := vof.MethodByName("DefaultCreateAt")
		if defaultCreateAt.Kind() != reflect.Invalid {
			defaultCreateAt.Call([]reflect.Value{})
		}
		defaultId := vof.MethodByName("DefaultId")
		if defaultId.Kind() != reflect.Invalid {
			defaultId.Call([]reflect.Value{})
		}
		modelName := GetModelName(reflect.TypeOf(doc).String(), sf)
		var err error
		_, err = sf.db.Collection(modelName).ReplaceOne(sf.ctx, sf.filter, doc)

		defer func() {
		}()

		return err
	}
	return fmt.Errorf("need ptr")
}
func (sf *MongoWork) Create(doc interface{}) (*mongo.InsertOneResult, error) {
	if reflect.TypeOf(doc).Kind().String() == "ptr" {
		modelName := GetModelName(reflect.TypeOf(doc).String(), sf)
		vof := reflect.ValueOf(doc)
		defaultUpdateAt := vof.MethodByName("DefaultUpdateAt")
		if defaultUpdateAt.Kind() != reflect.Invalid {
			defaultUpdateAt.Call([]reflect.Value{})
		}
		defaultCreateAt := vof.MethodByName("DefaultCreateAt")
		if defaultCreateAt.Kind() != reflect.Invalid {
			defaultCreateAt.Call([]reflect.Value{})
		}
		defaultId := vof.MethodByName("DefaultId")
		if defaultId.Kind() != reflect.Invalid {
			defaultId.Call([]reflect.Value{})
		}

		// for i := 0; i < vof.NumMethod(); i++ {
		// 	vof.Method(i).Call([]reflect.Value{})
		// }
		var result *mongo.InsertOneResult
		var err error

		result, err = sf.db.Collection(modelName).InsertOne(sf.ctx, doc)

		defer func() {
		}()

		return result, err
	}
	return nil, fmt.Errorf("need ptr")
}

func (sf *MongoWork) Delete(model interface{}) (*mongo.DeleteResult, error) {
	modelName := GetModelName(reflect.TypeOf(model).String(), sf)
	var result *mongo.DeleteResult
	var err error
	if sf.limit == 1 {
		result, err = sf.db.Collection(modelName).DeleteOne(sf.ctx, sf.filter)
	} else {
		result, err = sf.db.Collection(modelName).DeleteMany(sf.ctx, sf.filter)
	}

	defer func() {
	}()

	return result, err
}

func (sf *MongoWork) Update(set interface{}) (*mongo.UpdateResult, error) {
	var result *mongo.UpdateResult
	var err error
	if sf.limit == 1 {
		result, err = sf.db.Collection(sf.collection).UpdateOne(sf.ctx, sf.filter, bson.M{"$set": set})
	} else {
		result, err = sf.db.Collection(sf.collection).UpdateMany(sf.ctx, sf.filter, bson.M{"$set": set})
	}

	defer func() {
	}()

	return result, err
}
func (sf *MongoWork) UpdateAtom(set interface{}) (*mongo.UpdateResult, error) {
	var result *mongo.UpdateResult
	var err error
	if sf.limit == 1 {
		result, err = sf.db.Collection(sf.collection).UpdateOne(sf.ctx, sf.filter, set)
	} else {
		result, err = sf.db.Collection(sf.collection).UpdateMany(sf.ctx, sf.filter, set)
	}
	defer func() {
	}()

	return result, err
}
func (sf *MongoWork) NoResult(err error) bool {
	if err == nil {
		return false
	}
	if err.Error() == "mongo: no documents in result" {
		return true
	}
	return false
}
func (sf *MongoWork) Begin() *MongoWork {
	mwn := new(MongoWork)
	mwn.uri = sf.uri
	mwn.user = sf.user
	mwn.password = sf.password
	mwn.database = sf.database
	mwn.filter = bson.M{}
	mwn.skip = 0
	mwn.limit = 0
	mwn.sort = make([]string, 0)
	var err error
	mwn.client, err = mwn.Init()
	if err != nil {
		return nil
	}
	mwn.db = mwn.client.Database(sf.database)
	session, err := mwn.client.StartSession()
	if err != nil {
		return nil
	}
	mwn.session = &session
	err = session.StartTransaction()
	if err != nil {
		return nil
	}
	sc := mongo.NewSessionContext(mwn.ctx, session)
	mwn.ctx = sc
	return mwn
}

func (sf *MongoWork) Rollback() error {
	if sf.session != nil {
		return fmt.Errorf("no session")
	}

	defer func() {
		if sf.session != nil {
			session := *sf.session
			session.EndSession(sf.ctx)
		}
	}()

	session := *sf.session
	return session.AbortTransaction(sf.ctx)
}

func (sf *MongoWork) Commit() error {
	if sf.session != nil {
		return fmt.Errorf("no session")
	}

	defer func() {
		if sf.session != nil {
			session := *sf.session
			session.EndSession(sf.ctx)
		}
	}()

	session := *sf.session
	return session.CommitTransaction(sf.ctx)
}

func (sf *MongoWork) GetIndexes(model interface{}) ([]MongoIndex, error) {
	modelName := GetModelName(reflect.TypeOf(model).String(), sf)
	indexes := sf.db.Collection(modelName).Indexes()
	cur, err := indexes.List(sf.ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = cur.Close(sf.ctx)
	}()

	var mis []MongoIndex
	err = cur.All(sf.ctx, &mis)
	if err != nil {
		return nil, err
	}
	return mis, nil
}

func (sf *MongoWork) EnsureIndex(model interface{}, key string, sort int) error {
	modelName := GetModelName(reflect.TypeOf(model).String(), sf)
	_, err := sf.db.Collection(modelName).Indexes().CreateOne(sf.ctx, mongo.IndexModel{Keys: bson.D{{key, sort}}})
	return err
}

func (sf *MongoWork) InsertMany(docs []interface{}) (*mongo.InsertManyResult, error) {
	res, err := sf.db.Collection(sf.collection).InsertMany(sf.ctx, docs)
	return res, err
}

func (sf *MongoWork) BulkWrite(docs []mongo.WriteModel) (*mongo.BulkWriteResult, error) {
	res, err := sf.db.Collection(sf.collection).BulkWrite(sf.ctx, docs)
	return res, err
}

func (sf *MongoWork) Distinct(fieldName string) ([]interface{}, int, error) {
	respSlice, err := sf.db.Collection(sf.collection).Distinct(sf.ctx, fieldName, sf.filter)
	return respSlice, len(respSlice), err
}
