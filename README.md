# 自建拷贝漫画代理服务器
# Self-hosted CopyManga proxy server
## build
```
go build -o mangaproxy main.go
```

## setup
1. 将`mangaproxy.conf`放到nginx include配置目录下
2. 替换`mangaproxy.conf`里所有的`<your-server-name>`为你的域名
2. 将`mangaproxy.service`放到`/lib/systemd/system`目录下
3. 将`mangaproxy`添加可执行权限，放到`/usr/bin`目录下


## run
```
systemctl --now enable mangaproxy
service nginx reload
```


## 请与安卓版copymanga一起食用
[https://github.com/fumiama/copymanga](https://github.com/fumiama/copymanga)