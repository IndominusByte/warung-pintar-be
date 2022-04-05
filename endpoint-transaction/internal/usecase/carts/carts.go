package carts

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	cartsentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/carts"
	"github.com/go-chi/jwtauth"
	"github.com/gomodule/redigo/redis"
	"gopkg.in/guregu/null.v4"

	"github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/constant"
	"github.com/creent-production/cdk-go/response"
	"github.com/creent-production/cdk-go/validation"
)

type CartsUsecase struct {
	cartsRepo    cartsRepo
	authRepo     authRepo
	productsRepo productsRepo
}

func NewCartsUsecase(cartRepo cartsRepo, authRepo authRepo, productRepo productsRepo) *CartsUsecase {
	return &CartsUsecase{
		cartsRepo:    cartRepo,
		authRepo:     authRepo,
		productsRepo: productRepo,
	}
}

func (uc *CartsUsecase) CreateUpdate(ctx context.Context, rw http.ResponseWriter,
	payload *cartsentity.JsonCreateUpdateSchema) {

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

	product, err := uc.productsRepo.GetProductById(ctx, payload.ProductId)
	if err != nil {
		response.WriteJSONResponse(rw, 404, nil, map[string]interface{}{
			constant.App: "Product not found.",
		})
		return
	}

	cartDb, err := uc.cartsRepo.GetCartByUserIdAndProductId(ctx, user.Id, product.Id)
	cartQty := payload.Qty

	if err == nil && payload.Operation == "create" {
		cartQty += cartDb.Qty
	}

	if cartQty > product.Stock {
		var msg string
		msg = "The amount you input exceeds the available stock."
		if err == nil {
			msg = fmt.Sprintf("This item has %d stock left and you already have %d in your basket.", product.Stock, cartDb.Qty)
		}
		response.WriteJSONResponse(rw, 400, nil, map[string]interface{}{
			constant.App: msg,
		})
		return
	}

	// update data if exists
	if err == nil {
		uc.cartsRepo.Update(ctx, &cartsentity.Cart{
			Id:        cartDb.Id,
			Notes:     null.StringFrom(payload.Notes),
			Qty:       cartQty,
			UserId:    user.Id,
			ProductId: product.Id,
		})
		response.WriteJSONResponse(rw, 200, nil, map[string]interface{}{
			constant.App: "Shopping cart successfully updated.",
		})
		return
	}

	// insert into db
	uc.cartsRepo.Insert(ctx, &cartsentity.Cart{
		Notes:     null.StringFrom(payload.Notes),
		Qty:       cartQty,
		UserId:    user.Id,
		ProductId: product.Id,
	})

	response.WriteJSONResponse(rw, 201, nil, map[string]interface{}{
		constant.App: "The product has been successfully added to the shopping cart.",
	})
}

func (uc *CartsUsecase) GetAll(ctx context.Context, rw http.ResponseWriter, payload *cartsentity.QueryParamAllCartSchema) {
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

	payload.UserId = user.Id
	results, _ := uc.cartsRepo.GetAllCarts(ctx, payload)

	response.WriteJSONResponse(rw, 200, results, nil)
}

func (uc *CartsUsecase) Delete(ctx context.Context, rw http.ResponseWriter, payload *cartsentity.JsonMultipleSchema) {
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

	// delete carts
	payload.UserId = user.Id
	uc.cartsRepo.Delete(ctx, payload)

	response.WriteJSONResponse(rw, 200, nil, map[string]interface{}{
		constant.App: fmt.Sprintf("%d items were removed.", len(payload.ListId)),
	})
}

func (uc *CartsUsecase) MoveToPayment(ctx context.Context, rw http.ResponseWriter,
	redisCli *redis.Pool, payload *cartsentity.JsonMultipleSchema) {

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

	// move to payment
	payload.UserId = user.Id
	uc.cartsRepo.MoveItemToPayment(ctx, redisCli, payload)

	response.WriteJSONResponse(rw, 200, nil, map[string]interface{}{
		constant.App: fmt.Sprintf("%d items successfully moved to the payment.", len(payload.ListId)),
	})
}

func (uc *CartsUsecase) ItemInPayment(ctx context.Context, rw http.ResponseWriter, redisCli *redis.Pool) {
	_, claims, _ := jwtauth.FromContext(ctx)
	sub, _ := strconv.Atoi(claims["sub"].(string))

	user, err := uc.authRepo.GetUserById(ctx, sub)
	if err != nil {
		response.WriteJSONResponse(rw, 401, nil, map[string]interface{}{
			constant.Header: constant.UserNotFound,
		})
		return
	}

	results, _ := uc.cartsRepo.ItemInPayment(ctx, redisCli, user.Id)

	response.WriteJSONResponse(rw, 200, results, nil)
}
