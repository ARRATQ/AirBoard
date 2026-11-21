package handlers

import (
	"reflect"
	"testing"
)

func TestMergeGroupAdminTargets(t *testing.T) {
	tests := []struct {
		name      string
		existing  []uint
		requested []uint
		managed   []uint
		want      []uint
		wantOK    bool
	}{
		{"cible ses groupes", nil, []uint{1, 2}, []uint{1, 2}, []uint{1, 2}, true},
		{"garde le groupe qu'il n'administre pas", []uint{1, 9}, []uint{1}, []uint{1}, []uint{1, 9}, true},
		{"retire un de ses groupes, garde l'autre", []uint{1, 2, 9}, []uint{2}, []uint{1, 2}, []uint{2, 9}, true},
		{"renvoie la cible existante d'un autre groupe", []uint{1, 9}, []uint{1, 9}, []uint{1}, []uint{1, 9}, true},
		{"refuse un nouveau groupe qu'il n'administre pas", []uint{1}, []uint{1, 9}, []uint{1}, nil, false},
		{"sans doublon", []uint{9}, []uint{1, 1}, []uint{1}, []uint{1, 9}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := mergeGroupAdminTargets(tt.existing, tt.requested, tt.managed)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, attendu %v", ok, tt.wantOK)
			}
			if ok && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("cibles = %v, attendu %v", got, tt.want)
			}
		})
	}
}
