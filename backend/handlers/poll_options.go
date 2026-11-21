package handlers

import "airboard/models"

// pollOptionPlan décrit les changements à appliquer aux options d'un sondage modifié.
type pollOptionPlan struct {
	Update    []models.PollOption // options existantes conservées (texte, « autre » et ordre mis à jour)
	Create    []models.PollOption // nouvelles options
	DeleteIDs []uint              // options retirées
}

// planPollOptions compare les options existantes à celles de la requête.
// Une option envoyée avec l'id d'une option existante du même sondage est mise à jour
// sur place : ses votes restent attachés. Un id absent, inconnu ou déjà utilisé crée
// une nouvelle option. Les options existantes non renvoyées sont supprimées.
// L'ordre suit la position dans la requête.
func planPollOptions(pollID uint, existing []models.PollOption, requested []models.PollOptionRequest) pollOptionPlan {
	byID := make(map[uint]models.PollOption, len(existing))
	for _, opt := range existing {
		byID[opt.ID] = opt
	}

	plan := pollOptionPlan{}
	kept := make(map[uint]bool, len(existing))
	for i, req := range requested {
		if opt, ok := byID[req.ID]; ok && req.ID != 0 && !kept[req.ID] {
			kept[req.ID] = true
			updated := opt
			updated.Text = req.Text
			updated.IsOther = req.IsOther
			updated.Order = i
			plan.Update = append(plan.Update, updated)
			continue
		}
		plan.Create = append(plan.Create, models.PollOption{
			PollID:  pollID,
			Text:    req.Text,
			IsOther: req.IsOther,
			Order:   i,
		})
	}

	for _, opt := range existing {
		if !kept[opt.ID] {
			plan.DeleteIDs = append(plan.DeleteIDs, opt.ID)
		}
	}
	return plan
}
