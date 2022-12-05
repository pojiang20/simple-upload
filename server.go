package simple_upload

import "net/http"

type Server struct {
	DocumentRoot  string
	MaxUploadSize int64
}

func NewServer() (*Server, error) {

}

// 下载文件
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
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

}