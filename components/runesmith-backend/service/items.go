package service

import (
	"github.com/fukaraca/runesmith/shared"
)

func (s *Service) AllItems() []shared.MagicalItem {
	return s.Items
}

func (s *Service) randomItem() shared.MagicalItem {
	return s.Items[0]
	// return s.Items[rand.Intn(len(s.Items))]
}
