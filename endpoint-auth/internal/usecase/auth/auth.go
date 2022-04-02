package auth

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/IndominusByte/warung-pintar-be/endpoint-auth/internal/config"
	"github.com/IndominusByte/warung-pintar-be/endpoint-auth/internal/constant"
	authentity "github.com/IndominusByte/warung-pintar-be/endpoint-auth/internal/entity/auth"
	"github.com/creent-production/cdk-go/auth"
	"github.com/creent-production/cdk-go/mail"
	"github.com/creent-production/cdk-go/queue"
	"github.com/creent-production/cdk-go/response"
	"github.com/creent-production/cdk-go/validation"
	"github.com/go-chi/jwtauth"
	"github.com/gomodule/redigo/redis"
)

type AuthUsecase struct {
	authRepo authRepo
}

func NewAuthUsecase(authRepo authRepo) *AuthUsecase {
	return &AuthUsecase{
		authRepo: authRepo,
	}
}

func (uc *AuthUsecase) Register(ctx context.Context, rw http.ResponseWriter,
	payload *authentity.JsonRegisterSchema, m *mail.Mail) {

	if err := validation.StructValidate(payload); err != nil {
		response.WriteJSONResponse(rw, 422, nil, err)
		return
	}

	_, err := uc.authRepo.GetUserByEmail(ctx, payload.Email)
	if err == nil {
		response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
			constant.App: fmt.Sprintf(constant.AlreadyTaken, "email"),
		})
		return
	}

	// save into db
	user_id := uc.authRepo.InsertUser(ctx, payload)
	confirm_id := uc.authRepo.InsertUserConfirm(ctx, user_id)

	q := queue.NewQueue(func(val interface{}) {
		m.SendEmail(
			[]string{},
			[]string{payload.Email},
			"Activated User",
			"dont-reply@example.com",
			"templates/email/EmailConfirm.html",
			struct{ Link string }{Link: fmt.Sprintf("http://localhost:3000/auth/confirm/%s", confirm_id)},
		)
	}, 20)
	q.Push("send")

	response.WriteJSONResponse(rw, 201, nil, map[string]interface{}{
		constant.App: "Check your email to activated user.",
	})
}

func (uc *AuthUsecase) UserConfirm(ctx context.Context, rw http.ResponseWriter, token string, cfg *config.Config) {
	confirm, err := uc.authRepo.GetUserConfirmById(ctx, token)
	if err != nil {
		response.WriteJSONResponse(rw, 404, nil, map[string]interface{}{
			constant.App: "Token not found.",
		})
		return
	}

	if !confirm.Activated {
		uc.authRepo.SetUserConfirmActivatedTrue(ctx, confirm.Id)
	}

	// create token
	accessToken := auth.GenerateAccessToken(&auth.AccessToken{Sub: strconv.Itoa(confirm.UserId), Exp: jwtauth.ExpireIn(cfg.JWT.AccessExpires), Fresh: true})
	refreshToken := auth.GenerateRefreshToken(&auth.RefreshToken{Sub: strconv.Itoa(confirm.UserId), Exp: jwtauth.ExpireIn(cfg.JWT.RefreshExpires)})

	response.WriteJSONResponse(rw, 200, map[string]interface{}{
		"access_token":  auth.NewJwtTokenRSA(cfg.JWT.PublicKey, cfg.JWT.PrivateKey, cfg.JWT.Algorithm, accessToken),
		"refresh_token": auth.NewJwtTokenRSA(cfg.JWT.PublicKey, cfg.JWT.PrivateKey, cfg.JWT.Algorithm, refreshToken),
	}, nil)
}

func (uc *AuthUsecase) ResendEmail(ctx context.Context, rw http.ResponseWriter,
	payload *authentity.JsonEmailSchema, m *mail.Mail) {

	if err := validation.StructValidate(payload); err != nil {
		response.WriteJSONResponse(rw, 422, nil, err)
		return
	}

	user, err := uc.authRepo.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		response.WriteJSONResponse(rw, 404, nil, map[string]interface{}{
			constant.App: "Email not found.",
		})
		return
	}

	confirm, _ := uc.authRepo.GetUserConfirmByUserId(ctx, user.Id)

	if confirm.Activated {
		response.WriteJSONResponse(rw, 400, nil, map[string]interface{}{
			constant.App: "Your account already activated.",
		})
		return
	}

	//init the loc
	loc, _ := time.LoadLocation("Asia/Ujung_Pandang")
	now, format := time.Now().In(loc), "2006-01-02 15:04:05"

	nowF, _ := time.Parse(format, now.Format(format))
	ExpiredF, _ := time.Parse(format, confirm.ResendExpired.Format(format))

	if !nowF.After(ExpiredF) {
		response.WriteJSONResponse(rw, 400, nil, map[string]interface{}{
			constant.App: "You can try 5 minute later.",
		})
		return
	}

	q := queue.NewQueue(func(val interface{}) {
		m.SendEmail(
			[]string{},
			[]string{user.Email},
			"Activated User",
			"dont-reply@example.com",
			"templates/email/EmailConfirm.html",
			struct{ Link string }{Link: fmt.Sprintf("http://localhost:3000/auth/confirm/%s", confirm.Id)},
		)
	}, 20)
	q.Push("send")

	// generate resend expired 5 minute again
	uc.authRepo.GenerateUserConfirmResendExpired(ctx, confirm.Id)

	response.WriteJSONResponse(rw, 200, nil, map[string]interface{}{
		constant.App: "Email confirmation has send.",
	})
}

func (uc *AuthUsecase) Login(ctx context.Context, rw http.ResponseWriter, payload *authentity.JsonLoginSchema, cfg *config.Config) {
	if err := validation.StructValidate(payload); err != nil {
		response.WriteJSONResponse(rw, 422, nil, err)
		return
	}

	user, err := uc.authRepo.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
			constant.App: "Invalid credential.",
		})
		return
	}

	if !uc.authRepo.IsPasswordSameAsHash(ctx, []byte(user.Password), []byte(payload.Password)) {
		response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
			constant.App: "Invalid credential.",
		})
		return
	}

	confirm, _ := uc.authRepo.GetUserConfirmByUserId(ctx, user.Id)

	if !confirm.Activated {
		response.WriteJSONResponse(rw, 400, nil, map[string]interface{}{
			constant.App: "Please check your email to activate your account.",
		})
		return
	}

	// create token
	accessToken := auth.GenerateAccessToken(&auth.AccessToken{Sub: strconv.Itoa(user.Id), Exp: jwtauth.ExpireIn(cfg.JWT.AccessExpires), Fresh: true})
	refreshToken := auth.GenerateRefreshToken(&auth.RefreshToken{Sub: strconv.Itoa(user.Id), Exp: jwtauth.ExpireIn(cfg.JWT.RefreshExpires)})

	response.WriteJSONResponse(rw, 200, map[string]interface{}{
		"access_token":  auth.NewJwtTokenRSA(cfg.JWT.PublicKey, cfg.JWT.PrivateKey, cfg.JWT.Algorithm, accessToken),
		"refresh_token": auth.NewJwtTokenRSA(cfg.JWT.PublicKey, cfg.JWT.PrivateKey, cfg.JWT.Algorithm, refreshToken),
	}, nil)
}

func (uc *AuthUsecase) FreshToken(ctx context.Context, rw http.ResponseWriter,
	payload *authentity.JsonPasswordOnlySchema, cfg *config.Config) {

	if err := validation.StructValidate(payload); err != nil {
		response.WriteJSONResponse(rw, 422, nil, err)
		return
	}

	_, claims, _ := jwtauth.FromContext(ctx)
	sub, _ := strconv.Atoi(claims["sub"].(string))

	user, err := uc.authRepo.GetUserById(ctx, sub)
	if err != nil {
		response.WriteJSONResponse(rw, 401, nil, map[string]interface{}{
			constant.Header: constant.UserNotFound,
		})
		return
	}

	if !uc.authRepo.IsPasswordSameAsHash(ctx, []byte(user.Password), []byte(payload.Password)) {
		response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
			constant.App: "Password does not match with our records.",
		})
		return
	}

	// create token
	accessToken := auth.GenerateAccessToken(&auth.AccessToken{Sub: strconv.Itoa(user.Id), Exp: jwtauth.ExpireIn(cfg.JWT.AccessExpires), Fresh: true})

	response.WriteJSONResponse(rw, 200, map[string]interface{}{
		"access_token": auth.NewJwtTokenRSA(cfg.JWT.PublicKey, cfg.JWT.PrivateKey, cfg.JWT.Algorithm, accessToken),
	}, nil)
}

func (uc *AuthUsecase) RefreshToken(ctx context.Context, rw http.ResponseWriter, cfg *config.Config) {
	_, claims, _ := jwtauth.FromContext(ctx)
	sub, _ := strconv.Atoi(claims["sub"].(string))

	user, err := uc.authRepo.GetUserById(ctx, sub)
	if err != nil {
		response.WriteJSONResponse(rw, 401, nil, map[string]interface{}{
			constant.Header: constant.UserNotFound,
		})
		return
	}

	// create token
	accessToken := auth.GenerateAccessToken(&auth.AccessToken{Sub: strconv.Itoa(user.Id), Exp: jwtauth.ExpireIn(cfg.JWT.AccessExpires)})

	response.WriteJSONResponse(rw, 200, map[string]interface{}{
		"access_token": auth.NewJwtTokenRSA(cfg.JWT.PublicKey, cfg.JWT.PrivateKey, cfg.JWT.Algorithm, accessToken),
	}, nil)
}

func (uc *AuthUsecase) AccessRevoke(ctx context.Context, rw http.ResponseWriter, redisCli *redis.Pool, cfg *config.Config) {
	conn := redisCli.Get()
	defer conn.Close()

	_, claims, _ := jwtauth.FromContext(ctx)
	conn.Do("SETEX", claims["jti"], cfg.JWT.AccessExpires.Seconds(), "ok")

	response.WriteJSONResponse(rw, 200, nil, map[string]interface{}{
		constant.App: "An access token has revoked.",
	})
}

func (uc *AuthUsecase) RefreshRevoke(ctx context.Context, rw http.ResponseWriter, redisCli *redis.Pool, cfg *config.Config) {
	conn := redisCli.Get()
	defer conn.Close()

	_, claims, _ := jwtauth.FromContext(ctx)
	conn.Do("SETEX", claims["jti"], cfg.JWT.RefreshExpires.Seconds(), "ok")

	response.WriteJSONResponse(rw, 200, nil, map[string]interface{}{
		constant.App: "An refresh token has revoked.",
	})
}

func (uc *AuthUsecase) PasswordResetSend(ctx context.Context,
	rw http.ResponseWriter, payload *authentity.JsonEmailSchema, m *mail.Mail) {

	if err := validation.StructValidate(payload); err != nil {
		response.WriteJSONResponse(rw, 422, nil, err)
		return
	}

	user, err := uc.authRepo.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		response.WriteJSONResponse(rw, 400, nil, map[string]interface{}{
			constant.App: "We can't find a user with that e-mail address.",
		})
		return
	}

	confirm, _ := uc.authRepo.GetUserConfirmByUserId(ctx, user.Id)

	if !confirm.Activated {
		response.WriteJSONResponse(rw, 400, nil, map[string]interface{}{
			constant.App: "Please activate your account first.",
		})
		return
	}
}
