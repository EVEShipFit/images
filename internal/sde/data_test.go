package sde

import (
	"encoding/json"
	"testing"
)

func TestTypeListContains(t *testing.T) {
	list := &TypeList{
		IncludedCategoryIDs: []int32{6},
		ExcludedGroupIDs:    []int32{30},
		IncludedGroupIDs:    []int32{105},
		IncludedTypeIDs:     []int32{3},
		ExcludedTypeIDs:     []int32{4},
	}

	tests := []struct {
		name string
		item Type
		want bool
	}{
		{"included category", Type{Key: 1, GroupID: 25, CategoryID: 6}, true},
		{"other category", Type{Key: 2, GroupID: 60, CategoryID: 7}, false},
		{"excluded group in included category", Type{Key: 5, GroupID: 30, CategoryID: 6}, false},
		{"included group in other category", Type{Key: 6, GroupID: 105, CategoryID: 9}, true},
		{"included type in excluded group", Type{Key: 3, GroupID: 30, CategoryID: 6}, true},
		{"excluded type in included category", Type{Key: 4, GroupID: 25, CategoryID: 6}, false},
	}
	for _, test := range tests {
		if got := list.Contains(&test.item); got != test.want {
			t.Errorf("%s: got %v, want %v", test.name, got, test.want)
		}
	}
}

func TestGraphicUnmarshal(t *testing.T) {
	var graphic Graphic
	if err := json.Unmarshal([]byte(`{"_key": 38, "iconFolder": "res:/dx9/Model/Ship/Caldari/Frigate/CF1/Icons/"}`), &graphic); err != nil {
		t.Fatal(err)
	}
	if graphic.IconFolder != "res:/dx9/model/ship/caldari/frigate/cf1/icons" {
		t.Errorf("icon folder = %q", graphic.IconFolder)
	}
}
