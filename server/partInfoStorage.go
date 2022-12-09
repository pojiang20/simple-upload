package server

// 改接口表示对分片信息持久化操作
type partInfoStorage interface {
	Exist(key string) bool
	SetInit(info InitInfo)
	GetInit(key string) InitInfo
	SetPart(uploadId string, etage UploadPartInfo)
	GetPart(key string) []UploadPartInfo
}

type InitInfo struct {
	key      string
	uploadId string
}

type CompleteExtra struct {
	Progress []UploadPartInfo
}

type UploadPartInfo struct {
	Etag       string
	PartNumber int64
	partSize   int
}
