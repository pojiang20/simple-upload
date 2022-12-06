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
	DocumentRoot  string
	MaxUploadSize int64
	SecureToken   string
}

func NewServer(documentRoot string, maxUploadSize int64, token string) Server {
	return Server{
		documentRoot,
		maxUploadSize,
		token,
	}
}

// 下载文件
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	if !rePathFiles.MatchString(r.URL.Path) {
		util.WriteError(w, http.StatusNotFound, fmt.Errorf("\"%s\" is not found", r.URL.Path))
		return
	}
	/*
		 FileServer 已经明确静态文件的根目录在"/tmp"，但是我们希望URL以"/tmpfiles/"开头。
		如果有人请求"/tempfiles/example.txt"，我们希望服务器能将文件发送给他。
		为了达到这个目的，我们必须从URL中过滤掉"/tmpfiles", 而剩下的路径是相对于根目录"/tmp"的相对路径。
	*/
	http.StripPrefix("/files/", http.FileServer(http.Dir(s.DocumentRoot))).ServeHTTP(w, r)
}

// 表单上传
func (s *Server) handlePost(w http.ResponseWriter, r *http.Request) {
	srcFile, info, err := r.FormFile("file")
	if err != nil {
		util.Zlog.Error("Failed to acquire the uploaded content")
		util.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	defer srcFile.Close()

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

	body, err := io.ReadAll(srcFile)
	if err != nil {
		util.Zlog.Error("Failed to read the uploaded content")
		util.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	filename := info.Filename
	if filename == "" {
		filename = fmt.Sprintf("%x", sha1.Sum(body))
	}

	dstPath := path.Join(s.DocumentRoot, filename)
	dstFile, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		util.Zlog.Error("failed to open the file")
		util.WriteError(w, http.StatusInternalServerError, err)
		return
	}
	defer dstFile.Close()
	if written, err := dstFile.Write(body); err != nil {
		util.WriteError(w, http.StatusInternalServerError, err)
		return
	} else if int64(written) != size {
		util.Zlog.Error("uploaded file size and written size differ")
		util.WriteError(w, http.StatusInternalServerError, fmt.Errorf("the size of uploaded content is %d, but %d bytes written", size, written))
	}
	util.Zlog.Info("file uploaded by Post")
	util.WriteSuccess(w, util.GetUrl(dstPath, s.DocumentRoot))
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
