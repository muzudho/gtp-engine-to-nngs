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
# 使っていないパッケージを、インストールのリストから削除するなら
# go mod tidy

# 自作のパッケージを更新(再インストール)したいなら
# go get -u all

go build
```

## Settings

### 大会サーバーと通信するなら

1. 例えば、このプロジェクトの中に 📂`workspace-b-cgfopen2021` といったようなフォルダーを作ってください。  
   黒番（`b`）と白番（`w`）でフォルダーを変えてください  
   （大会によっては常に黒番でログインすればいいものもあります）
2. その中に 📂`input`, 📂`output` フォルダ―を作ってください  
   また、 📄`output/connector/.gitkeep` ファイルを置いてください。
3. 他の 📂`workspace-*/input` フォルダ―の中から 📄`connector.conf.toml` ファイルと、 📄`engine.conf.toml` ファイルを  
   コピーして さきほど作った 📂`input` へ持ってきてください。  
   黒番と白番で内容が違うことに注意してください
4. 大会によって変わるのが以下の箇所です  
   1. 📄`connector.conf.toml` の `[Server]` テーブル。`Host` と `Port` を編集してください
   2. 📄`connector.conf.toml` の `[MatchApplication]` テーブル。`BoardSize`, `AvailableTimeMinutes` など間違えずに設定してください
   3. 📄`engine.conf.toml` の `[Profile]` テーブルの `Name`。大会指定のエンジン名を入力する必要があるかもしれません
5. 囲碁エンジンによって変わるのが 📄`connector.conf.toml` の `[User]` テーブルです。  
   `InterfaceType` は `GTP`、 `EngineCommand` には 思考エンジンの `.exe` ファイルへの絶対パスを記述してください
6. 以下のコマンドを打鍵して、ログインしてください

Example:  

```shell
# 黒番の例
gtp-engine-to-nngs --workdir C:/Users/むずでょ/go/src/github.com/muzudho/gtp-engine-to-nngs/workspace-b-cgfopen2021
```
