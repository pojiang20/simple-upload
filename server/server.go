package server

import (
	"github.com/gin-gonic/gin"
	"github.com/pojiang20/simple-upload/util"
	"net/http"
	"os"
	"path"
)

type Server struct {
	r *gin.Engine

	publiDirPath  string
	MaxUploadSize int64
	SecureToken   string
}

func NewServer(maxUploadSize int64, token string) *Server {
	absPath, _ := os.Getwd()
	publicDir, err := util.GenDir(absPath, "public")
	if err != nil {
		return nil
	}
	return &Server{
		r:             gin.New(),
		publiDirPath:  publicDir,
		MaxUploadSize: maxUploadSize,
		SecureToken:   token,
	}
}

func (s *Server) Run(Address string) error {
	r := s.r

	r.Static("/zenfs/public", s.publiDirPath)
	v1 := r.Group("/zenfs")
	{
		v1.GET("/", s.index)
		v1.POST("/upload", s.uploadHandler)

		//分片上传

	}
	return s.r.Run(Address)
}

func (s *Server) index(c *gin.Context) {
	c.JSON(http.StatusOK, "This is Zenfs")
}

func (s *Server) uploadHandler(c *gin.Context) {
	info, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}
	filePath := path.Join(s.publiDirPath, info.Filename)
	err = c.SaveUploadedFile(info, filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "upload failed："+err.Error())
	} else {
		c.JSON(http.StatusOK, "uploaded successfully")
	}
}
