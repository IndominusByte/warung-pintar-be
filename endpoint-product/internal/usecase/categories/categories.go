package categories

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/constant"
	categoriesentity "github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/entity/categories"
	"github.com/creent-production/cdk-go/response"
	"github.com/creent-production/cdk-go/validation"
	"github.com/go-chi/jwtauth"
)

type CategoriesUsecase struct {
	categoriesRepo categoriesRepo
	authRepo       authRepo
}

func NewCategoriesUsecase(categoryRepo categoriesRepo, authRepo authRepo) *CategoriesUsecase {
	return &CategoriesUsecase{
		categoriesRepo: categoryRepo,
		authRepo:       authRepo,
	}
}

func (uc *CategoriesUsecase) Create(ctx context.Context,
	rw http.ResponseWriter, payload *categoriesentity.JsonCreateUpdateSchema) {

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

	if user.Role != "admin" {
		response.WriteJSONResponse(rw, 401, nil, map[string]interface{}{
			constant.Header: fmt.Sprintf(constant.PrivilegesOnly, "admin"),
		})
		return
	}

	if _, err := uc.categoriesRepo.GetCategoryByName(ctx, payload.Name); err == nil {
		response.WriteJSONResponse(rw, 400, nil, map[string]interface{}{
			constant.App: fmt.Sprintf(constant.AlreadyTaken, "name"),
		})
		return
	}

	// save into database
	uc.categoriesRepo.Insert(ctx, payload)

	response.WriteJSONResponse(rw, 201, nil, map[string]interface{}{
		constant.App: "Successfully add a new category.",
	})
}

func (uc *CategoriesUsecase) GetAll(ctx context.Context,
	rw http.ResponseWriter, payload *categoriesentity.QueryParamAllCategorySchema) {

	if err := validation.StructValidate(payload); err != nil {
		response.WriteJSONResponse(rw, 422, nil, err)
		return
	}

	results, _ := uc.categoriesRepo.GetAllCategoryPaginate(ctx, payload)

	response.WriteJSONResponse(rw, 200, results, nil)
}

func (uc *CategoriesUsecase) GetById(ctx context.Context, rw http.ResponseWriter, categoryId int) {
	t, err := uc.categoriesRepo.GetCategoryById(ctx, categoryId)
	if err != nil {
		response.WriteJSONResponse(rw, 404, nil, map[string]interface{}{
			"_app": "Category not found.",
		})
		return
	}
	response.WriteJSONResponse(rw, 200, t, nil)
}

func (uc *CategoriesUsecase) Update(ctx context.Context,
	rw http.ResponseWriter, payload *categoriesentity.JsonCreateUpdateSchema, categoryId int) {

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

	if user.Role != "admin" {
		response.WriteJSONResponse(rw, 401, nil, map[string]interface{}{
			constant.Header: fmt.Sprintf(constant.PrivilegesOnly, "admin"),
		})
		return
	}

	category, err := uc.categoriesRepo.GetCategoryById(ctx, categoryId)
	if err != nil {
		response.WriteJSONResponse(rw, 404, nil, map[string]interface{}{
			constant.App: "Category not found.",
		})
		return
	}

	// check name duplicate
	if _, err := uc.categoriesRepo.GetCategoryByName(ctx, payload.Name); err == nil && category.Name != payload.Name {
		response.WriteJSONResponse(rw, 400, nil, map[string]interface{}{
			constant.App: fmt.Sprintf(constant.AlreadyTaken, "name"),
		})
		return
	}

	// update guardian
	payload.Id = category.Id
	uc.categoriesRepo.Update(ctx, payload)

	response.WriteJSONResponse(rw, 200, nil, map[string]interface{}{
		constant.App: "Successfully update the category.",
	})
}

func (uc *CategoriesUsecase) Delete(ctx context.Context, rw http.ResponseWriter, categoryId int) {
	_, claims, _ := jwtauth.FromContext(ctx)
	sub, _ := strconv.Atoi(claims["sub"].(string))

	user, err := uc.authRepo.GetUserById(ctx, sub)
	if err != nil {
		response.WriteJSONResponse(rw, 401, nil, map[string]interface{}{
			constant.Header: constant.UserNotFound,
		})
		return
	}

	if user.Role != "admin" {
		response.WriteJSONResponse(rw, 401, nil, map[string]interface{}{
			constant.Header: fmt.Sprintf(constant.PrivilegesOnly, "admin"),
		})
		return
	}

	category, err := uc.categoriesRepo.GetCategoryById(ctx, categoryId)
	if err != nil {
		response.WriteJSONResponse(rw, 404, nil, map[string]interface{}{
			constant.App: "Category not found.",
		})
		return
	}

	// delete category
	uc.categoriesRepo.Delete(ctx, category.Id)

	response.WriteJSONResponse(rw, 200, nil, map[string]interface{}{
		constant.App: "Successfully delete the category.",
	})
}
