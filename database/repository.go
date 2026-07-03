package database

import (
	"context"

	"github.com/google/uuid"
	"github.com/nguereza-tony/corekit/common"
	"gorm.io/gorm"
)

type FilterFunc func(*gorm.DB) *gorm.DB
type FilterDefinition struct {
	Column    string
	Operator  string
	Transform func(any) any
	Apply     func(
		query *gorm.DB,
		value any,
	) *gorm.DB
}

type Repository[T any] interface {
	Create(
		ctx context.Context,
		entity *T,
	) error

	Find(
		ctx context.Context,
		id uuid.UUID,
	) (*T, error)

	FindBy(
		ctx context.Context,
		where map[string]any,
	) (*T, error)

	FindAllBy(
		ctx context.Context,
		where map[string]any,
	) ([]T, error)

	List(
		ctx context.Context,
		opts common.QueryOptions,
	) (*common.PageResult[T], error)

	Count(
		ctx context.Context,
		where map[string]any,
	) (int64, error)

	Exists(
		ctx context.Context,
		where map[string]any,
	) (bool, error)

	Update(
		ctx context.Context,
		entity *T,
	) error

	UpdateFields(
		ctx context.Context,
		id uuid.UUID,
		fields map[string]any,
	) error

	UpdateBy(
		ctx context.Context,
		where map[string]any,
		fields map[string]any,
	) error

	Delete(
		ctx context.Context,
		id uuid.UUID,
	) error

	DeleteBy(
		ctx context.Context,
		where map[string]any,
	) error
}

type BaseRepository[T any] struct {
	db        *gorm.DB
	relations []ModelRelation
	filters   map[string]FilterDefinition
}

func NewBaseRepository[T any](
	db *gorm.DB,
	filters map[string]FilterDefinition,
) *BaseRepository[T] {

	return &BaseRepository[T]{
		db:      db,
		filters: filters,
	}
}

func (r *BaseRepository[T]) Create(
	ctx context.Context,
	entity *T,
) error {
	return r.db.WithContext(ctx).
		Create(entity).
		Error
}

func (r *BaseRepository[T]) Update(
	ctx context.Context,
	entity *T,
) error {
	return r.db.WithContext(ctx).
		Save(entity).
		Error
}

func (r *BaseRepository[T]) UpdateFields(
	ctx context.Context,
	id uuid.UUID,
	fields map[string]any,
) error {

	var entity T

	return r.db.WithContext(ctx).
		Model(&entity).
		Where("id = ?", id).
		Updates(fields).
		Error
}

func (r *BaseRepository[T]) UpdateBy(
	ctx context.Context,
	where map[string]any,
	fields map[string]any,
) error {
	var entity T
	query := r.db.WithContext(ctx)

	query = r.applyWhere(
		query,
		where,
	)

	return query.Model(&entity).Updates(fields).Error
}

func (r *BaseRepository[T]) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	var entity T

	return r.db.WithContext(ctx).
		Delete(&entity, "id = ?", id).
		Error
}

func (r *BaseRepository[T]) DeleteBy(
	ctx context.Context,
	where map[string]any,
) error {
	var entity T

	query := r.db.WithContext(ctx)

	query = r.applyWhere(
		query,
		where,
	)

	return query.Delete(&entity).Error
}

func (r *BaseRepository[T]) Find(
	ctx context.Context,
	id uuid.UUID,
) (*T, error) {

	var entity T

	query := r.db.WithContext(ctx)

	query = r.applyRelations(query)

	err := query.
		First(&entity, "id = ?", id).
		Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &entity, nil
}

func (r *BaseRepository[T]) FindBy(
	ctx context.Context,
	where map[string]any,
) (*T, error) {

	var entity T

	query := r.db.WithContext(ctx)
	query = r.applyRelations(query)

	query = r.applyWhere(
		query,
		where,
	)

	err := query.First(&entity).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &entity, nil
}

func (r *BaseRepository[T]) FindAllBy(
	ctx context.Context,
	where map[string]any,
) ([]T, error) {

	var entities []T

	query := r.db.WithContext(ctx)
	query = r.applyRelations(query)

	query = r.applyWhere(
		query,
		where,
	)

	err := query.Find(&entities).Error

	return entities, err
}

func (r *BaseRepository[T]) Count(
	ctx context.Context,
	where map[string]any,
) (int64, error) {

	var count int64

	query := r.db.
		WithContext(ctx).
		Model(new(T))

	query = r.applyWhere(
		query,
		where,
	)

	err := query.Count(&count).Error

	return count, err
}

func (r *BaseRepository[T]) Exists(
	ctx context.Context,
	where map[string]any,
) (bool, error) {

	count, err := r.Count(
		ctx,
		where,
	)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *BaseRepository[T]) List(
	ctx context.Context,
	opts common.QueryOptions,
) (*common.PageResult[T], error) {

	query := r.db.
		WithContext(ctx).
		Model(new(T))

	query = r.applyRelations(query)

	query = r.applyFilters(
		query,
		opts.Filters,
	)

	return r.paginate(
		query,
		opts,
	)
}

func (r *BaseRepository[T]) With(relations ...ModelRelation) *BaseRepository[T] {
	c := *r

	c.relations = append(
		append([]ModelRelation{}, r.relations...),
		relations...,
	)

	return &c
}

func CastRepo[T any, R any](base *BaseRepository[T]) R {
	return any(base).(R)
}

func (r *BaseRepository[T]) applyRelations(query *gorm.DB) *gorm.DB {

	for _, rel := range r.relations {
		query = query.Preload(string(rel))
	}

	return query
}

func (r *BaseRepository[T]) paginate(
	query *gorm.DB,
	opts common.QueryOptions,
) (*common.PageResult[T], error) {

	var total int64

	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	order := "created_at DESC"

	if opts.SortBy != "" {

		order = opts.SortBy

		if opts.SortDesc {
			order += " DESC"
		}
	}

	var rows []T

	if err := query.
		Order(order).
		Offset(opts.Offset()).
		Limit(opts.LimitValue()).
		Find(&rows).
		Error; err != nil {

		return nil, err
	}

	limit := opts.LimitValue()

	totalPages :=
		int((total + int64(limit) - 1) /
			int64(limit))

	return &common.PageResult[T]{
		Data:       rows,
		Total:      total,
		Page:       opts.Page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (r *BaseRepository[T]) applyWhere(
	query *gorm.DB,
	where map[string]any,
) *gorm.DB {

	for field, value := range where {
		query = query.Where(
			field+" = ?",
			value,
		)
	}

	return query
}

func (r *BaseRepository[T]) applyFilters(
	query *gorm.DB,
	filters map[string]any,
) *gorm.DB {

	for key, value := range filters {

		def, ok := r.filters[key]
		if !ok {
			continue
		}

		if def.Apply != nil {
			query = def.Apply(query, value)
			continue
		}

		column := def.Column
		if column == "" {
			column = key
		}

		operator := def.Operator
		if operator == "" {
			operator = "="
		}

		if def.Transform != nil {
			value = def.Transform(value)
		}

		query = query.Where(
			column+" "+operator+" ?",
			value,
		)
	}

	return query
}

func (r *BaseRepository[T]) clone() *BaseRepository[T] {
	return &BaseRepository[T]{
		db:        r.db,
		relations: append([]ModelRelation{}, r.relations...),
	}
}
