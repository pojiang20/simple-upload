package server

import (
	"fmt"
	"github.com/pojiang20/simple-upload/util"
	"io"
	"os"
	"path"
)

const (
	TEMPUPLOADID = "15445"
)

type uploader struct {
	partInfo partInfoStorage

	cacheDir  string
	publicDir string
}

func NewUploader(publicDir string) (*uploader, error) {
	if len(publicDir) == 0 {
		return nil, util.ErrInvalidArgument
	}
	absPath, _ := os.Getwd()
	cacheDir, _ := util.GenDir(absPath, "cache")
	return &uploader{
		publicDir: publicDir,
		cacheDir:  cacheDir,
	}, nil
}

func (u *uploader) Init(key string) (string, error) {
	if !u.partInfo.Exist(key) {
		return "", fmt.Errorf("file not exist")
	}
	//分配uploadId
	uploadId := genUploadId()
	info := UploadInfo{Key: key, UploadId: uploadId}
	//不存在则记录
	u.partInfo.SetInit(info)
	return uploadId, nil
}

func (u *uploader) UploadPart(body io.Reader, key string, partNum int64, uploadId string) (int64, error) {
	//校验一下key和uploadId
	if !u.keyUploadIdMatch(key, uploadId) {
		util.Zlog.Info("key dose not match uploadId")
		return 0, nil
	}

	partName := fmt.Sprintf("%s_%d", key, partNum)

	fileSize := u.partSave(body, partName)
	util.Zlog.Infof("pieceSave %s success,filesize is %d", partName, fileSize)

	etag := "etag"
	partinfo := UploadPartInfo{
		Etag:       etag,
		PartNumber: partNum,
		partSize:   int(fileSize),
	}
	u.partInfo.SetPart(key, partinfo)
	return fileSize, nil
}

func (u *uploader) Complete(key string, uploadId string, extra CompleteExtra) error {
	if !u.partValid(key, uploadId, extra) {
		return fmt.Errorf("Invalid")
	}
	n, err := u.partMerge(key)
	if err != nil {
		//返回错误消息
		return fmt.Errorf("part merge error:%v", err)
	}
	util.Zlog.Infof("merge success,file size:%d", n)
	return nil
}

func (u *uploader) PartList(key string) ([]UploadPartInfo, error) {
	if !u.partInfo.Exist(key) {
		return nil, fmt.Errorf("key does not exist")
	}
	parts := u.partInfo.GetPart(key)
	return parts, nil
}

// 先不考虑DB，只能申请唯一UploadId
func genUploadId() string {
	return TEMPUPLOADID
}

func (u *uploader) partSave(body io.Reader, name string) int64 {
	f, err := os.Stat(name)
	if os.IsExist(err) {
		util.Zlog.Info("part %s exist", name)
		return f.Size()
	}

	curF, err := os.Create(name)
	defer curF.Close()
	if err != nil {
		util.Zlog.Fatal(err)
	}

	r, err := io.Copy(curF, body)
	if err != nil {
		util.Zlog.Fatal(err)
	}
	return r
}

func (u *uploader) partMerge(key string) (int, error) {
	filePath := path.Join(u.publicDir, key)
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0777)
	if err != nil {
		return 0, fmt.Errorf("file create failed")
	}

	mergeCnt := 0
	partInfo := u.partInfo.GetPart(key)
	for _, item := range partInfo {
		partFile := genPartName(key, item.PartNumber)
		content, err := os.ReadFile(partFile)
		if err != nil {
			return 0, fmt.Errorf("part file open failed")
		}
		if len(content) != item.partSize {
			return 0, fmt.Errorf("part file read size error")
		}
		n, _ := f.Write(content)
		mergeCnt += n
	}
	return mergeCnt, nil
}

func (u *uploader) partValid(key, uploadId string, extra CompleteExtra) bool {
	//key是否存在
	if !u.partInfo.Exist(key) {
		return false
	}
	if !u.keyUploadIdMatch(key, uploadId) {
		return false
	}
	//校验分片数量
	partInfo := u.partInfo.GetPart(key)
	if len(partInfo) != len(extra.Progress) {
		return false
	}
	return true
}

func (u *uploader) keyUploadIdMatch(key, uploadId string) bool {
	//key和uploadId一致
	initInfo := u.partInfo.GetInit(key)
	if uploadId == initInfo.UploadId {
		return true
	}
	return false
}

func genPartName(key string, partNum int64) string {
	return fmt.Sprintf("%s_%d", key, partNum)
}
