package service

import (
	"github.com/felix-001/qnHackathon/internal/model"
	"time"
)

type ReleaseService struct {
	releases []model.Release
}

func NewReleaseService() *ReleaseService {
	return &ReleaseService{
		releases: make([]model.Release, 0),
	}
}

func (s *ReleaseService) List() []model.Release {
	return s.releases
}

func (s *ReleaseService) Create(release *model.Release) error {
	release.CreatedAt = time.Now()
	release.Status = "pending_approval"
	s.releases = append(s.releases, *release)
	return nil
}

func (s *ReleaseService) Get(id string) (*model.Release, error) {
	for _, r := range s.releases {
		if r.ID == id {
			return &r, nil
		}
	}
	return nil, nil
}

func (s *ReleaseService) Rollback(id string, targetVersion string, reason string) error {
	for i, r := range s.releases {
		if r.ID == id {
			s.releases[i].Status = "rolled_back"
			return nil
		}
	}
	return nil
}

func (s *ReleaseService) UpdateGitlabPR(id string, prURL string) error {
	for i, r := range s.releases {
		if r.ID == id {
			s.releases[i].GitlabPRURL = prURL
			return nil
		}
	}
	return nil
}

func (s *ReleaseService) ApproveReview(id string) error {
	for i, r := range s.releases {
		if r.ID == id {
			s.releases[i].Status = "approved"
			return nil
		}
	}
	return nil
}

func (s *ReleaseService) Deploy(id string) error {
	for i, r := range s.releases {
		if r.ID == id {
			now := time.Now()
			s.releases[i].Status = "deploying"
			s.releases[i].StartedAt = &now
			return nil
		}
	}
	return nil
}

func (s *ReleaseService) CompleteDeployment(id string) error {
	for i, r := range s.releases {
		if r.ID == id {
			now := time.Now()
			s.releases[i].Status = "completed"
			s.releases[i].CompletedAt = &now
			return nil
		}
	}
	return nil
}
