package fbs

import (
	"testing"

	"github.com/EVEShipFit/images/internal/fbs/images"
	"github.com/EVEShipFit/images/internal/icons"
)

func source(resource string, profile icons.Profile) icons.Source {
	return icons.Source{Resource: resource, Profile: profile, MD5: resource}
}

func testPlan() *icons.Plan {
	module := source("res:/module.png", icons.Icon)
	bpc := source("res:/bpc.png", icons.Icon)
	frame := func(name string) icons.FrameLayers {
		return icons.FrameLayers{Background: source("res:/"+name+".png", icons.Lossless), Overlay: source("res:/"+name+"_overlay.png", icons.Lossless)}
	}

	return &icons.Plan{
		Types: map[int32]icons.TypeIcon{
			20: {Icon: module, Frame: icons.BlueprintFrame, MetaGroup: 2},
			10: {Icon: module, MetaGroup: 2},
			30: {Icon: source("res:/bp.png", icons.Icon), Copy: &bpc},
		},
		MarketGroups:  map[int32]icons.Source{5: module},
		MetaGroups:    map[int32]icons.Source{1: source("res:/t1.png", icons.Lossless), 2: source("res:/t2.png", icons.Lossless)},
		UI:            map[string]icons.Source{"statsicons/armorhp": source("res:/armorhp.png", icons.Lossless), "shared/b": module},
		Blueprint:     frame("bpo"),
		BlueprintCopy: frame("bpc"),
		Relic:         frame("relic"),
		Reaction:      frame("reaction"),
	}
}

func TestEncode(t *testing.T) {
	plan := testPlan()

	raw, err := Encode(plan)
	if err != nil {
		t.Fatal(err)
	}
	if !images.ImagesBufferHasIdentifier(raw) {
		t.Fatal("identifier is missing")
	}
	root := images.GetRootAsImages(raw, 0)

	hashOf := func(index uint16) uint64 {
		var image images.Image
		root.Images(&image, int(index))
		return image.Hash()
	}
	markerOf := func(entry *images.TypeIcon) uint64 {
		var group images.KeyedImage
		root.MetaGroups(&group, int(entry.Marker()))
		return hashOf(group.Image())
	}

	var entry images.TypeIcon
	var ids []int32
	for i := 0; i < root.TypesLength(); i++ {
		root.Types(&entry, i)
		ids = append(ids, entry.TypeId())
	}
	if len(ids) != 3 || ids[0] != 10 || ids[1] != 20 || ids[2] != 30 {
		t.Fatalf("type IDs = %v, want sorted [10 20 30]", ids)
	}

	root.Types(&entry, 1)
	if hashOf(entry.Icon()) != plan.Types[20].Icon.Hash() || markerOf(&entry) != plan.MetaGroups[2].Hash() || entry.Frame() != images.FrameBlueprint {
		t.Errorf("blueprint = icon %d, marker %d, frame %v", entry.Icon(), entry.Marker(), entry.Frame())
	}
	root.Types(&entry, 2)
	if entry.Marker() != NoMarker || entry.Frame() != images.FrameNone {
		t.Errorf("ship blueprint = marker %d, frame %v", entry.Marker(), entry.Frame())
	}

	var copy images.KeyedImage
	if root.BlueprintCopiesLength() != 1 {
		t.Fatalf("blueprint copies = %d, want 1", root.BlueprintCopiesLength())
	}
	root.BlueprintCopies(&copy, 0)
	if copy.Id() != 30 || hashOf(copy.Image()) != plan.Types[30].Copy.Hash() {
		t.Errorf("blueprint copy = type %d, image %d", copy.Id(), copy.Image())
	}

	frame := root.BlueprintCopy(nil)
	if hashOf(frame.Background()) != plan.BlueprintCopy.Background.Hash() || hashOf(frame.Overlay()) != plan.BlueprintCopy.Overlay.Hash() {
		t.Error("blueprint copy frame points at the wrong images")
	}

	var texture images.UiTexture
	root.UiTextures(&texture, 0)
	if string(texture.Name()) != "shared/b" {
		t.Errorf("first UI texture = %q, want the textures sorted on name", texture.Name())
	}
}

func TestEncodeUnknownMetaGroup(t *testing.T) {
	plan := testPlan()
	plan.Types[10] = icons.TypeIcon{Icon: plan.Types[10].Icon, MetaGroup: 99}

	defer func() {
		if recover() == nil {
			t.Error("expected a panic for a meta group the plan does not have")
		}
	}()
	Encode(plan)
}
