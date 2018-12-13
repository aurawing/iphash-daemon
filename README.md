# iphash-daemon

iphash-daemon是星际哈希的守护进程兼自动更新程序，程序运行后会自动驻留系统后台运行，同时开始定时向网站请求当前iphash的程序发布版本信息，发现有新版本后会自动下载程序包至本地解压，并停止旧版本程序，然后启动新版本程序。

升级文件位于`http://hash.iptokenmain.com/upgrade/iphash-${sys}-${arch}.json`，其中`${sys}`为操作系统类型（linux或windows），`${arch}`为硬件架构（amd64或arm64），比如amd64平台下linux版本的升级文件为`http://hash.iptokenmain.com/upgrade/iphash-linux-amd64.json`，不同平台的iphash-daemon会根据自身架构和系统获取对应的升级文件，升级文件内容如下：
```
{
  "version":"v0.01",
  "url":"http://hash.iptokenmain.com/download/iphash-linux-amd64-v0.01.tar.gz",
  "sha1":"00cabe77b5b9af7c1ee03acb5b07e573c27149e8"
}
```
为一个JSON文件，其中`version`为版本号，`url`为更新程序包的下载路径，`sha1`为更新程序包的SHA1摘要（用于正确性校验），修改此文件并部署好更新程序包即可实现节点端的自动更新。

更新程序包的位置可以根据升级文件中URL的值自由确定，目前放在`http://hash.iptokenmain.com/download/`下，更新程序包的命名建议遵循`iphash-${sys}-${arch}-${version}.tar.gz`，其中`${sys}`为操作系统类型（linux或windows），`${arch}`为硬件架构（amd64或arm64），`${version}`为更新包版本号。

更新程序包为`tar`打包的`gunzip`压缩文件（后缀`tar.gz`），解压后为一个文件夹，文件夹名与压缩包的文件名相同（去掉扩展名）。其内容如下：
```
iphash-linux-amd64-v0.01
├── install.sh
├── ipfs
├── ipfs-monitor
└── swarm.key
```
`iphash-daemon`会在下载解压后依次执行`ipfs init -> install.sh -> ipfs daemon -> ipfs-monitor`
当`ipfs`或`ipfs-monitor`进程关闭时`iphash-daemon`会自动重启进程。
停止`iphash-daemon`可以执行`iphash-damon -s stop`
