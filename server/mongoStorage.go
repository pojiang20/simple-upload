package server

import (
	"context"
	"fmt"
	"github.com/pojiang20/simple-upload/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"
)

const (
	DATABASE_PART = "PartUpload"
	TABLE_INFO    = "Info"
	RETRY_TIME    = 5
)

type MongoStorage struct {
	client *mongo.Client
	lock   sync.Mutex
}

func NewMongoStorage(addr string) partInfoStorage {
	uri := fmt.Sprintf("mongodb://%s", addr)
	cli, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		util.Zlog.Errorf("NewMongoStorage error: %v", err)
		return nil
	}
	return MongoStorage{
		client: cli,
	}
}

func (m MongoStorage) Close() {
	m.client.Disconnect(context.TODO())
}

func (m MongoStorage) Exist(key string) bool {
	query := bson.M{
		"_id": key,
	}
	count, err := m.client.Database(DATABASE_PART).Collection(TABLE_INFO).CountDocuments(context.TODO(), query)
	if err != nil {
		return false
	}
	if count != 1 {
		return false
	}
	return true
}

func (m MongoStorage) SetInit(info UploadInfo) {
	data := &UploadInfo{
		Key:      info.Key,
		UploadId: info.UploadId,
	}

	retry := 0
	for {
		if retry > RETRY_TIME {
			util.Zlog.Fatal("mongo insert failed")
		}
		_, err := m.client.Database(DATABASE_PART).Collection(TABLE_INFO).InsertOne(context.TODO(), data)
		if err != nil {
			util.Zlog.Infof("mongo insert error: %v,so I will sleep 1s and retry", err)
			time.Sleep(time.Second)
			retry++
			continue
		}
		break
	}
	return
}

func (m MongoStorage) GetInit(key string) (res UploadInfo) {
	query := bson.M{
		"_id": key,
	}

	retry := 0
	for {
		if retry > RETRY_TIME {
			util.Zlog.Fatal("mongo get failed")
		}
		err := m.client.Database(DATABASE_PART).Collection(TABLE_INFO).FindOne(context.TODO(), query).Decode(&res)
		if err != nil {
			util.Zlog.Infof("mongo get error: %v,so I will sleep 1s and retry", err)
			time.Sleep(time.Second)
			retry++
			continue
		}
		break
	}
	return
}

func (m MongoStorage) SetPart(key string, etage UploadPartInfo) {
	parts := m.GetPart(key)
	m.lock.Lock()
	parts = append(parts, etage)
	m.lock.Unlock()
	filter := bson.M{"_id": key}
	update := bson.M{
		"$set": bson.M{"etags": parts},
	}

	retry := 0
	for {
		if retry > RETRY_TIME {
			util.Zlog.Fatal("mongo insert failed")
		}
		_, err := m.client.Database(DATABASE_PART).Collection(TABLE_INFO).UpdateOne(context.TODO(), filter, update)
		if err != nil {
			util.Zlog.Infof("mongo insert error: %v,so I will sleep 1s and retry", err)
			time.Sleep(time.Second)
			retry++
			continue
		}
		break
	}
	return
}

func (m MongoStorage) GetPart(key string) []UploadPartInfo {
	query := bson.M{
		"_id": key,
	}

	res := UploadInfo{}
	retry := 0
	for {
		if retry > RETRY_TIME {
			util.Zlog.Fatal("mongo get failed")
			return nil
		}
		err := m.client.Database(DATABASE_PART).Collection(TABLE_INFO).FindOne(context.TODO(), query).Decode(&res)
		if err != nil {
			util.Zlog.Infof("mongo get error: %v,so I will sleep 1s and retry", err)
			time.Sleep(time.Second)
			retry++
			continue
		}
		break
	}
	return res.Etags
}
