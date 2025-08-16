package service

import "github.com/fukaraca/runesmith/components/runesmith-backend/service/artifactory"

func (s *Service) GetArtifacts(completed bool) []artifactory.Artifact {
	if completed {
		return s.depot.Done()
	}
	return s.depot.Pending()
}
