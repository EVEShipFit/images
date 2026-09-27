// Package fbs writes which images every type, market group and meta group
// uses out as a flatbuffer.
//
// The images sub-package is generated; re-generate it after changing the schema.
package fbs

//go:generate flatc --go -o . ../../specs/images.fbs

import (
	"fmt"
	"maps"
	"os"
	"slices"

	flatbuffers "github.com/google/flatbuffers/go"

	"github.com/EVEShipFit/images/internal/fbs/images"
	"github.com/EVEShipFit/images/internal/icons"
)

const NoMarker = 0xFF

func Write(plan *icons.Plan, filename string) error {
	raw, err := Encode(plan)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, raw, 0o644)
}

func Encode(plan *icons.Plan) ([]byte, error) {
	sources := plan.Sources()
	if len(sources) > 1<<16 {
		return nil, fmt.Errorf("%d images do not fit an index of 16 bits", len(sources))
	}
	indices := map[icons.Source]uint16{}
	for i, source := range sources {
		indices[source] = uint16(i)
	}
	// Every source of the plan is in Sources; a miss is a bug, not bad data.
	index := func(source icons.Source) uint16 {
		i, ok := indices[source]
		if !ok {
			panic(fmt.Sprintf("%s is not among the plan's sources", source.Resource))
		}
		return i
	}

	metaGroupIDs := slices.Sorted(maps.Keys(plan.MetaGroups))
	if len(metaGroupIDs) >= NoMarker {
		return nil, fmt.Errorf("%d meta groups do not fit an index of 8 bits", len(metaGroupIDs))
	}
	markers := map[int32]uint8{}
	for i, id := range metaGroupIDs {
		markers[id] = uint8(i)
	}
	marker := func(metaGroup int32) uint8 {
		if metaGroup == 0 {
			return NoMarker
		}
		i, ok := markers[metaGroup]
		if !ok {
			panic(fmt.Sprintf("meta group %d is not among the plan's meta groups", metaGroup))
		}
		return i
	}

	builder := flatbuffers.NewBuilder(1024 * 1024)

	// A flatbuffer needs tables and strings written before the vectors holding
	// them, and builds vectors back to front.
	uiNames := slices.Sorted(maps.Keys(plan.UI))
	uiOffsets := make([]flatbuffers.UOffsetT, len(uiNames))
	for i, name := range uiNames {
		nameOffset := builder.CreateString(name)
		images.UiTextureStart(builder)
		images.UiTextureAddName(builder, nameOffset)
		images.UiTextureAddImage(builder, index(plan.UI[name]))
		uiOffsets[i] = images.UiTextureEnd(builder)
	}
	images.ImagesStartUiTexturesVector(builder, len(uiOffsets))
	for _, offset := range slices.Backward(uiOffsets) {
		builder.PrependUOffsetT(offset)
	}
	uiVector := builder.EndVector(len(uiOffsets))

	images.ImagesStartImagesVector(builder, len(sources))
	for _, source := range slices.Backward(sources) {
		images.CreateImage(builder, source.Hash())
	}
	imagesVector := builder.EndVector(len(sources))

	typeIDs := slices.Sorted(maps.Keys(plan.Types))
	images.ImagesStartTypesVector(builder, len(typeIDs))
	for _, typeID := range slices.Backward(typeIDs) {
		icon := plan.Types[typeID]
		images.CreateTypeIcon(builder, typeID, index(icon.Icon), marker(icon.MetaGroup), frame(icon.Frame))
	}
	typesVector := builder.EndVector(len(typeIDs))

	copies := map[int32]icons.Source{}
	for typeID, icon := range plan.Types {
		if icon.Copy != nil {
			copies[typeID] = *icon.Copy
		}
	}
	copiesVector := keyedVector(builder, images.ImagesStartBlueprintCopiesVector, copies, index)
	marketGroupsVector := keyedVector(builder, images.ImagesStartMarketGroupsVector, plan.MarketGroups, index)
	metaGroupsVector := keyedVector(builder, images.ImagesStartMetaGroupsVector, plan.MetaGroups, index)
	factionsVector := keyedVector(builder, images.ImagesStartFactionsVector, plan.Factions, index)

	frameImages := func(layers icons.FrameLayers) flatbuffers.UOffsetT {
		return images.CreateFrameImages(builder, index(layers.Background), index(layers.Overlay))
	}

	images.ImagesStart(builder)
	images.ImagesAddImages(builder, imagesVector)
	images.ImagesAddBlueprint(builder, frameImages(plan.Blueprint))
	images.ImagesAddBlueprintCopy(builder, frameImages(plan.BlueprintCopy))
	images.ImagesAddRelic(builder, frameImages(plan.Relic))
	images.ImagesAddReaction(builder, frameImages(plan.Reaction))
	images.ImagesAddTypes(builder, typesVector)
	images.ImagesAddBlueprintCopies(builder, copiesVector)
	images.ImagesAddMarketGroups(builder, marketGroupsVector)
	images.ImagesAddMetaGroups(builder, metaGroupsVector)
	images.ImagesAddUiTextures(builder, uiVector)
	images.ImagesAddFactions(builder, factionsVector)
	builder.FinishWithFileIdentifier(images.ImagesEnd(builder), []byte("ESFI"))

	return builder.FinishedBytes(), nil
}

func frame(frame icons.Frame) images.Frame {
	switch frame {
	case icons.NoFrame:
		return images.FrameNone
	case icons.BlueprintFrame:
		return images.FrameBlueprint
	case icons.RelicFrame:
		return images.FrameRelic
	case icons.ReactionFrame:
		return images.FrameReaction
	default:
		panic(fmt.Sprintf("frame %d has no counterpart in the schema", frame))
	}
}

type startVector func(*flatbuffers.Builder, int) flatbuffers.UOffsetT

func keyedVector(builder *flatbuffers.Builder, start startVector, entries map[int32]icons.Source, index func(icons.Source) uint16) flatbuffers.UOffsetT {
	ids := slices.Sorted(maps.Keys(entries))
	start(builder, len(ids))
	for _, id := range slices.Backward(ids) {
		images.CreateKeyedImage(builder, id, index(entries[id]))
	}
	return builder.EndVector(len(ids))
}
