package service

import (
	"github.com/felix-001/qnHackathon/internal/model"
	"time"
)

type ProjectService struct {
	projects []model.Project
}

func NewProjectService() *ProjectService {
	return &ProjectService{
		projects: make([]model.Project, 0),
	}
}

func (s *ProjectService) List() []model.Project {
	return s.projects
}

func (s *ProjectService) Create(project *model.Project) error {
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()
	project.Status = "active"
	s.projects = append(s.projects, *project)
	return nil
}

func (s *ProjectService) Update(id string, project *model.Project) error {
	for i, p := range s.projects {
		if p.ID == id {
			project.UpdatedAt = time.Now()
			s.projects[i] = *project
			return nil
		}
	}
	return nil
}

func (s *ProjectService) Delete(id string) error {
	for i, p := range s.projects {
		if p.ID == id {
			s.projects = append(s.projects[:i], s.projects[i+1:]...)
			return nil
		}
	}
	return nil
}
