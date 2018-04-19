## 概要
[株式会社Moiの2019年度新卒採用用課題](https://saiyo2019.moi.st/) のためのサーバ。  
Websocket を介して FizzBuzzMoi の問題に答える。

## サーバーの起動方法

```
$ dep ensure
$ go run main.go
```

## クライアント側の操作方法

### 接続先
ws://localhost:12345/websocket/:YourID

### ルール説明
#### 1. 開始メッセージを送信する

```json
{
  "signal": "start"
}
```

#### 2. 出題メッセージに返信する

開始メッセージ送信後に次のような出題メッセージがサーバーから送信される。

```json
{
  "number": 15,
  "previous": null
}
```

ここで受け取った `number` に対して、次の表に従った回答メッセージを作成し送信する。

|number|answer|example|
|:-|:-|:-|
|3で割り切れる|Fizz|`number` が6のとき、`answer` は Fizz|
|5で割り切れる|Buzz|`number` が10のとき、`answer` は Buzz|
|7で割り切れる|Moi|`number` が14のとき、`answer` は Moi|
|上記のうち複数にあてはまる|`answer` は、Fizz, Buzz, Moi の順で、あてはまる文字列を結合する|`number` が35のとき、5で割り切れるかつ7で割り切れるので、`answer` は BuzzMoi|
|いずれにもあてはまらない|`number` の値をそのまま格納する|`number` が1のとき、`answer` は1|

たとえば `number` が21のとき、3で割り切れるかつ7で割り切れるので、回答メッセージは Fizz と Buzz を結合して次のようになる。

```json
{
  "answer": "FizzMoi"
}
```

また、たとえば `number` が2のとき、上記のいずれにおあてはまらないので、回答メッセージは次のようになる。

```json
{
  "answer": "2"
}
```

#### 3. 制限時間が経過するまで 2. を繰り返す
回答メッセージを送信するたびにサーバーから出題メッセージが送信されるので、2. の通り繰り返し回答メッセージを送信する。
ただし、回答メッセージに `answer` フィールドが含まれていない場合、その場でサーバーから終了メッセージ（後述）が送信されチャレンジは強制終了される。
また、出題メッセージの `previous` フィールドは、直前の回答メッセージが正解だったかどうかを `"success"` or `"failure"`（初回は `null`）で表している。

制限時間が経過すると、その後に一番最初にサーバーから送信されるメッセージは回答メッセージの代わりに次のような終了メッセージとなる。

```json
{
  "signal": "end",
  "result": "success",
  "score": 10,
  "message": "Challenge is successful! Your record is 10 / 10 👏"
}
```

また、サーバーから終了メッセージが送信されると同時にコネクションは close されチャレンジは終了する。制限時間以内に既定の回数以上正解すると、`result` フィールドは success となる。
