# MCP Greeter Server

このプロジェクトは、[Model Context Protocol (MCP)](https://modelcontextprotocol.io/) を使用したシンプルな挨拶サーバーです。名前を受け取り、挨拶メッセージを返す MCP ツールを提供します。

## 概要

MCP Greeter Server は、[mcp/go-sdk](https://github.com/modelcontextprotocol/go-sdk) を使用して実装された MCP サーバーです。以下の機能を提供します：

- **greet ツール**: 名前を受け取り、`Hi {name}` 形式の挨拶を返します

## 前提条件

以下のソフトウェアがインストールされている必要があります：

- [Docker](https://www.docker.com/) (コンテナ実行用)
- [Go](https://golang.org/) 1.24.5 以上 (ローカル開発用、オプション)

## サーバーの立て方

### 1. Docker イメージのビルド

```bash
docker build -t mcp-greeter .
```

### 2. サーバーの実行

```bash
docker run --rm -i mcp-greeter
```

サーバーは標準入力/出力を通じて MCP プロトコルで通信します。

## 動作確認方法

### MCP Inspector を使用した確認

[MCP Inspector](https://github.com/modelcontextprotocol/inspector) を使用して、サーバーの動作を確認できます。

1. MCP Inspector をインストール：

```bash
npx @modelcontextprotocol/inspector
```

2. サーバー設定で以下のコマンドを指定：

```bash
docker run --rm -i mcp-greeter
```

3. Inspector 上で以下の操作を実行：

   - サーバーに接続
   - `greet` ツールを選択
   - `name` パラメータに任意の名前（例：`Alice`）を入力
   - ツールを実行

4. 期待される結果：

```json
{
  "greeting": "Hi Alice"
}
```
