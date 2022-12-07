# simple-upload

### upload0.01
可以直接使用默认值启动进程，可以从返回日志中获取token
```shell
{"level":"info","ts":1670309567.1186619,"caller":"simple-upload/main.go:19","msg":"ip 127.0.0.1,port 23456,token cf4afce824ae9e7359f9,upload_limit 4194304,root /tmp/simpleUpload"}
```
#### 文件上传
```shell
☁  /tmp  curl -Ffile=@file.txt 'http://localhost:23456/upload?token=cf4afce824ae9e7359f9'
{"ok":true,"path":"/files/file"}%
```
#### 文件下载
```shell
☁  simpleUpload  curl 'http://localhost:23456/files/file.txt?token=3805c7b1b8abacb55e47'
aa
```
#### 问题
- [ ] http部分需要手写，考虑框架
- [ ] 没有测试
- [ ] 鉴权过于简单

### upload0.11
gin框架下的upload，实现了上传和下载，并有相应的测试。


