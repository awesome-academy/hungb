package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/models"
	"sun-booking-tours/internal/repository"
	"sun-booking-tours/internal/utils"

	"gorm.io/gorm"
)

type CategoryForm struct {
	Name        string `form:"name" binding:"required,max=255"`
	Description string `form:"description"`
	ParentID    uint   `form:"parent_id"`
}

type CategoryTree struct {
	models.Category
	Children []models.Category
}

type CategoryService struct {
	repo repository.CategoryRepo
}

func NewCategoryService(repo repository.CategoryRepo) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) ListCategories(ctx context.Context) ([]CategoryTree, error) {
	all, err := s.repo.FindAllParents(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryServiceListCategories, err)
	}

	trees := make([]CategoryTree, 0, len(all))
	for _, cat := range all {
		trees = append(trees, CategoryTree{
			Category: cat,
			Children: cat.Children,
		})
	}
	return trees, nil
}

func (s *CategoryService) AllFlatCategories(ctx context.Context) ([]models.Category, error) {
	cats, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryServiceAllFlat, err)
	}
	return cats, nil
}

func (s *CategoryService) GetCategory(ctx context.Context, id uint) (*models.Category, error) {
	cat, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryServiceGetCategory, err)
	}
	return cat, nil
}

func (s *CategoryService) CreateCategory(ctx context.Context, form *CategoryForm) error {
	name := strings.TrimSpace(form.Name)
	if name == "" {
		return appErrors.NewAppError(400, appErrors.ErrMsgCategoryNameRequired)
	}

	slug := utils.Slugify(name)

	exists, err := s.repo.ExistsBySlug(ctx, slug)
	if err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryServiceCreateCheckSlug, err)
	}
	if exists {
		return appErrors.NewAppError(409, appErrors.ErrMsgCategoryNameDuplicate)
	}

	var parentID *uint
	if form.ParentID > 0 {
		parent, pErr := s.repo.FindByID(ctx, form.ParentID)
		if pErr != nil {
			return appErrors.NewAppError(400, appErrors.ErrMsgCategoryParentNotFound)
		}
		if parent.ParentID != nil {
			return appErrors.NewAppError(400, appErrors.ErrMsgCategoryParentMustBeRoot)
		}
		parentID = &form.ParentID
	}

	cat := models.Category{
		Name:        name,
		Slug:        slug,
		Description: strings.TrimSpace(form.Description),
		ParentID:    parentID,
	}

	if err := s.repo.Create(ctx, &cat); err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryServiceCreate, err)
	}
	return nil
}

func (s *CategoryService) UpdateCategory(ctx context.Context, id uint, form *CategoryForm) error {
	cat, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrCategoryNotFound
		}
		return fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryServiceUpdateFindCategory, err)
	}

	name := strings.TrimSpace(form.Name)
	if name == "" {
		return appErrors.NewAppError(400, appErrors.ErrMsgCategoryNameRequired)
	}

	slug := utils.Slugify(name)

	slugExists, err := s.repo.ExistsBySlugExcluding(ctx, slug, id)
	if err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryServiceUpdateCheckSlug, err)
	}
	if slugExists {
		return appErrors.NewAppError(409, appErrors.ErrMsgCategoryNameDuplicate)
	}

	var parentID *uint
	if form.ParentID > 0 {
		if form.ParentID == id {
			return appErrors.NewAppError(400, appErrors.ErrMsgCategorySelfParent)
		}

		descendants, dErr := s.repo.GetDescendantIDs(ctx, id)
		if dErr != nil {
			return fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryServiceUpdateGetDescendants, dErr)
		}
		for _, did := range descendants {
			if did == form.ParentID {
				return appErrors.NewAppError(400, appErrors.ErrMsgCategoryChildAsParent)
			}
		}

		parent, pErr := s.repo.FindByID(ctx, form.ParentID)
		if pErr != nil {
			return appErrors.NewAppError(400, appErrors.ErrMsgCategoryParentNotFound)
		}
		if parent.ParentID != nil {
			return appErrors.NewAppError(400, appErrors.ErrMsgCategoryParentMustBeRoot)
		}
		parentID = &form.ParentID
	}

	cat.Name = name
	cat.Slug = slug
	cat.Description = strings.TrimSpace(form.Description)
	cat.ParentID = parentID

	if err := s.repo.Update(ctx, cat); err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryServiceUpdate, err)
	}
	return nil
}

func (s *CategoryService) DeleteCategory(ctx context.Context, id uint) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrCategoryNotFound
		}
		return fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryServiceDeleteFindCategory, err)
	}

	hasChildren, err := s.repo.HasChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryServiceDeleteCheckChildren, err)
	}
	if hasChildren {
		return appErrors.NewAppError(400, appErrors.ErrMsgCategoryCannotDeleteWithChildren)
	}

	hasTours, err := s.repo.HasTours(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryServiceDeleteCheckTours, err)
	}
	if hasTours {
		return appErrors.ErrCategoryHasTours
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryServiceDelete, err)
	}
	return nil
}
