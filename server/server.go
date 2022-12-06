package main

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/pojiang20/simple-upload/util"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
)

var (
	//仅匹配"/upload"，前后内容
	rePathUpload = regexp.MustCompile(`^/upload$`)
	//匹配以/files/为前缀的路径
	rePathFiles = regexp.MustCompile("^/files/([^/]+)$")
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
		v1.POST("/upload/", s.uploadHandler)
	}
	return s.r.Run(Address)
}

func (s *Server) uploadHandler(c *gin.Context) {
	info, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}
	c.SaveUploadedFile(info, s.publiDirPath)
}

func (s *Server) handlePut(w http.ResponseWriter, r *http.Request) {
	//FindStringSubmatchIndex首先匹配正则，然后匹配其()的子表达式
	//这里的结果是解析出表达式路径及其子路径
	matches := rePathFiles.FindStringSubmatch(r.URL.Path)
	if matches == nil || len(matches) < 1 {
		util.Zlog.Error("invalid path")
		util.WriteError(w, http.StatusNotFound, fmt.Errorf("\"%s\" is not found", r.URL.Path))
	}
	targetPath := path.Join(s.DocumentRoot, matches[1])
	file, err := os.OpenFile(targetPath, os.O_RDONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {

	}
	defer file.Close()
	defer r.Body.Close()

	srcFile, info, err := r.FormFile("file")
	if err != nil {
		util.Zlog.Error("Failed to acquire the uploaded content")
		util.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	defer srcFile.Close()
	util.Zlog.Debug(info.Header)

	size, err := util.GetSize(srcFile)
	if err != nil {
		util.Zlog.Error("Failed to get the size of the uploaded content")
		util.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	if size > s.MaxUploadSize {
		util.Zlog.Error("File size exceeded")
		util.WriteError(w, http.StatusRequestEntityTooLarge, errors.New("uploaded file size exceeds the limit"))
		return
	}

	n, err := io.Copy(file, srcFile)
	if err != nil {
		util.Zlog.Error("Filaed to write body to the file")
		util.WriteError(w, http.StatusInternalServerError, err)
	}
	util.Zlog.Infof("file uploaded by PUT,file size :%d", n)
	util.WriteSuccess(w, targetPath)
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		token = r.Form.Get("token")
	}
	if token != s.SecureToken {
		util.WriteError(w, http.StatusUnauthorized, fmt.Errorf("authentication required"))
		return
	}

	switch r.Method {
	case http.MethodGet:
		fallthrough
	case http.MethodHead:
		s.handleGet(w, r)
	case http.MethodPost:
		s.handlePost(w, r)
	case http.MethodPut:
		s.handlePut(w, r)
	default:
		w.Header().Add("Allow", "GET,HEAD,POST,PUT")
		util.WriteError(w, http.StatusMethodNotAllowed, fmt.Errorf("method \"%s\" is not allowed", r.Method))
	}
}
