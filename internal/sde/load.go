package sde

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Load reads the parts of the SDE zip we need. Files are streamed, as the
// archive does not fit comfortably in memory.
func Load(filename string, build int32) (*Data, error) {
	reader, err := zip.OpenReader(filename)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	data := &Data{
		BuildNumber:     build,
		Types:           map[int32]*Type{},
		Groups:          map[int32]*Group{},
		MarketGroups:    map[int32]*MarketGroup{},
		MetaGroups:      map[int32]*MetaGroup{},
		Factions:        map[int32]*Faction{},
		DogmaAttributes: map[int32]*DogmaAttribute{},
		DogmaEffects:    map[int32]*DogmaEffect{},
		Icons:           map[int32]string{},
		Graphics:        map[int32]*Graphic{},
		TypeLists:       map[int32]*TypeList{},
		SkinMaterials:   map[int32]int32{},
	}

	if err := decode(&reader.Reader, "groups.jsonl", func(entry *Group) {
		data.Groups[entry.Key] = entry
	}); err != nil {
		return nil, err
	}
	if err := decode(&reader.Reader, "marketGroups.jsonl", func(entry *MarketGroup) {
		data.MarketGroups[entry.Key] = entry
	}); err != nil {
		return nil, err
	}
	if err := decode(&reader.Reader, "metaGroups.jsonl", func(entry *MetaGroup) {
		data.MetaGroups[entry.Key] = entry
	}); err != nil {
		return nil, err
	}
	if err := decode(&reader.Reader, "factions.jsonl", func(entry *Faction) {
		data.Factions[entry.Key] = entry
	}); err != nil {
		return nil, err
	}
	if err := decode(&reader.Reader, "dogmaAttributes.jsonl", func(entry *DogmaAttribute) {
		data.DogmaAttributes[entry.Key] = entry
	}); err != nil {
		return nil, err
	}
	if err := decode(&reader.Reader, "dogmaEffects.jsonl", func(entry *DogmaEffect) {
		data.DogmaEffects[entry.Key] = entry
	}); err != nil {
		return nil, err
	}
	if err := decode(&reader.Reader, "types.jsonl", func(entry *Type) {
		data.Types[entry.Key] = entry
	}); err != nil {
		return nil, err
	}
	if err := decode(&reader.Reader, "icons.jsonl", func(entry *icon) {
		data.Icons[entry.Key] = strings.ToLower(entry.IconFile)
	}); err != nil {
		return nil, err
	}
	if err := decode(&reader.Reader, "graphics.jsonl", func(entry *Graphic) {
		data.Graphics[entry.Key] = entry
	}); err != nil {
		return nil, err
	}
	if err := decode(&reader.Reader, "typeLists.jsonl", func(entry *TypeList) {
		data.TypeLists[entry.Key] = entry
	}); err != nil {
		return nil, err
	}

	skins := map[int32]int32{}
	if err := decode(&reader.Reader, "skins.jsonl", func(entry *skin) {
		skins[entry.Key] = entry.SkinMaterialID
	}); err != nil {
		return nil, err
	}
	if err := decode(&reader.Reader, "skinLicenses.jsonl", func(entry *skinLicense) {
		if material, ok := skins[entry.SkinID]; ok {
			data.SkinMaterials[entry.Key] = material
		}
	}); err != nil {
		return nil, err
	}

	for _, item := range data.Types {
		if group, ok := data.Groups[item.GroupID]; ok {
			item.CategoryID = group.CategoryID
		}
	}

	return data, nil
}

func decode[T any](reader *zip.Reader, name string, add func(*T)) error {
	file, err := reader.Open(name)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	for {
		var entry T
		if err := decoder.Decode(&entry); err == io.EOF {
			return nil
		} else if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		add(&entry)
	}
}
