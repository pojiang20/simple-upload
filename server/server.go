package server

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pojiang20/simple-upload/util"
	"net/http"
	"os"
	"path"
	"strconv"
)

type Server struct {
	r  *gin.Engine
	up *uploader

	publicDirPath string
	//MaxUploadSize int64
	//SecureToken   string
}

func NewServer(maxUploadSize int64, token string, dbAddr string) *Server {
	absPath, _ := os.Getwd()
	publicDir, err := util.GenDir(absPath, "public")
	if err != nil {
		util.Zlog.Errorf("NewServer init error %v", err)
		return nil
	}
	up, err := NewUploader(publicDir, dbAddr)
	if err != nil {
		util.Zlog.Errorf("NewServer init error %v", err)
		return nil
	}
	return &Server{
		r:             gin.New(),
		up:            up,
		publicDirPath: publicDir,
		//MaxUploadSize: maxUploadSize,
		//SecureToken:   token,
	}
}

func (s *Server) Run(Address string) error {
	r := s.r

	r.Static("/zenfs/public", s.publicDirPath)
	v1 := r.Group("/zenfs")
	{
		v1.GET("/", s.index)
		v1.POST("/upload", s.OneTimeUpload)

		//分片上传
		v1.POST("/uploads/init", s.initPart)
		v1.POST("/uploads/uploadPart", s.uploadPart)
		v1.POST("/uploads/complete", s.complete)
	}
	return s.r.Run(Address)
}

func (s *Server) index(c *gin.Context) {
	c.JSON(http.StatusOK, "Welcome! This is Zenfs")
}

func (s *Server) OneTimeUpload(c *gin.Context) {
	info, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}
	filePath := path.Join(s.publicDirPath, info.Filename)
	err = c.SaveUploadedFile(info, filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "upload failed："+err.Error())
	} else {
		c.JSON(http.StatusOK, "uploaded successfully")
	}
}

func (s *Server) initPart(c *gin.Context) {
	key := c.Query("key")
	var msg respMsg
	if len(key) == 0 {
		msg.Msg = fmt.Sprintf("initPart error: %v", util.ErrKeyNotExist)
		c.JSON(http.StatusInternalServerError, msg)
		return
	}
	info, err := s.up.Init(key)
	if err != nil {
		msg.Msg = fmt.Sprintf("initPart error: %v", err)
		c.JSON(http.StatusInternalServerError, msg)
		return
	}
	msg.Msg = "initPart success"
	msg.Info = info
	c.JSON(http.StatusOK, msg)
	return
}

// PUT /uploads/<UploadId>/<PartNumber> HTTP/1.1
func (s *Server) uploadPart(c *gin.Context) {
	var msg respMsg
	partSize := c.Request.ContentLength
	if partSize == -1 {
		msg.Msg = "invalid part size"
		c.JSON(http.StatusBadRequest, msg)
	}
	key := c.Query("key")
	uploadId := c.Query("UploadId")
	partNum, _ := strconv.Atoi(c.Query("PartNumber"))
	etag, fileSize, err := s.up.UploadPart(c.Request.Body, key, partNum, uploadId)
	if err != nil {
		msg.Msg = "upload part failed, error " + err.Error()
		c.JSON(http.StatusBadRequest, msg)
		return
	}
	msg.Msg = "uploadPart success"
	msg.Info = UploadInfo{
		Key:      key,
		UploadId: uploadId,
		Etags:    []UploadPartInfo{{Etag: etag, PartNumber: partNum, PartSize: fileSize}},
	}
	c.JSON(http.StatusOK, msg)
	return
}

func (s *Server) complete(c *gin.Context) {
	var msg respMsg
	key := c.Query("key")
	uploadId := c.Query("uploadId")
	progress := []UploadPartInfo{}
	err := json.NewDecoder(c.Request.Body).Decode(&progress)
	if err != nil {
		msg.Msg = "complete failed, error " + err.Error()
		c.JSON(http.StatusBadRequest, msg)
		return
	}
	err = s.up.Complete(key, uploadId, CompleteExtra{progress})
	if err != nil {
		msg.Msg = "complete failed, error " + err.Error()
		c.JSON(http.StatusBadRequest, msg)
		return
	}
	msg.Msg = "complete success"
	c.JSON(http.StatusOK, msg)
	return
}

type respMsg struct {
	Msg  string      `json:"msg"`
	Info interface{} `json:"info"`
}
