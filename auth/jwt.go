package auth

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"time"

	"github.com/GenkiSugiyama/go_todo_app_2/clock"
	"github.com/GenkiSugiyama/go_todo_app_2/entity"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

const (
	RoleKey     = "role"
	UserNameKey = "user_name"
)

//go:embed cert/secret.pem
var rawPrivKey []byte

//go:embed cert/public.pem
var rawPubKey []byte

type JWTer struct {
	PrivateKey, PublicKey jwk.Key
	Store                 Store
	Clocker               clock.Clocker
}

//go:generate go run github.com/matryer/moq -out moq_test.go . Store
type Store interface {
	Save(ctx context.Context, key string, userID entity.UserID) error
	Load(ctx context.Context, key string) (entity.UserID, error)
}

func NewJWTer(s Store, c clock.Clocker) (*JWTer, error) {
	j := &JWTer{Store: s}
	privkey, err := parse(rawPrivKey)
	if err != nil {
		return nil, fmt.Errorf("failed in NewJWTer: private key: %w", err)
	}
	pubkey, err := parse(rawPubKey)
	if err != nil {
		fmt.Errorf("failed in NewJWTer: public key: %w", err)
	}
	j.PrivateKey = privkey
	j.PublicKey = pubkey
	j.Clocker = c
	return j, nil
}

func parse(rawKey []byte) (jwk.Key, error) {
	key, err := jwk.ParseKey(rawKey, jwk.WithPEM(true))
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (j *JWTer) GenerateToken(ctx context.Context, u entity.User) ([]byte, error) {
	// jwtを作成するためのビルダーパターンによるjwt作成
	// 「誰向けに」「どんな用途で」「いつまで有効か」を設定している
	tok, err := jwt.NewBuilder().
		JwtID(uuid.New().String()).                       // トークンごとの一意なIDを発行
		Issuer(`github.com/GenkiSugiyama/go_todo_app_2`). // jwt発行元の情報を明示
		Subject("access_token").                          // 発行したjwtの用途を明示「アクセストークン」
		IssuedAt(j.Clocker.Now()).                        // 発行時刻を設定
		Expiration(j.Clocker.Now().Add(30*time.Minute)).  // トークンの有効期限を設定（ここでは30分後に執行）
		Claim(RoleKey, u.Role).                           // jwtにおける「Claim」とは認証に必要な特定の情報
		Claim(UserNameKey, u.Name).                       // ここでは"role"クレームにユーザーのロール情報をセット、"name"クレームにユーザー名をセットしている
		Build()
	if err != nil {
		return nil, fmt.Errorf("GetToken: failed to build token: %w", err)
	}

	// 署名前jwtトークン内のjtiフィールドの値（トークンID）とユーザーIDの組み合わせを保存する
	if err := j.Store.Save(ctx, tok.JwtID(), u.ID); err != nil {
		return nil, err
	}

	// 秘密鍵で署名してjwtトークンを完成させ、そのトークンを呼び出し元に返却する
	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256, j.PrivateKey))
	if err != nil {
		return nil, err
	}
	return signed, nil
}

func (j *JWTer) GetToken(ctx context.Context, r *http.Request) (jwt.Token, error) {
	// HTTPリクエストをパースしてjwtトークンを取得する
	// その際公開鍵で署名検証しjwt.Tokenにパースしてjwtトークンを返している
	token, err := jwt.ParseRequest(r, jwt.WithKey(jwa.RS256, j.PublicKey), jwt.WithValidate(false))
	if err != nil {
		return nil, err
	}

	// jwtトークンの有効期限やストア照合を行いトークンが正しいものかを検証している
	if err := jwt.Validate(token, jwt.WithClock(j.Clocker)); err != nil {
		return nil, fmt.Errorf("GetToken: failed to validate token: %w", err)
	}
	if _, err := j.Store.Load(ctx, token.JwtID()); err != nil {
		return nil, fmt.Errorf("GetToken: %q expired: %w", token.JwtID(), err)
	}
	// 署名検証と期限検証・ストア照合に成功したjwtトークンを返却する
	return token, nil
}

// HTTPリクエストのコンテキストにユーザーIDやロール情報を含めるためのメソッド
func (j *JWTer) FillContext(r *http.Request) (*http.Request, error) {
	// 渡されたHTTPリクエストからjwtトークンを取得
	token, err := j.GetToken(r.Context(), r)
	if err != nil {
		return nil, err
	}
	// jwtトークンのjtiをキーにユーザーIDを取得
	uid, err := j.Store.Load(r.Context(), token.JwtID())
	if err != nil {
		return nil, err
	}
	// ユーザーIDやトークンに含まれるロール情報をこのメソッド内でコンテキストに追加する
	ctx := SetUserID(r.Context(), uid)
	ctx = SetRole(ctx, token)
	clone := r.Clone(ctx)
	return clone, nil
}

type userIDKey struct{}
type roleKey struct{}

func SetUserID(ctx context.Context, uid entity.UserID) context.Context {
	// context.WithValueで値を設定する際のkey情報はstruct{}で定義したDefined Typeを使用する
	return context.WithValue(ctx, userIDKey{}, uid)
}

func GetUserID(ctx context.Context) (entity.UserID, bool) {
	id, ok := ctx.Value(userIDKey{}).(entity.UserID)
	return id, ok
}

func SetRole(ctx context.Context, tok jwt.Token) context.Context {
	// GenerateTokenでのjwtトークン生成時に"role"クレームにユーザーのロール情報が追加されているのでそれを取得
	get, ok := tok.Get(RoleKey)
	if !ok {
		return context.WithValue(ctx, roleKey{}, "")
	}
	// jwtトークンからロール情報が取得できればコンテキストにセットする
	return context.WithValue(ctx, roleKey{}, get)
}

func GetRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(roleKey{}).(string)
	return role, ok
}

func IsAdmin(ctx context.Context) bool {
	role, ok := GetRole(ctx)
	if !ok {
		return false
	}
	return role == "admin"
}
