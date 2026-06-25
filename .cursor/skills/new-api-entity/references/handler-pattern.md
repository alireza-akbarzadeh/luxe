# Handler & Swagger pattern

Load when writing handler methods for a new entity. Business logic lives in `application/<ctx>/`; GORM in `infrastructure/postgres/`.

## Service method

```go
func (s *entityService) GetByID(ctx context.Context, id uint) (*dto.EntityResponse, error) {
    var m models.Entity
    if err := s.db.WithContext(ctx).First(&m, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, utils.ErrNotFound("entity not found")
        }
        return nil, utils.ErrInternal(err)
    }
    return mapEntityToDTO(&m), nil
}
```

## Controller handler

```go
func (c *EntityController) GetByID(ctx *gin.Context) {
    id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
    if err != nil {
        utils.BadRequest(ctx, "invalid id")
        return
    }
    out, err := c.service.GetByID(ctx.Request.Context(), uint(id))
    if err != nil {
        utils.HandleAppError(ctx, err)
        return
    }
    utils.OK(ctx, out)
}
```

## Swagger

```go
// @Summary Get entity by ID
// @Tags entities
// @Produce json
// @Param id path int true "Entity ID"
// @Success 200 {object} utils.Response{data=dto.EntityResponse}
// @Failure 404 {object} utils.Response
// @Security BearerAuth
// @Router /api/v1/entities/{id} [get]
```

Match sibling routes for auth: `AuthMiddleware` vs admin role groups in `*_routes.go`.
