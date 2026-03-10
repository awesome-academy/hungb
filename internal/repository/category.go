package repository

import (
	"context"
	"fmt"

	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/models"

	"gorm.io/gorm"
)

type CategoryRepo interface {
	FindAll(ctx context.Context) ([]models.Category, error)
	FindAllParents(ctx context.Context) ([]models.Category, error)
	FindByID(ctx context.Context, id uint) (*models.Category, error)
	FindBySlug(ctx context.Context, slug string) (*models.Category, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	ExistsBySlugExcluding(ctx context.Context, slug string, excludeID uint) (bool, error)
	CountByIDs(ctx context.Context, ids []uint) (int64, error)
	Create(ctx context.Context, cat *models.Category) error
	Update(ctx context.Context, cat *models.Category) error
	Delete(ctx context.Context, id uint) error
	HasTours(ctx context.Context, id uint) (bool, error)
	HasChildren(ctx context.Context, id uint) (bool, error)
	GetDescendantIDs(ctx context.Context, parentID uint) ([]uint, error)
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepo {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) FindAll(ctx context.Context) ([]models.Category, error) {
	var cats []models.Category
	if err := r.db.WithContext(ctx).
		Preload("Children", func(db *gorm.DB) *gorm.DB {
			return db.Order("name ASC")
		}).
		Preload("Parent").
		Order("name ASC").
		Find(&cats).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryFindAll, err)
	}
	return cats, nil
}

func (r *categoryRepository) FindAllParents(ctx context.Context) ([]models.Category, error) {
	var cats []models.Category
	if err := r.db.WithContext(ctx).
		Where("parent_id IS NULL").
		Preload("Children", func(db *gorm.DB) *gorm.DB {
			return db.Order("name ASC")
		}).
		Order("name ASC").
		Find(&cats).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryFindAllParents, err)
	}
	return cats, nil
}

func (r *categoryRepository) FindByID(ctx context.Context, id uint) (*models.Category, error) {
	var cat models.Category
	if err := r.db.WithContext(ctx).
		Preload("Parent").
		Preload("Children").
		First(&cat, id).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryFindByID, err)
	}
	return &cat, nil
}

func (r *categoryRepository) FindBySlug(ctx context.Context, slug string) (*models.Category, error) {
	var cat models.Category
	if err := r.db.WithContext(ctx).
		Where("slug = ?", slug).
		First(&cat).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryFindBySlug, err)
	}
	return &cat, nil
}

func (r *categoryRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Category{}).
		Where("slug = ?", slug).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryCheckSlugExists, err)
	}
	return count > 0, nil
}

func (r *categoryRepository) ExistsBySlugExcluding(ctx context.Context, slug string, excludeID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Category{}).
		Where("slug = ? AND id != ?", slug, excludeID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryCheckSlugExcluding, err)
	}
	return count > 0, nil
}

func (r *categoryRepository) CountByIDs(ctx context.Context, ids []uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Category{}).
		Where("id IN ?", ids).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryCountByIDs, err)
	}
	return count, nil
}

func (r *categoryRepository) Create(ctx context.Context, cat *models.Category) error {
	if err := r.db.WithContext(ctx).Create(cat).Error; err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryCreate, err)
	}
	return nil
}

func (r *categoryRepository) Update(ctx context.Context, cat *models.Category) error {
	if err := r.db.WithContext(ctx).Save(cat).Error; err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryUpdate, err)
	}
	return nil
}

func (r *categoryRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&models.Category{}, id).Error; err != nil {
		return fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryDelete, err)
	}
	return nil
}

func (r *categoryRepository) HasTours(ctx context.Context, id uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Table("tour_categories").
		Where("category_id = ?", id).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryHasTours, err)
	}
	return count > 0, nil
}

func (r *categoryRepository) HasChildren(ctx context.Context, id uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Category{}).
		Where("parent_id = ?", id).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryHasChildren, err)
	}
	return count > 0, nil
}

func (r *categoryRepository) GetDescendantIDs(ctx context.Context, parentID uint) ([]uint, error) {
	var ids []uint
	if err := r.collectDescendants(ctx, parentID, &ids); err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxCategoryGetDescendantIDs, err)
	}
	return ids, nil
}

func (r *categoryRepository) collectDescendants(ctx context.Context, parentID uint, ids *[]uint) error {
	var childIDs []uint
	if err := r.db.WithContext(ctx).Model(&models.Category{}).
		Where("parent_id = ?", parentID).
		Pluck("id", &childIDs).Error; err != nil {
		return err
	}
	for _, cid := range childIDs {
		*ids = append(*ids, cid)
		if err := r.collectDescendants(ctx, cid, ids); err != nil {
			return err
		}
	}
	return nil
}
