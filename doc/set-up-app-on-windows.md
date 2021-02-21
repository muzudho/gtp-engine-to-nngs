# Set up app on Windows

## Pre-install

Pre-install:  

```shell
go mod init

go get github.com/pelletier/go-toml
```

Telnet:  

```shell
# Go言語 は 個人作成の同名のライブラリがいっぱいあるので 一番良さそうなのを検索してください。
go get -v -u github.com/reiver/go-telnet
```

## Install

Build:  

```shell
# 自作のパッケージを更新(再インストール)したいなら
# go get -u all

go build
```
