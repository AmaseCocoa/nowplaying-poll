# nowplaying-poll

シンプルなListenBrainzの「Now Playing」監視とMastodonへの自動投稿ツール。

- ポーリング対象: ListenBrainz の playing-now API（ユーザー単位）
- キャッシュ: BoltDB (my.db) に直近の再生情報とMastodonの資格情報を保存
- 投稿先: Mastodonインスタンス（初回はOAuth認証でアクセストークンを登録）

## 必要環境
- Go 1.26.x
- ネットワーク接続

## ビルド

```bash
go build -o nowplaying-poll
```

## 実行

環境変数は .env ファイルに置くのが簡単です。必須/任意のキー:

- LISTENBRAINZ_USER (必須): ListenBrainz のユーザー名
- MASTODON_INSTANCE (必須): 投稿先 Mastodon インスタンス URL (例: https://mastodon.example)
- SOCIAL_SENDER_FORMAT (任意): 投稿テンプレート（Go の text/template 構文）。デフォルト: `{{.Track}} - {{.Artist}} ({{.Album}})\n#NowPlaying`
- LISTENBRAINZ_TOKEN (任意): ListenBrainz の API トークン（必要な場合）

初回に Mastodon のクレデンシャルが保存されていない場合、アプリが OAuth の登録 URL を表示し、認証コードの入力を促します。アクセストークン等は my.db の BoltDB バケット `MastodonCredentials` に保存されます。

実行例:

```bash
# .env を用意してから
./nowplaying-poll
```

## 主なファイル
- main.go: ポーリング、テンプレート、投稿ロジック
- mstdn.go: Mastodon クレデンシャル管理とトークン取得
- utils.go: BoltDB の読み書き、シークレット入力補助
- my.db: BoltDB データファイル（実行時に生成/更新）

## 注意
- 本ツールは定期的に ListenBrainz の playing-now エンドポイントを呼び出します。過度なリクエストは避けてください。
- my.db にアクセストークンを平文で保存します。取り扱いに注意してください。

## ライセンス
リポジトリのライセンスに従ってください。

