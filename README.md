# nth

[north](https://north.rip)をターミナルから使うための非公式クライアントです。

> [!WARNING]
>
> - north公式のクライアントではありません。
> - ブラウザセッションを使う機能は、予告なく動作しなくなる可能性があります。

タイムライン、投稿、検索、通知、ブックマーク、メッセージの閲覧に対応しています。

## インストール

```bash
go install github.com/Hayao0819/nth
```

Go 1.26以上が必要です。

## 使い方

```bash
nth
```

初回起動時に、APIトークンまたはnorthへログイン済みのブラウザを設定します。
認証情報はシステムキーリングへ保存されます。`NORTH_API_KEY`が設定されている場合は、
キーリング内のAPIトークンより優先されます。

設定を変更する場合は`nth setup`を実行します。キー操作は`?`で確認できます。

## 開発

```bash
direnv allow
nix fmt
nix flake check
```

direnvを使わない場合は`nix develop`で開発環境に入れます。

## License

MIT
