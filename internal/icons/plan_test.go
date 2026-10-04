package icons

import (
	"reflect"
	"testing"

	"github.com/EVEShipFit/images/internal/res"
	"github.com/EVEShipFit/images/internal/sde"
)

const (
	shipGroup   = 25
	moduleGroup = 60
	bpGroup     = 105
	relicGroup  = 971
	skinGroup   = 1950
)

func testData() (*sde.Data, res.Index) {
	data := &sde.Data{
		Types: map[int32]*sde.Type{},
		Groups: map[int32]*sde.Group{
			shipGroup:   {Key: shipGroup, CategoryID: 6},
			moduleGroup: {Key: moduleGroup, CategoryID: 7},
			bpGroup:     {Key: bpGroup, CategoryID: categoryBlueprint},
			relicGroup:  {Key: relicGroup, CategoryID: categoryRelic},
			1888:        {Key: 1888, CategoryID: categoryBlueprint},
			skinGroup:   {Key: skinGroup, CategoryID: categorySkin},
		},
		MarketGroups: map[int32]*sde.MarketGroup{
			1: {Key: 1, IconID: 10},
			2: {Key: 2},
		},
		MetaGroups: map[int32]*sde.MetaGroup{
			1: {Key: 1},
			2: {Key: 2, IconID: 20},
		},
		Factions: map[int32]*sde.Faction{
			500001: {Key: 500001, IconID: 30},
			500020: {Key: 500020},
		},
		DogmaAttributes: map[int32]*sde.DogmaAttribute{
			9:  {Key: 9, IconID: 40, Published: true},
			37: {Key: 37, Published: true},
			48: {Key: 48, IconID: 12, Published: true},
			49: {Key: 49, IconID: 40},
		},
		DogmaEffects: map[int32]*sde.DogmaEffect{
			11: {Key: 11, IconID: 40, Published: true},
			12: {Key: 12, IconID: 40},
		},
		Icons: map[int32]string{
			0:  "res:/ui/texture/icons/question.png",
			10: "res:/ui/texture/icons/module.png",
			11: "res:/ui/texture/icons/ship.png",
			12: "res:/ui/texture/icons/gone.png",
			20: "res:/ui/texture/icons/t2.png",
			30: "res:/ui/texture/icons/caldari.png",
			40: "res:/ui/texture/icons/hp.png",
		},
		Graphics: map[int32]*sde.Graphic{
			100: {Key: 100, IconFolder: "res:/dx9/model/ship/rifter/icons"},
		},
		TypeLists: map[int32]*sde.TypeList{
			renderableTypeList: {Key: renderableTypeList, IncludedCategoryIDs: []int32{6}, IncludedGroupIDs: []int32{bpGroup}},
		},
		SkinMaterials: map[int32]int32{600: 7},
	}

	add := func(item *sde.Type) {
		item.Published = true
		item.CategoryID = data.Groups[item.GroupID].CategoryID
		data.Types[item.Key] = item
	}
	add(&sde.Type{Key: 1, GroupID: moduleGroup, IconID: 10})
	add(&sde.Type{Key: 2, GroupID: moduleGroup, IconID: 10, MetaGroupID: 2})
	add(&sde.Type{Key: 3, GroupID: shipGroup, IconID: 11, GraphicID: 100})
	add(&sde.Type{Key: 4, GroupID: moduleGroup, IconID: 11, GraphicID: 100})
	add(&sde.Type{Key: 5, GroupID: bpGroup, IconID: 11, GraphicID: 100})
	add(&sde.Type{Key: 6, GroupID: bpGroup, IconID: 10, MetaGroupID: 2})
	add(&sde.Type{Key: 7, GroupID: 1888, IconID: 10})
	add(&sde.Type{Key: 8, GroupID: relicGroup, IconID: 10})
	add(&sde.Type{Key: 9, GroupID: moduleGroup, IconID: 12})
	add(&sde.Type{Key: 10, GroupID: moduleGroup})
	add(&sde.Type{Key: 600, GroupID: skinGroup})
	data.Types[11] = &sde.Type{Key: 11, GroupID: moduleGroup, IconID: 10}

	index := res.Index{}
	for _, resource := range []string{
		"res:/ui/texture/icons/bpo.png", "res:/ui/texture/icons/bpo_overlay.png",
		"res:/ui/texture/icons/bpc.png", "res:/ui/texture/icons/bpc_overlay.png",
		"res:/ui/texture/icons/relic.png", "res:/ui/texture/icons/relic_overlay.png",
		"res:/ui/texture/icons/reaction.png",
		"res:/ui/texture/icons/question.png",
		"res:/ui/texture/icons/module.png",
		"res:/ui/texture/icons/ship.png",
		"res:/ui/texture/icons/t2.png",
		"res:/ui/texture/icons/caldari.png",
		"res:/ui/texture/icons/hp.png",
		"res:/dx9/model/ship/rifter/icons/100_64.png",
		"res:/dx9/model/ship/rifter/icons/100_64_bp.png",
		"res:/dx9/model/ship/rifter/icons/100_64_bpc.png",
		"res:/ui/texture/classes/skins/icons/7.png",
		"res:/ui/texture/classes/fitting/statsicons/armorhp.png",
		"res:/ui/texture/classes/fitting/statsicons/armorhp.dds",
		"res:/ui/texture/classes/other/thing.png",
		"res:/ui/texture/icons/38_16_200.png",
		"res:/ui/texture/icons/38_16_199.png",
	} {
		index[resource] = res.Entry{MD5: resource}
	}
	return data, index
}

func TestNewPlan(t *testing.T) {
	data, index := testData()
	plan, err := NewPlan(data, index)
	if err != nil {
		t.Fatal(err)
	}

	icon := func(resource string) Source { return Source{Resource: resource, Profile: Icon, MD5: resource} }
	lossless := func(resource string) Source { return Source{Resource: resource, Profile: Lossless, MD5: resource} }
	module := icon("res:/ui/texture/icons/module.png")
	bpc := icon("res:/dx9/model/ship/rifter/icons/100_64_bpc.png")
	t2 := lossless("res:/ui/texture/icons/t2.png")

	tests := []struct {
		name   string
		typeID int32
		want   TypeIcon
	}{
		{"module", 1, TypeIcon{Icon: module}},
		{"tech II module gets a marker", 2, TypeIcon{Icon: module, MetaGroup: 2}},
		{"ship uses its render", 3, TypeIcon{Icon: icon("res:/dx9/model/ship/rifter/icons/100_64.png")}},
		{"unrenderable type falls back to its icon", 4, TypeIcon{Icon: icon("res:/ui/texture/icons/ship.png")}},
		{"ship blueprint uses its renders", 5, TypeIcon{Icon: icon("res:/dx9/model/ship/rifter/icons/100_64_bp.png"), Copy: &bpc}},
		{"module blueprint is framed", 6, TypeIcon{Icon: module, Frame: BlueprintFrame, MetaGroup: 2}},
		{"reaction", 7, TypeIcon{Icon: module, Frame: ReactionFrame}},
		{"relic", 8, TypeIcon{Icon: module, Frame: RelicFrame}},
		{"skin", 600, TypeIcon{Icon: icon("res:/ui/texture/classes/skins/icons/7.png")}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := plan.Types[test.typeID]
			if !ok {
				t.Fatalf("type %d has no icon", test.typeID)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("got %+v, want %+v", got, test.want)
			}
		})
	}

	for _, typeID := range []int32{9, 10, 11} {
		if _, ok := plan.Types[typeID]; ok {
			t.Errorf("type %d has an icon, want none", typeID)
		}
	}
	if !reflect.DeepEqual(plan.Missing, []int32{9, 10}) {
		t.Errorf("missing = %v, want [9 10]", plan.Missing)
	}

	if want := map[int32]Source{1: module}; !reflect.DeepEqual(plan.MarketGroups, want) {
		t.Errorf("market groups = %+v, want %+v", plan.MarketGroups, want)
	}
	if want := map[int32]Source{2: t2}; !reflect.DeepEqual(plan.MetaGroups, want) {
		t.Errorf("meta groups = %+v, want %+v", plan.MetaGroups, want)
	}
	if want := map[int32]Source{500001: icon("res:/ui/texture/icons/caldari.png")}; !reflect.DeepEqual(plan.Factions, want) {
		t.Errorf("factions = %+v, want %+v", plan.Factions, want)
	}
	hp := lossless("res:/ui/texture/icons/hp.png")
	if want := map[int32]Source{9: hp}; !reflect.DeepEqual(plan.Attributes, want) {
		t.Errorf("attributes = %+v, want %+v", plan.Attributes, want)
	}
	if want := map[int32]Source{11: hp}; !reflect.DeepEqual(plan.Effects, want) {
		t.Errorf("effects = %+v, want %+v", plan.Effects, want)
	}
	wantUI := map[string]Source{
		"classes/fitting/statsicons/armorhp": lossless("res:/ui/texture/classes/fitting/statsicons/armorhp.png"),
		"icons/38_16_200":                    lossless("res:/ui/texture/icons/38_16_200.png"),
	}
	if !reflect.DeepEqual(plan.UI, wantUI) {
		t.Errorf("ui = %+v, want %+v", plan.UI, wantUI)
	}

	wantFrame := FrameLayers{lossless("res:/ui/texture/icons/reaction.png"), lossless("res:/ui/texture/icons/bpo_overlay.png")}
	if plan.Reaction != wantFrame {
		t.Errorf("reaction frame = %+v, want %+v", plan.Reaction, wantFrame)
	}
}

func TestNewPlanNeedsBlueprintLayers(t *testing.T) {
	data, index := testData()
	delete(index, "res:/ui/texture/icons/bpo_overlay.png")
	if _, err := NewPlan(data, index); err == nil {
		t.Error("expected an error for a missing blueprint layer")
	}
}
