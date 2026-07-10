package team

import (
	"context"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

// Queries orchestrates team read use cases.
type Queries struct {
	repo *postgres.TeamRepository
}

// NewQueries creates team query use cases.
func NewQueries(repo *postgres.TeamRepository) *Queries {
	return &Queries{repo: repo}
}

// ListTeams returns all teams with member counts.
func (q *Queries) ListTeams(ctx context.Context) ([]dto.TeamResponse, error) {
	teams, err := q.repo.ListTeams(ctx)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	result := make([]dto.TeamResponse, 0, len(teams))
	for i := range teams {
		item, err := q.toTeamSummary(ctx, &teams[i])
		if err != nil {
			return nil, err
		}
		result = append(result, *item)
	}
	return result, nil
}

// GetTeam returns a team with members.
func (q *Queries) GetTeam(ctx context.Context, id uint) (*dto.TeamResponse, error) {
	team, err := q.repo.FindByIDWithMembers(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, utils.ErrNotFound("team not found")
		}
		return nil, utils.ErrInternal(err)
	}
	return q.toTeamDetail(team)
}

func (q *Queries) toTeamSummary(ctx context.Context, team *models.Team) (*dto.TeamResponse, error) {
	memberCount, err := q.repo.CountMembers(ctx, team.ID)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	return &dto.TeamResponse{
		ID:          team.ID,
		Name:        team.Name,
		Slug:        team.Slug,
		Description: team.Description,
		MemberCount: memberCount,
		CreatedAt:   team.CreatedAt,
		UpdatedAt:   team.UpdatedAt,
	}, nil
}

func (q *Queries) toTeamDetail(team *models.Team) (*dto.TeamResponse, error) {
	members := make([]dto.TeamMemberResponse, 0, len(team.Members))
	for _, member := range team.Members {
		members = append(members, dto.TeamMemberResponse{
			UserID:    member.UserID,
			Email:     member.User.Email,
			FirstName: member.User.FirstName,
			LastName:  member.User.LastName,
			Role:      member.Role,
			CreatedAt: member.CreatedAt,
		})
	}

	return &dto.TeamResponse{
		ID:          team.ID,
		Name:        team.Name,
		Slug:        team.Slug,
		Description: team.Description,
		MemberCount: int64(len(members)),
		Members:     members,
		CreatedAt:   team.CreatedAt,
		UpdatedAt:   team.UpdatedAt,
	}, nil
}

// Commands orchestrates team write use cases.
type Commands struct {
	repo    *postgres.TeamRepository
	queries *Queries
}

// NewCommands creates team command use cases.
func NewCommands(repo *postgres.TeamRepository, queries *Queries) *Commands {
	return &Commands{repo: repo, queries: queries}
}

// CreateTeam inserts a new team.
func (c *Commands) CreateTeam(ctx context.Context, req *dto.CreateTeamRequest) (*dto.TeamResponse, error) {
	slug := strings.ToLower(strings.TrimSpace(req.Slug))
	if slug == "" {
		return nil, utils.ErrBadRequest("slug is required")
	}

	existing, err := c.repo.CountBySlug(ctx, slug)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	if existing > 0 {
		return nil, utils.ErrBadRequest("team slug already exists")
	}

	var description *string
	if trimmed := strings.TrimSpace(req.Description); trimmed != "" {
		description = &trimmed
	}

	team := models.Team{
		Name:        strings.TrimSpace(req.Name),
		Slug:        slug,
		Description: description,
	}
	if err := c.repo.Create(ctx, &team); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return c.queries.GetTeam(ctx, team.ID)
}

// UpdateTeam modifies an existing team.
func (c *Commands) UpdateTeam(ctx context.Context, id uint, req *dto.UpdateTeamRequest) (*dto.TeamResponse, error) {
	team, err := c.repo.FindByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return nil, utils.ErrNotFound("team not found")
		}
		return nil, utils.ErrInternal(err)
	}

	var description *string
	if trimmed := strings.TrimSpace(req.Description); trimmed != "" {
		description = &trimmed
	}

	team.Name = strings.TrimSpace(req.Name)
	team.Description = description
	if err := c.repo.Save(ctx, team); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return c.queries.GetTeam(ctx, team.ID)
}

// DeleteTeam removes a team and its memberships.
func (c *Commands) DeleteTeam(ctx context.Context, id uint) error {
	team, err := c.repo.FindByID(ctx, id)
	if err != nil {
		if postgres.IsNotFound(err) {
			return utils.ErrNotFound("team not found")
		}
		return utils.ErrInternal(err)
	}

	if err := c.repo.Delete(ctx, team); err != nil {
		return utils.ErrInternal(err)
	}
	return nil
}

// AddMember assigns a user to a team.
func (c *Commands) AddMember(ctx context.Context, teamID uint, req *dto.AddTeamMemberRequest) (*dto.TeamResponse, error) {
	if _, err := c.repo.FindByID(ctx, teamID); err != nil {
		if postgres.IsNotFound(err) {
			return nil, utils.ErrNotFound("team not found")
		}
		return nil, utils.ErrInternal(err)
	}

	exists, err := c.repo.UserExists(ctx, req.UserID)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}
	if !exists {
		return nil, utils.ErrNotFound("user not found")
	}

	if _, err := c.repo.FindMember(ctx, teamID, req.UserID); err == nil {
		return nil, utils.ErrBadRequest("user is already a team member")
	} else if !postgres.IsNotFound(err) {
		return nil, utils.ErrInternal(err)
	}

	role := strings.TrimSpace(req.Role)
	if role == "" {
		role = "member"
	}
	if role != "lead" && role != "member" {
		return nil, utils.ErrBadRequest("role must be lead or member")
	}

	member := models.TeamMember{
		TeamID: teamID,
		UserID: req.UserID,
		Role:   role,
	}
	if err := c.repo.AddMember(ctx, &member); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return c.queries.GetTeam(ctx, teamID)
}

// RemoveMember unassigns a user from a team.
func (c *Commands) RemoveMember(ctx context.Context, teamID, userID uint) (*dto.TeamResponse, error) {
	if _, err := c.repo.FindByID(ctx, teamID); err != nil {
		if postgres.IsNotFound(err) {
			return nil, utils.ErrNotFound("team not found")
		}
		return nil, utils.ErrInternal(err)
	}

	if _, err := c.repo.FindMember(ctx, teamID, userID); err != nil {
		if postgres.IsNotFound(err) {
			return nil, utils.ErrNotFound("team member not found")
		}
		return nil, utils.ErrInternal(err)
	}

	if err := c.repo.RemoveMember(ctx, teamID, userID); err != nil {
		return nil, utils.ErrInternal(err)
	}
	return c.queries.GetTeam(ctx, teamID)
}
