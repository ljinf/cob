package token

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/ljinfu/cob"
	"net/http"
	"time"
)

const JWTToken = "cob_token"

type JwtHandler struct {
	//jwt 算法
	Alg string
	//过期时间
	Timeout        time.Duration
	RefreshTimeout time.Duration
	//时间函数
	TimeFunc   func() time.Time
	Key        []byte
	RefreshKey string
	//私钥
	PrivateKey    string
	SendCookie    bool
	Authenticator func(ctx *cob.Context) (map[string]interface{}, error)

	Header      string
	AuthHandler func(ctx *cob.Context, err error)

	CookieName     string
	CookieMaxAge   int64
	CookieDomain   string
	SecureCookie   bool
	CookieHTTPOnly bool
}

type JwtResponse struct {
	Token        string
	RefreshToken string
}

//登录 用户认证（用户名密码） -> 用户id 将id生成jwt，并且保存到cookie或者进行返回
func (j *JwtHandler) LoginHandler(ctx *cob.Context) (*JwtResponse, error) {
	data, err := j.Authenticator(ctx)
	if err != nil {
		return nil, err
	}
	if j.Alg == "" {
		j.Alg = "HS256"
	}
	//header部分
	signingMethod := jwt.GetSigningMethod(j.Alg)
	token := jwt.New(signingMethod)
	//payload部分
	claims := token.Claims.(jwt.MapClaims)
	if data != nil {
		for key, value := range data {
			claims[key] = value
		}
	}
	if j.TimeFunc == nil {
		j.TimeFunc = func() time.Time {
			return time.Now()
		}
	}
	//过期时间
	expire := j.TimeFunc().Add(j.Timeout)
	claims["exp"] = expire.Unix()
	claims["iat"] = j.TimeFunc().Unix()
	var tokenString string
	var tokenErr error
	//secret 部分
	if j.usingPublicKeyAlg() {
		tokenString, tokenErr = token.SignedString(j.PrivateKey)
	} else {
		tokenString, tokenErr = token.SignedString(j.Key)
	}

	if tokenErr != nil {
		return nil, tokenErr
	}
	jr := &JwtResponse{
		Token: tokenString,
	}
	refreshToken, err := j.refreshToken(token)
	if err != nil {
		return nil, err
	}
	jr.RefreshToken = refreshToken

	if j.SendCookie {
		if j.CookieName == "" {
			j.CookieName = JWTToken
		}
		if j.CookieMaxAge == 0 {
			j.CookieMaxAge = expire.Unix() - j.TimeFunc().Unix()
		}
		ctx.SetCookie(j.CookieName, tokenString, int(j.CookieMaxAge), "/",
			j.CookieDomain, j.SecureCookie, j.CookieHTTPOnly)
	}
	return jr, nil
}

func (j *JwtHandler) usingPublicKeyAlg() bool {
	switch j.Alg {
	case "RS256", "RS512", "RS384":
		return true
	}
	return false
}

func (j *JwtHandler) refreshToken(token *jwt.Token) (string, error) {
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = j.TimeFunc().Add(j.RefreshTimeout).Unix()
	var tokenString string
	var tokenErr error
	//secret 部分
	if j.usingPublicKeyAlg() {
		tokenString, tokenErr = token.SignedString(j.PrivateKey)
	} else {
		tokenString, tokenErr = token.SignedString(j.Key)
	}

	if tokenErr != nil {
		return "", tokenErr
	}
	return tokenString, nil
}

func (j *JwtHandler) RefreshHandler(ctx *cob.Context) (*JwtResponse, error) {
	rToken, ok := ctx.Get(j.RefreshKey)
	if !ok {
		return nil, errors.New("refresh token is null")
	}
	if j.Alg == "" {
		j.Alg = "HS256"
	}
	//解析
	t, err := jwt.Parse(rToken.(string), func(token *jwt.Token) (i interface{}, err error) {
		if j.usingPublicKeyAlg() {
			return []byte(j.PrivateKey), nil
		}
		return j.Key, nil
	})
	if err != nil {
		return nil, err
	}
	//payload部分
	claims := t.Claims.(jwt.MapClaims)
	if j.TimeFunc == nil {
		j.TimeFunc = func() time.Time {
			return time.Now()
		}
	}
	//过期时间
	expire := j.TimeFunc().Add(j.Timeout)
	claims["exp"] = expire.Unix()
	claims["iat"] = j.TimeFunc().Unix()
	var tokenString string
	var tokenErr error
	//secret 部分
	if j.usingPublicKeyAlg() {
		tokenString, tokenErr = t.SignedString(j.PrivateKey)
	} else {
		tokenString, tokenErr = t.SignedString(j.Key)
	}

	if tokenErr != nil {
		return nil, tokenErr
	}
	jr := &JwtResponse{
		Token: tokenString,
	}
	refreshToken, err := j.refreshToken(t)
	if err != nil {
		return nil, err
	}
	jr.RefreshToken = refreshToken

	if j.SendCookie {
		if j.CookieName == "" {
			j.CookieName = JWTToken
		}
		if j.CookieMaxAge == 0 {
			j.CookieMaxAge = expire.Unix() - j.TimeFunc().Unix()
		}
		ctx.SetCookie(j.CookieName, tokenString, int(j.CookieMaxAge), "/",
			j.CookieDomain, j.SecureCookie, j.CookieHTTPOnly)
	}
	return jr, nil
}

//退出登录
func (j *JwtHandler) LogoutHandler(ctx *cob.Context) error {
	//清楚cookie即可
	if j.SendCookie {
		if j.CookieName == "" {
			j.CookieName = JWTToken
		}
		ctx.SetCookie(j.CookieName, "", -1, "/", j.CookieDomain, j.SecureCookie, j.CookieHTTPOnly)
		return nil
	}
	return nil
}

//认证中间件
func (j *JwtHandler) AuthInterceptor(next cob.HandleFunc) cob.HandleFunc {
	return func(ctx *cob.Context) {
		if j.Header == "" {
			j.Header = "Authorization"
		}
		token := ctx.Request.Header.Get(j.Header)
		if token == "" {
			if j.SendCookie {
				cookie, err := ctx.Request.Cookie(j.CookieName)
				if err != nil {
					j.AuthErrorHandler(ctx, err)
					return
				}
				token = cookie.String()
			}
		}

		if token == "" {
			j.AuthErrorHandler(ctx, errors.New("token is null"))
			return
		}
		//解析
		t, err := jwt.Parse(token, func(token *jwt.Token) (i interface{}, err error) {
			if j.usingPublicKeyAlg() {
				return []byte(j.PrivateKey), nil
			}
			return j.Key, nil
		})
		if err != nil {
			j.AuthErrorHandler(ctx, err)
			return
		}
		claims := t.Claims.(jwt.MapClaims)
		ctx.Set("jwt_claims", claims)
		next(ctx)
	}
}

func (j *JwtHandler) AuthErrorHandler(ctx *cob.Context, err error) {
	if j.AuthHandler == nil {
		ctx.Writer.WriteHeader(http.StatusUnauthorized)
	} else {
		j.AuthHandler(ctx, err)
	}
}
