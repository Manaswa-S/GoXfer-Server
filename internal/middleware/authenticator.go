package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"goxfer/server/internal/auth"
	"goxfer/server/internal/consts/errs"
	"goxfer/server/internal/utils"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var CtxSetBucKey = "bucKey"

type Authenticator struct {
	auth         auth.Authenticator
	AllowedDrift int64 // in seconds
}

func NewAuthenticator(auth auth.Authenticator) *Authenticator {
	return &Authenticator{
		auth:         auth,
		AllowedDrift: 15,
	}
}

func (s *Authenticator) Authenticator() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		serverTS := time.Now().Unix()

		clientTs := ctx.Request.Header.Get("X-Timestamp")
		clientSessId := ctx.Request.Header.Get("X-Session-ID")
		clientReqSign := utils.DecodeBase64(ctx.Request.Header.Get("X-Req-Signature"))
		clientBodySign := utils.DecodeBase64(ctx.Request.Header.Get("X-Body-Signature"))

		clientTS, err := strconv.ParseInt(clientTs, 10, 64)
		if err != nil {
			fmt.Println("client timestamp not ok")
			ctx.AbortWithStatusJSON(http.StatusBadRequest, errs.Errorf{
				Type:      errs.ErrBadForm,
				Message:   "client timestamp not ok",
				ReturnRaw: true,
			})
			return
		}

		if serverTS-clientTS > s.AllowedDrift {
			fmt.Println("UNAUTHORIZED: drifted too long")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errs.Errorf{
				Type:      errs.ErrUnauthorized,
				Message:   "UNAUTHORIZED",
				ReturnRaw: true,
			})
			return
		}

		session, err := s.auth.ValidateSession(ctx, clientSessId)
		if err != nil {
			fmt.Println("session expired: ", err)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errs.Errorf{
				Type:      errs.ErrUnauthorized,
				Message:   "session expired",
				ReturnRaw: true,
			})
			return
		}

		meta := fmt.Sprintf("%s\n%s\n%s\n%s",
			ctx.Request.Method,
			ctx.Request.URL.Path,
			ctx.Request.URL.RawQuery,
			clientTs,
		)
		serverReqSign, err := hash([]byte(meta), session.SessionKey)
		if err != nil {
			fmt.Println("internal error")
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if !hmac.Equal(clientReqSign, serverReqSign) {
			fmt.Println("could not match")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errs.Errorf{
				Type:      errs.ErrUnauthorized,
				Message:   "could not match",
				ReturnRaw: true,
			})
			return
		}

		body, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			fmt.Println("internal error")
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
		serverBodyHash, err := hash(body, session.SessionKey)
		if err != nil {
			fmt.Println("internal error")
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if !hmac.Equal(clientBodySign, serverBodyHash) {
			fmt.Println("could not match")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errs.Errorf{
				Type:      errs.ErrUnauthorized,
				Message:   "could not match",
				ReturnRaw: true,
			})
			return
		}

		// Authorized

		// TODO: maybe this should directly be the internal buc_id instead ?
		ctx.Set(CtxSetBucKey, session.ClientID)
		fmt.Println(session.ClientID)
	}
}

func hash(data []byte, key []byte) ([]byte, error) {
	mac := hmac.New(sha256.New, key)
	if _, err := mac.Write(data); err != nil {
		return nil, err
	}
	return mac.Sum(nil), nil
}
