package server

// 改接口表示对分片信息持久化操作
type partInfoStorage interface {
	Exist(key string) bool
	SetInit(info UploadInfo)
	GetInit(key string) UploadInfo
	SetPart(key string, etage UploadPartInfo)
	GetPart(key string) []UploadPartInfo
	Close()
}

type UploadInfo struct {
	Key      string           `json:"_id" bson:"_id"`
	UploadId string           `json:"upload_id" bson:"upload_id"`
	Etags    []UploadPartInfo `json:"etags" bson:"etags"`
}

type CompleteExtra struct {
	Progress []UploadPartInfo
}

type UploadPartInfo struct {
	Etag       string
	PartNumber int
	PartSize   int64
}
