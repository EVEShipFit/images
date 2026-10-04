// Package icons decides which image every type, market group and meta group
// uses, and converts them to WebP.
//
// Which image a type uses follows EVE-TurtleTools' eveicongenerator
// (https://github.com/SentientTurtle/EVE-TurtleTools). The images are not
// composited, so the marker can be left off.
package icons

import (
	"cmp"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"

	"github.com/EVEShipFit/images/internal/res"
	"github.com/EVEShipFit/images/internal/sde"
)

type Profile string

const (
	// Icon is scaled to 64 pixels, as layers are drawn at their own size.
	// Smaller ones are UI glyphs, which scaling would blur; they are kept as-is.
	Icon Profile = "icon"
	// Lossless is for markers, frames and UI textures, where compression artefacts show.
	Lossless Profile = "lossless"
)

type Source struct {
	Resource string
	Profile  Profile
	MD5      string
}

// Hash only depends on the source's content and how it is converted, so an
// unchanged image keeps its file name between builds.
func (s Source) Hash() uint64 {
	sum := sha256.Sum256(fmt.Appendf(nil, "%s %d %d\x00%s", s.Profile, iconSize, iconQuality, s.MD5))
	return binary.BigEndian.Uint64(sum[:8])
}

func (s Source) fileName() string {
	return fmt.Sprintf("%016x.webp", s.Hash())
}

// Frame is the background and overlay the client draws around blueprints, relics and reactions.
type Frame uint8

const (
	NoFrame Frame = iota
	BlueprintFrame
	RelicFrame
	ReactionFrame
)

// The overlay is drawn additively.
type FrameLayers struct {
	Background, Overlay Source
}

type TypeIcon struct {
	Icon Source
	// Copy is only set for blueprints whose copy has its own icon.
	Copy  *Source
	Frame Frame
	// MetaGroup is zero for types without a marker.
	MetaGroup int32
}

type Plan struct {
	Types        map[int32]TypeIcon
	MarketGroups map[int32]Source
	// A meta group's icon is also the marker of its types.
	MetaGroups map[int32]Source
	Factions   map[int32]Source
	Attributes map[int32]Source
	Effects    map[int32]Source
	UI         map[string]Source

	Blueprint, BlueprintCopy, Relic, Reaction FrameLayers

	// Missing lists the published types without an icon.
	Missing []int32
}

func (p *Plan) frames() []FrameLayers {
	return []FrameLayers{p.Blueprint, p.BlueprintCopy, p.Relic, p.Reaction}
}

// Sources are sorted, so an image keeps its index while the set of images does not change.
func (p *Plan) Sources() []Source {
	seen := map[Source]bool{}
	add := func(sources ...Source) {
		for _, source := range sources {
			seen[source] = true
		}
	}

	for _, frame := range p.frames() {
		add(frame.Background, frame.Overlay)
	}
	for _, icon := range p.Types {
		add(icon.Icon)
		if icon.Copy != nil {
			add(*icon.Copy)
		}
	}
	add(slices.Collect(maps.Values(p.MarketGroups))...)
	add(slices.Collect(maps.Values(p.MetaGroups))...)
	add(slices.Collect(maps.Values(p.Factions))...)
	add(slices.Collect(maps.Values(p.Attributes))...)
	add(slices.Collect(maps.Values(p.Effects))...)
	add(slices.Collect(maps.Values(p.UI))...)

	return slices.SortedFunc(maps.Keys(seen), func(a, b Source) int {
		return cmp.Or(cmp.Compare(a.Resource, b.Resource), cmp.Compare(a.Profile, b.Profile))
	})
}

const (
	categoryBlueprint = 9
	categoryRelic     = 34
	categorySkin      = 91

	renderableTypeList = 140
)

var reactionGroups = []int32{1888, 1889, 1890, 4097}

var uiFolders = []string{
	"res:/ui/texture/classes/carrierbay",
	"res:/ui/texture/classes/fitting",
	"res:/ui/texture/classes/fitting/statsicons",
	"res:/ui/texture/classes/radialmenu/fitting",
	"res:/ui/texture/classes/shipui",
	"res:/ui/texture/classes/shipui/ewarbar",
	"res:/ui/texture/classes/shipui/fighters",
	"res:/ui/texture/eveicon/system_icons",
	"res:/ui/texture/shared",
	"res:/ui/texture/windowicons",
}

// UI textures from folders too large to take whole, like icons/, which holds every type's icon.
var uiFiles = []string{
	"res:/ui/texture/icons/38_16_200.png", // Unfit Module and Remove Charge, on a fitting slot
	"res:/ui/texture/icons/38_16_207.png", // Put Online and Put Offline, on a fitting slot
	"res:/ui/texture/icons/38_16_208.png", // Show Info, on a fitting slot
	"res:/ui/texture/icons/defence.png",   // Defense Mode, of a ship with modes
	"res:/ui/texture/icons/speed.png",     // Propulsion Mode, of a ship with modes
	"res:/ui/texture/icons/target.png",    // Sharpshooter Mode, of a ship with modes
}

type planner struct {
	data       *sde.Data
	index      res.Index
	renderable *sde.TypeList
	metaGroups map[int32]Source
}

func NewPlan(data *sde.Data, index res.Index) (*Plan, error) {
	renderable, ok := data.TypeLists[renderableTypeList]
	if !ok {
		return nil, fmt.Errorf("type list %d (RenderableTypeIDs) is missing from the SDE", renderableTypeList)
	}

	p := &planner{data: data, index: index, renderable: renderable}
	plan := &Plan{
		Types:        map[int32]TypeIcon{},
		MarketGroups: map[int32]Source{},
		MetaGroups:   map[int32]Source{},
		Factions:     map[int32]Source{},
		Attributes:   map[int32]Source{},
		Effects:      map[int32]Source{},
		UI:           map[string]Source{},
	}

	var err error
	if plan.Blueprint, err = p.frame("bpo.png", "bpo_overlay.png"); err != nil {
		return nil, err
	}
	if plan.BlueprintCopy, err = p.frame("bpc.png", "bpc_overlay.png"); err != nil {
		return nil, err
	}
	if plan.Relic, err = p.frame("relic.png", "relic_overlay.png"); err != nil {
		return nil, err
	}
	if plan.Reaction, err = p.frame("reaction.png", "bpo_overlay.png"); err != nil {
		return nil, err
	}

	for id, group := range data.MetaGroups {
		if source, ok := p.iconFile(group.IconID, Lossless); ok {
			plan.MetaGroups[id] = source
		}
	}
	p.metaGroups = plan.MetaGroups

	for id, item := range data.Types {
		if !item.Published {
			continue
		}
		icon, ok := p.typeIcon(item)
		if !ok {
			plan.Missing = append(plan.Missing, id)
			continue
		}
		plan.Types[id] = icon
	}
	slices.Sort(plan.Missing)

	for id, group := range data.MarketGroups {
		if source, ok := p.iconFile(group.IconID, Icon); ok {
			plan.MarketGroups[id] = source
		}
	}

	for id, faction := range data.Factions {
		if source, ok := p.iconFile(faction.IconID, Icon); ok {
			plan.Factions[id] = source
		}
	}

	for id, attribute := range data.DogmaAttributes {
		if !attribute.Published {
			continue
		}
		if source, ok := p.iconFile(attribute.IconID, Lossless); ok {
			plan.Attributes[id] = source
		}
	}

	for id, effect := range data.DogmaEffects {
		if !effect.Published {
			continue
		}
		if source, ok := p.iconFile(effect.IconID, Lossless); ok {
			plan.Effects[id] = source
		}
	}

	for resource := range index {
		if name, ok := uiName(resource); ok {
			plan.UI[name], _ = p.resource(resource, Lossless)
		}
	}

	return plan, nil
}

func (p *planner) frame(background, overlay string) (FrameLayers, error) {
	var layers FrameLayers
	var ok bool
	if layers.Background, ok = p.resource("res:/ui/texture/icons/"+background, Lossless); !ok {
		return layers, fmt.Errorf("frame background %s is not in the resource index", background)
	}
	if layers.Overlay, ok = p.resource("res:/ui/texture/icons/"+overlay, Lossless); !ok {
		return layers, fmt.Errorf("frame overlay %s is not in the resource index", overlay)
	}
	return layers, nil
}

func (p *planner) typeIcon(item *sde.Type) (TypeIcon, bool) {
	var icon TypeIcon

	if item.CategoryID == categoryBlueprint || item.CategoryID == categoryRelic {
		if bp, ok := p.render(item, "_64_bp.png"); ok {
			icon.Icon = bp
			if bpc, ok := p.render(item, "_64_bpc.png"); ok {
				icon.Copy = &bpc
			}
		} else if base, ok := p.iconFile(item.IconID, Icon); ok {
			icon.Icon = base
			switch {
			case item.CategoryID == categoryRelic:
				icon.Frame = RelicFrame
			case slices.Contains(reactionGroups, item.GroupID):
				icon.Frame = ReactionFrame
			default:
				icon.Frame = BlueprintFrame
			}
		} else {
			return icon, false
		}
	} else {
		base, ok := p.render(item, "_64.png")
		if !ok {
			base, ok = p.iconFile(item.IconID, Icon)
		}
		if !ok && item.CategoryID == categorySkin {
			base, ok = p.skin(item.Key)
		}
		if !ok {
			return icon, false
		}
		icon.Icon = base
	}

	if _, ok := p.metaGroups[item.MetaGroupID]; ok {
		icon.MetaGroup = item.MetaGroupID
	}
	return icon, true
}

// render returns the render of a ship, drone or structure.
func (p *planner) render(item *sde.Type, suffix string) (Source, bool) {
	graphic, ok := p.data.Graphics[item.GraphicID]
	if item.GraphicID == 0 || !ok || graphic.IconFolder == "" || !p.renderable.Contains(item) {
		return Source{}, false
	}
	return p.resource(fmt.Sprintf("%s/%d%s", graphic.IconFolder, item.GraphicID, suffix), Icon)
}

// Icon 0 is a question mark; the SDE uses it for types without an icon.
func (p *planner) iconFile(iconID int32, profile Profile) (Source, bool) {
	resource, ok := p.data.Icons[iconID]
	if iconID == 0 || !ok {
		return Source{}, false
	}
	return p.resource(resource, profile)
}

func (p *planner) skin(typeID int32) (Source, bool) {
	material, ok := p.data.SkinMaterials[typeID]
	if !ok {
		return Source{}, false
	}
	return p.resource(fmt.Sprintf("res:/ui/texture/classes/skins/icons/%d.png", material), Icon)
}

// The SDE sometimes refers to files the client no longer has.
func (p *planner) resource(resource string, profile Profile) (Source, bool) {
	entry, ok := p.index[resource]
	if !ok {
		return Source{}, false
	}
	return Source{Resource: resource, Profile: profile, MD5: entry.MD5}, true
}

func uiName(resource string) (string, bool) {
	if path.Ext(resource) != ".png" || !(slices.Contains(uiFolders, path.Dir(resource)) || slices.Contains(uiFiles, resource)) {
		return "", false
	}
	return strings.TrimSuffix(strings.TrimPrefix(resource, "res:/ui/texture/"), ".png"), true
}
