package handlers

import (
	"reflect"
	"testing"

	"airboard/models"
)

func TestGroupCommentTargets(t *testing.T) {
	comments := []models.Comment{
		{EntityType: "news", EntityID: 8},
		{EntityType: "news", EntityID: 8},
		{EntityType: "news", EntityID: 3},
		{EntityType: "event", EntityID: 5},
		{EntityType: "application", EntityID: 2},
		{EntityType: "inconnu", EntityID: 1},
	}
	got := groupCommentTargets(comments)
	want := map[string][]uint{
		"news":        {8, 3},
		"event":       {5},
		"application": {2},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("groupCommentTargets = %v, attendu %v", got, want)
	}
}

func TestCommentTargetKey(t *testing.T) {
	if got := commentTargetKey("news", 8); got != "news:8" {
		t.Errorf("commentTargetKey = %q", got)
	}
}
