package postgres

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// TeamRepository persists teams and membership with GORM.
type TeamRepository struct {
	db *gorm.DB
}

// NewTeamRepository creates a GORM-backed team repository.
func NewTeamRepository(db *gorm.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

// ListTeams returns all teams ordered by name.
func (r *TeamRepository) ListTeams(ctx context.Context) ([]models.Team, error) {
	var teams []models.Team
	err := r.db.WithContext(ctx).Order("name ASC").Find(&teams).Error
	return teams, err
}

// FindByID loads a team by primary key.
func (r *TeamRepository) FindByID(ctx context.Context, id uint) (*models.Team, error) {
	var team models.Team
	err := r.db.WithContext(ctx).First(&team, id).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// FindByIDWithMembers loads a team with members and user profiles preloaded.
func (r *TeamRepository) FindByIDWithMembers(ctx context.Context, id uint) (*models.Team, error) {
	var team models.Team
	err := r.db.WithContext(ctx).
		Preload("Members", func(db *gorm.DB) *gorm.DB {
			return db.Order("team_members.created_at ASC")
		}).
		Preload("Members.User").
		First(&team, id).Error
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// CountBySlug counts teams matching a slug.
func (r *TeamRepository) CountBySlug(ctx context.Context, slug string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Team{}).Where("slug = ?", slug).Count(&count).Error
	return count, err
}

// CountMembers counts members assigned to a team.
func (r *TeamRepository) CountMembers(ctx context.Context, teamID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.TeamMember{}).Where("team_id = ?", teamID).Count(&count).Error
	return count, err
}

// Create inserts a new team.
func (r *TeamRepository) Create(ctx context.Context, team *models.Team) error {
	return r.db.WithContext(ctx).Create(team).Error
}

// Save persists team changes.
func (r *TeamRepository) Save(ctx context.Context, team *models.Team) error {
	return r.db.WithContext(ctx).Save(team).Error
}

// Delete removes a team row.
func (r *TeamRepository) Delete(ctx context.Context, team *models.Team) error {
	return r.db.WithContext(ctx).Delete(team).Error
}

// FindMember loads a team membership row.
func (r *TeamRepository) FindMember(ctx context.Context, teamID, userID uint) (*models.TeamMember, error) {
	var member models.TeamMember
	err := r.db.WithContext(ctx).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// AddMember inserts a team membership row.
func (r *TeamRepository) AddMember(ctx context.Context, member *models.TeamMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// RemoveMember deletes a team membership row.
func (r *TeamRepository) RemoveMember(ctx context.Context, teamID, userID uint) error {
	return r.db.WithContext(ctx).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Delete(&models.TeamMember{}).Error
}

// UserExists reports whether a user id exists.
func (r *TeamRepository) UserExists(ctx context.Context, userID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Count(&count).Error
	return count > 0, err
}
