# search_project

> A Golang project

## Build Setup

``` bash
安装mongodb到本地，建立search_project数据库，并在该数据库下建立indextosource集合和keytoindex集合

安装gcc，目的是为了正常运行jieba分词库

第一次启动时去掉ReadCutAndWrite的注释，用于建立检索数据库，wukong50k_release.csv这个数据集大概需要2个小时建立数据库
```
用户请求：
get("/")-请求页面
get("/hot")-请求热榜
get("/search/text/:time/:text")-搜索text
get("/search/picture/:time/:picture")-搜图
get("/register/:time/:userName/:passWord")-注册用户名和密码
get("/login/:time/:userName/:passWord")-登录用户名密码，需要返回收藏夹
get("/drop/:time/:userName")-用户注销
