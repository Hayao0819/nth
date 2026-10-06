# nth

[north](https://north.rip)をターミナルから使うための非公式クライアントです。

> [!WARNING]
>
> - north公式のクライアントではありません。
> - ブラウザセッションを使う機能は、予告なく動作しなくなる可能性があります。

タイムライン、投稿、検索、通知、ブックマーク、メッセージの閲覧に対応しています。

## インストール

Linux、macOS:

```sh
curl -fsSL https://raw.githubusercontent.com/Hayao0819/nth/main/install.sh | sh
```

Windows PowerShell:

```powershell
irm https://raw.githubusercontent.com/Hayao0819/nth/main/install.ps1 | iex
```

インストール先は`~/.local/bin`です。`NTH_INSTALL_DIR`で変更できます。

Goからインストールする場合:

```sh
go install github.com/Hayao0819/nth
```

Goからインストールする場合はGo 1.26以上が必要です。

## 使い方

```bash
nth
```

初回起動時には設定画面が表示されます。このときの設定画面は`nth setup`で再表示できます。

キー操作は`?`で確認できます。

APIの動作確認には`nth test`を使用できます。

## 開発

```bash
direnv allow
nix fmt
nix flake check
```

direnvを使わない場合は`nix develop`で開発環境に入れます。

## License

MIT
