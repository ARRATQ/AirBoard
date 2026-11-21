package handlers

import (
	"reflect"
	"testing"

	"airboard/models"
)

func TestPlanPollOptions(t *testing.T) {
	existing := []models.PollOption{
		{ID: 10, PollID: 1, Text: "Oui", Order: 0},
		{ID: 11, PollID: 1, Text: "Non", Order: 1},
		{ID: 12, PollID: 1, Text: "Autre", IsOther: true, Order: 2},
	}

	t.Run("garde les options existantes (et donc leurs votes)", func(t *testing.T) {
		plan := planPollOptions(1, existing, []models.PollOptionRequest{
			{ID: 10, Text: "Oui !"}, {ID: 11, Text: "Non"}, {ID: 12, Text: "Autre", IsOther: true},
		})
		if len(plan.Create) != 0 || len(plan.DeleteIDs) != 0 {
			t.Fatalf("aucune création ni suppression attendue, obtenu %+v", plan)
		}
		if got := plan.Update[0]; got.ID != 10 || got.Text != "Oui !" || got.Order != 0 {
			t.Errorf("mise à jour de l'option 10 = %+v", got)
		}
	})

	t.Run("crée les nouvelles, supprime les retirées, réordonne", func(t *testing.T) {
		plan := planPollOptions(1, existing, []models.PollOptionRequest{
			{ID: 11, Text: "Non"}, {Text: "Peut-être"}, {ID: 10, Text: "Oui"},
		})
		if ids := []uint{plan.Update[0].ID, plan.Update[1].ID}; !reflect.DeepEqual(ids, []uint{11, 10}) {
			t.Errorf("ordre des mises à jour = %v", ids)
		}
		if plan.Update[0].Order != 0 || plan.Update[1].Order != 2 {
			t.Errorf("ordres = %d, %d", plan.Update[0].Order, plan.Update[1].Order)
		}
		if len(plan.Create) != 1 || plan.Create[0].Text != "Peut-être" || plan.Create[0].Order != 1 || plan.Create[0].PollID != 1 {
			t.Errorf("création = %+v", plan.Create)
		}
		if !reflect.DeepEqual(plan.DeleteIDs, []uint{12}) {
			t.Errorf("suppressions = %v", plan.DeleteIDs)
		}
	})

	t.Run("un id inconnu ou d'un autre sondage est traité comme une nouvelle option", func(t *testing.T) {
		plan := planPollOptions(1, existing, []models.PollOptionRequest{
			{ID: 10, Text: "Oui"}, {ID: 999, Text: "Ailleurs"}, {ID: 10, Text: "Doublon"},
		})
		if len(plan.Update) != 1 {
			t.Errorf("une seule mise à jour attendue, obtenu %+v", plan.Update)
		}
		if len(plan.Create) != 2 || plan.Create[0].ID != 0 || plan.Create[1].ID != 0 {
			t.Errorf("créations = %+v", plan.Create)
		}
		if !reflect.DeepEqual(plan.DeleteIDs, []uint{11, 12}) {
			t.Errorf("suppressions = %v", plan.DeleteIDs)
		}
	})
}
