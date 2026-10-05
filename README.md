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

初回起動時には設定画面が表示されます。このときの設定画面は`nth setup`で再表示できます。

キー操作は`?`で確認できます。

## 開発

```bash
direnv allow
nix fmt
nix flake check
```

direnvを使わない場合は`nix develop`で開発環境に入れます。

## License

MIT
