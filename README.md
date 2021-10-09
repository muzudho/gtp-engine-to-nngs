# gtp-engine-to-nngs

Connect to NNGS from GTP engine.  

## Documents

* GTP engine to nngs
  * Set up
    * [on Windows](./doc/set-up-app-on-windows.md)
  * Run
    * [on Windows](./doc/run-app-on-windows.md)
  * Operate
    * [on Windows](./doc/operate-app-on-windows.md)

## 大会用メモ

### CgfGoBan で対局するなら

GTP 設定:  

`C:\Users\むずでょ\go\src\github.com\muzudho\kifuwarabe-gtp\kifuwarabe-gtp.exe --workdir C:\Users\むずでょ\go\src\github.com\muzudho\gtp-engine-to-nngs\workspace-uec`  

自分と、対局相手のログイン名を入れる必要がある。（大文字小文字を区別しない）  
ログイン前に、白黒を合わせる必要もある。  

### このプログラム(gtp-engine-to-nngs)で対局するなら

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

## その他

ゾンビ・ユーザーを蹴飛ばすとき(admin ユーザーで？)  

```shell
nuke <name>
```
