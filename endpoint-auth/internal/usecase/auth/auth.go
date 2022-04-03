package auth

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/IndominusByte/magicimage"
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
	"gopkg.in/guregu/null.v4"
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
			"/app/templates/email/EmailConfirm.html",
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
			"/app/templates/email/EmailConfirm.html",
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
		response.WriteJSONResponse(rw, 404, nil, map[string]interface{}{
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

	passwordReset, err := uc.authRepo.GetPasswordResetByEmail(ctx, user.Email)

	loc, _ := time.LoadLocation("Asia/Ujung_Pandang")
	now, format := time.Now().In(loc), "2006-01-02 15:04:05"

	nowF, _ := time.Parse(format, now.Format(format))
	ExpiredF, _ := time.Parse(format, passwordReset.ResendExpired.Format(format))

	if err == nil && !nowF.After(ExpiredF) {
		response.WriteJSONResponse(rw, 400, nil, map[string]interface{}{
			constant.App: "You can try 5 minute later.",
		})
		return
	}

	var resetId string
	if err != nil {
		// insert to db
		resetId = uc.authRepo.InsertPasswordReset(ctx, payload)
	} else {
		// update expired time
		resetId = passwordReset.Id
		// generate resend expired 5 minute again
		uc.authRepo.GeneratePasswordResetResendExpired(ctx, resetId)
	}

	q := queue.NewQueue(func(val interface{}) {
		m.SendEmail(
			[]string{},
			[]string{user.Email},
			"Reset Password",
			"dont-reply@example.com",
			"/app/templates/email/EmailResetPassword.html",
			struct{ Link string }{Link: fmt.Sprintf("http://localhost:3000/auth/password-reset/%s", resetId)},
		)
	}, 20)
	q.Push("send")

	response.WriteJSONResponse(rw, 200, nil, map[string]interface{}{
		constant.App: "We have sent a password reset link to your email.",
	})
}

func (uc *AuthUsecase) PasswordReset(ctx context.Context, rw http.ResponseWriter,
	token string, payload *authentity.JsonPasswordResetSchema) {

	if err := validation.StructValidate(payload); err != nil {
		response.WriteJSONResponse(rw, 422, nil, err)
		return
	}

	passwordReset, err := uc.authRepo.GetPasswordResetById(ctx, token)
	if err != nil {
		response.WriteJSONResponse(rw, 404, nil, map[string]interface{}{
			constant.App: "Token not found.",
		})
		return
	}

	user, err := uc.authRepo.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		response.WriteJSONResponse(rw, 404, nil, map[string]interface{}{
			constant.App: "We can't find a user with that e-mail address.",
		})
		return
	}

	if user.Email != passwordReset.Email {
		response.WriteJSONResponse(rw, 400, nil, map[string]interface{}{
			constant.App: "The password reset token is invalid.",
		})
		return
	}

	// action on db
	uc.authRepo.UpdateUser(ctx, &authentity.User{
		Id:       user.Id,
		Password: payload.Password,
	})
	uc.authRepo.DeletePasswordReset(ctx, passwordReset.Id)

	response.WriteJSONResponse(rw, 200, nil, map[string]interface{}{
		constant.App: "Successfully reset your password.",
	})
}

func (uc *AuthUsecase) UpdatePassword(ctx context.Context,
	rw http.ResponseWriter, payload *authentity.JsonUpdatePasswordSchema) {

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

	if !uc.authRepo.IsPasswordSameAsHash(ctx, []byte(user.Password), []byte(payload.OldPassword)) {
		response.WriteJSONResponse(rw, 400, nil, map[string]interface{}{
			constant.App: "Password does not match with our records.",
		})
		return
	}

	// update password
	uc.authRepo.UpdateUser(ctx, &authentity.User{
		Id:       user.Id,
		Password: payload.Password,
	})

	response.WriteJSONResponse(rw, 200, nil, map[string]interface{}{
		constant.App: "Success update your password.",
	})
}

func (uc *AuthUsecase) UpdateAvatar(ctx context.Context, rw http.ResponseWriter, file *multipart.Form) {
	magic := magicimage.New(file)
	if err := magic.ValidateSingleImage("avatar"); err != nil {
		response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
			"avatar": err.Error(),
		})
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

	// if user avatar not same as default, delete the old one
	if user.Avatar != "default.jpg" {
		magicimage.DeleteFolderAndFile(fmt.Sprintf("/app/static/avatars/%s", user.Avatar))
	}

	magic.SaveImages(260, 260, "/app/static/avatars", true)

	// update avatar
	uc.authRepo.UpdateUser(ctx, &authentity.User{
		Id:     user.Id,
		Avatar: magic.FileNames[0],
	})

	response.WriteJSONResponse(rw, 200, nil, map[string]interface{}{
		constant.App: "Success update avatar.",
	})
}

func (uc *AuthUsecase) UpdateAccount(ctx context.Context, rw http.ResponseWriter, payload *authentity.JsonUpdateAccountSchema) {
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

	if _, err := uc.authRepo.GetUserByPhone(ctx, payload.Phone); err == nil && user.Phone.String != payload.Phone {
		response.WriteJSONResponse(rw, 400, nil, map[string]interface{}{
			constant.App: fmt.Sprintf(constant.AlreadyTaken, "phone"),
		})
		return
	}

	// update account
	uc.authRepo.UpdateUser(ctx, &authentity.User{
		Id:       user.Id,
		Fullname: null.StringFrom(payload.Fullname),
		Phone:    null.StringFrom(payload.Phone),
		Address:  null.StringFrom(payload.Address),
	})

	response.WriteJSONResponse(rw, 200, nil, map[string]interface{}{
		constant.App: "Success updated your account.",
	})
}

func (uc *AuthUsecase) GetUser(ctx context.Context, rw http.ResponseWriter) {
	_, claims, _ := jwtauth.FromContext(ctx)
	sub, _ := strconv.Atoi(claims["sub"].(string))

	user, err := uc.authRepo.GetUserById(ctx, sub)
	if err != nil {
		response.WriteJSONResponse(rw, 401, nil, map[string]interface{}{
			constant.Header: constant.UserNotFound,
		})
		return
	}

	response.WriteJSONResponse(rw, 200, user, nil)
}
