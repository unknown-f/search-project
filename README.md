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

get("/hot")-请求热词热榜

post("/search/text")-搜索text

收藏夹更名：在添加右侧文本框输入新名字点更名

收藏夹添加链接：进入收藏夹，在添加右侧文本框输入数字1-5，点击添加，则可添加左侧链接
