package sde

import (
	"encoding/json"
	"slices"
	"strings"
)

// Data holds the parts of the SDE that decide which image a type gets.
type Data struct {
	BuildNumber   int32
	Types         map[int32]*Type
	Groups        map[int32]*Group
	MarketGroups  map[int32]*MarketGroup
	MetaGroups    map[int32]*MetaGroup
	Factions      map[int32]*Faction
	Icons         map[int32]string
	Graphics      map[int32]*Graphic
	TypeLists     map[int32]*TypeList
	SkinMaterials map[int32]int32
}

type Type struct {
	Key         int32 `json:"_key"`
	GroupID     int32 `json:"groupID"`
	CategoryID  int32 `json:"-"`
	IconID      int32 `json:"iconID"`
	GraphicID   int32 `json:"graphicID"`
	MetaGroupID int32 `json:"metaGroupID"`
	Published   bool  `json:"published"`
}

type Group struct {
	Key        int32 `json:"_key"`
	CategoryID int32 `json:"categoryID"`
}

type MarketGroup struct {
	Key    int32 `json:"_key"`
	IconID int32 `json:"iconID"`
}

type MetaGroup struct {
	Key    int32 `json:"_key"`
	IconID int32 `json:"iconID"`
}

type Faction struct {
	Key      int32  `json:"_key"`
	FlatLogo string `json:"flatLogo"`
}

type Graphic struct {
	Key        int32  `json:"_key"`
	IconFolder string `json:"iconFolder"`
}

// The resource index uses lowercased paths.
func (g *Graphic) UnmarshalJSON(raw []byte) error {
	type plain Graphic
	if err := json.Unmarshal(raw, (*plain)(g)); err != nil {
		return err
	}
	g.IconFolder = strings.TrimSuffix(strings.ToLower(g.IconFolder), "/")
	return nil
}

type TypeList struct {
	Key                 int32   `json:"_key"`
	Name                string  `json:"name"`
	IncludedTypeIDs     []int32 `json:"includedTypeIDs"`
	ExcludedTypeIDs     []int32 `json:"excludedTypeIDs"`
	IncludedGroupIDs    []int32 `json:"includedGroupIDs"`
	ExcludedGroupIDs    []int32 `json:"excludedGroupIDs"`
	IncludedCategoryIDs []int32 `json:"includedCategoryIDs"`
	ExcludedCategoryIDs []int32 `json:"excludedCategoryIDs"`
}

func (l *TypeList) Contains(item *Type) bool {
	if slices.Contains(l.ExcludedTypeIDs, item.Key) {
		return false
	}
	if slices.Contains(l.IncludedTypeIDs, item.Key) {
		return true
	}
	if slices.Contains(l.ExcludedGroupIDs, item.GroupID) {
		return false
	}
	if slices.Contains(l.IncludedGroupIDs, item.GroupID) {
		return true
	}
	return slices.Contains(l.IncludedCategoryIDs, item.CategoryID) && !slices.Contains(l.ExcludedCategoryIDs, item.CategoryID)
}

type icon struct {
	Key      int32  `json:"_key"`
	IconFile string `json:"iconFile"`
}

type skin struct {
	Key            int32 `json:"_key"`
	SkinMaterialID int32 `json:"skinMaterialID"`
}

type skinLicense struct {
	Key    int32 `json:"_key"`
	SkinID int32 `json:"skinID"`
}
