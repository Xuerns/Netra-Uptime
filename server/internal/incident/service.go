package incident

import (
	"PulseWatch/internal/model"
	"errors"
	"time"

	"github.com/google/uuid"
)

type IncidentService interface {
	GetIncidentDetail(orgId, incidentId string) (*IncidentResponse, error)
	GetAllIncident(orgId string) ([]IncidentResponse, error)
	AddIncidentUpdate(orgid, incidentId, userId, message string) (*model.IncidentUpdate, error)
	JoinAssignee(orgId, incidentId, userId string) error
	GetIncidentUpdate(orgId, incidentId string) ([]GetIncidentUpdateResponse, error)
}

type IncidentServiceImpl struct {
	repo IncidentRepository
}

func NewIncidentService(repo IncidentRepository) IncidentService {
	return &IncidentServiceImpl{repo: repo}
}

func mapIncidentToResponse(inc model.Incident) IncidentResponse {
	sev := inc.Severity
	if sev == "" {
		sev = "medium"
	}

	resp := IncidentResponse{
		ID:        inc.ID.String(),
		MonitorID: inc.MonitorID.String(),
		Status:    inc.Status,
		Severity:  sev,
		Summary:   inc.Summary,
		StartedAt: inc.StartedAt.Format(time.RFC3339),
	}

	if inc.ResolvedAt != nil {
		resolvedAt := inc.ResolvedAt.Format(time.RFC3339)
		resp.ResolvedAt = &resolvedAt
	}

	if inc.Monitor.ID != uuid.Nil {
		resp.Monitor = &IncidentMonitor{
			Name:   inc.Monitor.Name,
			Target: inc.Monitor.Target,
		}
	}

	for _, assignee := range inc.Assignees {
		resp.Assignees = append(resp.Assignees, IncidentAssignee{
			UserID:    assignee.UserID.String(),
			Role:      assignee.Role,
			CreatedAt: assignee.CreatedAt.Format(time.RFC3339),
			User: IncidentAssigneeUser{
				Name:   assignee.User.Name,
				Avatar: assignee.User.AvatarURL,
			},
		})
	}

	return resp
}

func (s *IncidentServiceImpl) GetIncidentDetail(orgId, incidentId string) (*IncidentResponse, error) {
	incident, err := s.repo.GetById(orgId, incidentId)
	if err != nil {
		return nil, err
	}
	resp := mapIncidentToResponse(*incident)
	return &resp, nil
}

func (s *IncidentServiceImpl) GetAllIncident(orgId string) ([]IncidentResponse, error) {
	incidents, err := s.repo.GetAllByOrg(orgId)
	if err != nil {
		return nil, err
	}

	var response []IncidentResponse
	for _, inc := range incidents {
		response = append(response, mapIncidentToResponse(inc))
	}

	return response, nil
}

func (s *IncidentServiceImpl) AddIncidentUpdate(orgId, incidentId, userId, message string) (*model.IncidentUpdate, error) {
	incident, err := s.repo.GetById(orgId, incidentId)

	if err != nil {
		return nil, errors.New("Incident not found in this organization")
	}

	if incident.Status == "resolved" {
		return nil, errors.New("Cannot add update to resolved incident")
	}

	parsedIncidentId, err := uuid.Parse(incidentId)
	if err != nil {
		return nil, errors.New("Invalid incident ID format")
	}
	parsedUserId, err := uuid.Parse(userId)
	if err != nil {
		return nil, errors.New("Invalid user ID format")
	}

	assignee, err := s.repo.GetAssignee(userId, incidentId)

	if err != nil || assignee == nil {
		return nil, errors.New("You're not involved in this incident / Please participate in this incident first")
	}

	update := &model.IncidentUpdate{
		IncidentID: parsedIncidentId,
		CreatedBy:  parsedUserId,
		Message:    message,
	}

	if err := s.repo.CreateUpdate(update); err != nil {
		return nil, errors.New("Failed to create update Incident")
	}

	return update, nil
}

func (s *IncidentServiceImpl) JoinAssignee(orgId, incidentId, userId string) error {
	_, err := s.repo.GetById(orgId, incidentId)
	if err != nil {
		return errors.New("Incident not found or you don't have access")
	}

	assignee, err := s.repo.GetAssignee(userId, incidentId)

	if assignee != nil {
		return errors.New("You already joined this incident")
	}

	parsedIncidentId, err := uuid.Parse(incidentId)
	if err != nil {
		return errors.New("Invalid incident ID format")
	}
	parsedUserId, err := uuid.Parse(userId)
	if err != nil {
		return errors.New("Invalid user ID format")
	}

	incidentAssignee := &model.IncidentAssignee{
		IncidentID: parsedIncidentId,
		UserID:     parsedUserId,
	}

	if err := s.repo.AddAssignee(incidentAssignee); err != nil {
		return err
	}

	return nil
}

func (s *IncidentServiceImpl) GetIncidentUpdate(orgId, incidentId string) ([]GetIncidentUpdateResponse, error) {
	_, err := s.repo.GetById(orgId, incidentId)
	if err != nil {
		return nil, errors.New("Incident not found or you don't have access")
	}

	incidentUpdates, err := s.repo.GetMessageUpdate(incidentId)

	if err != nil {
		return nil, errors.New("failed to get Incident update")
	}

	var response []GetIncidentUpdateResponse
	for _, update := range incidentUpdates {
		response = append(response, GetIncidentUpdateResponse{
			Message:   update.Message,
			Username:  update.User.Name,
			AvatarUrl: update.User.AvatarURL,
			CreateAt:  update.CreatedAt,
		})
	}

	return response, nil
}
