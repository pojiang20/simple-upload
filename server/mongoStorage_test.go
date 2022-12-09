package server

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
)

const testAddr = "127.0.0.1:27017"

func TestMongoStorage_SetInit(t *testing.T) {
	r := assert.New(t)

	db := NewMongoStorage(testAddr)
	key := "test_setInit_key"
	uploadId1 := "abc"
	info := UploadInfo{
		Key:      key,
		UploadId: uploadId1,
	}
	db.SetInit(info)
	dbHelp := newMongoDB(testAddr)
	query := bson.M{"_id": key}
	var res UploadInfo
	dbHelp.Database(DATABASE_PART).Collection(TABLE_INFO).FindOne(context.TODO(), query).Decode(&res)
	r.Equal(res.UploadId, uploadId1)
}

func TestMongoStorage_SetPart(t *testing.T) {
	r := assert.New(t)

	db := NewMongoStorage(testAddr)
	key := "test_setInit_key"
	etag := UploadPartInfo{
		Etag: "a",
	}
	db.SetPart(key, etag)
	dbHelp := newMongoDB(testAddr)
	filter := bson.M{"_id": key}
	n, err := dbHelp.Database(DATABASE_PART).Collection(TABLE_INFO).CountDocuments(context.TODO(), filter)
	r.NoError(err)
	r.Equal(int64(1), n)
}

func TestNewMongoStorage(t *testing.T) {
	r := assert.New(t)

	storage := NewMongoStorage(testAddr)
	r.NotNil(storage)
	r.IsType(MongoStorage{}, storage)
}

func newMongoDB(addr string) *mongo.Client {
	uri := fmt.Sprintf("mongodb://%s", addr)
	cli, _ := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	return cli
}
