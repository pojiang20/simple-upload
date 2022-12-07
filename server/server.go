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
	//fs *FileSystem

	publiDirPath  string
	MaxUploadSize int64
	SecureToken   string
}

func NewServer(maxUploadSize int64, token string) *Server {
	publicDir, err := getPublicDir()
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

func getPublicDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	publicDir := path.Join(dir, "/public")
	if _, err := os.Stat(publicDir); err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(publicDir, 0755)
			if err != nil {
				util.Zlog.Errorf("mkdir %s error: %v", dir, err)
				return "", err
			}
		}
	}
	return publicDir, nil
}

func (s *Server) Run(Address string) error {
	r := s.r

	r.Static("/zenfs/public", s.publiDirPath)
	v1 := r.Group("/zenfs")
	{
		v1.GET("/", indexPage)
		v1.POST("/upload", s.uploadHandler)
	}
	return s.r.Run(Address)
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
