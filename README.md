# XSCAN
## 简介
XSCAN是基于Go语言开发的搜索引擎，完成了基础部分的全部要求包括文本搜索，相关搜索，文本信息上传，获取热点文档和热点关键词和进阶部分关于用户注册登录以及收藏夹管理的要求。XSCAN使用了go gin，gorm框架和Mysql，Mongodb，Redis数据库。通过Mysql管理用户数据和收藏夹，通过Mongodb存储源文本和索引，通过Redis对热点搜索数据和热榜进行缓存和管理。用户鉴权通过JWT完成。目前读取了大概200K图文对数据。

## 系统架构

![image](https://github.com/unknown-f/img/blob/239e3a962c176aa1e536dec8d33f881c70dcd24d/arch.png)
## 部署步骤

- 安装mongodb到本地，建立search_project数据库，并在该数据库下建立indextosource集合和keytoindex集合
- 安装redis；安装mysql，并修改config.ini中的相关配置
- 安装gcc，目的是为了正常运行jieba分词库
- 第一次启动时去掉ReadCutAndWrite的注释，用于建立检索数据库，wukong50k_release.csv这个数据集大概需要2个小时建立数据库

## 功能介绍
详细接口文档见： https://www.apifox.cn/apidoc/shared-423b7faa-ac68-4343-bd49-6cb069c6cb21 访问密码 : 2222 
- 搜索
  - 获取热点文档：从redis中获取被检索次数较多的热点文档。显示在界面右侧访问热榜中。
  - 获取热点关键词：从redis中获取被检索次数较多的热点关键词。搜索结果显示在搜索文本框中及其下方，可以直接点击访问。
  - 文本搜索：首先对搜索内容进行拆分，拆分出搜索目标和过滤词，并返回符合预期的搜索结果和相关搜索词。搜索目标和过滤词用' -'分隔。
  - 相关搜索：搜索后如果有相关搜索内容，出现在界面左下角。
  - 上传文本信息：上传新的文本信息，并对文本信息进行分词和倒排索引。在搜索文本框中输入内容并在尾部加上'&'即可点击按钮上传。
- 用户管理
  - 添加用户；登录；删除用户。界面右侧操作。
- 收藏夹
  - 添加收藏夹；查看收藏夹列表；编辑收藏夹；删除收藏夹。登录后显示收藏夹，编辑收藏夹在收藏夹名文本框中输入新名称，点击旧收藏夹后面更名按钮即可更名。
- 链接
  - 添加链接；删除链接。添加链接操作步骤：进入收藏夹，在收藏链接的序号文本框中输入序号（左侧搜索出结果的序号）点击添加。
