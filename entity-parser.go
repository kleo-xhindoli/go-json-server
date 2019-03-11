package main

import (
	"encoding/json"
	"errors"
	"math/rand"

	"github.com/clbanning/mxj"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const idLen = 10

// GlobalObject is the root object in the JSON
type GlobalObject struct {
	Entities []*Entity
}

// Entity is one instance of an entity
type Entity struct {
	Name    string
	Entries []*EntityEntry
}

// FieldsMap reference to key-value pairs of entity fields
type FieldsMap *map[string]interface{}

// EntityEntry defines one entry in an Entity
type EntityEntry struct {
	ID     string
	Fields FieldsMap
}

// ParseEntities parses mv Map tree to GlobalObject
func ParseEntities(mv *mxj.Map) (*GlobalObject, error) {
	root, _ := mv.ValueForPath("")
	entities := []*Entity{}
	switch obj := root.(type) {
	case map[string]interface{}:
		for k := range obj {
			parsed, err := ParseEntity(mv, k)
			if err != nil {
				return nil, err
			}
			entities = append(entities, parsed)
		}
	default:
		return nil, errors.New("malformed JSON: Root is not object")
	}

	g := GlobalObject{
		Entities: entities,
	}
	return &g, nil
}

// ParseEntity parses a mv sub tree to an Entity
func ParseEntity(mv *mxj.Map, path string) (*Entity, error) {
	entriesJSON, err := mv.ValuesForPath(path)
	if err != nil {
		return nil, errors.New("could not read Entity " + path)
	}
	entries := []*EntityEntry{}

	for _, entryJSON := range entriesJSON {
		entry, err := ParseEntry(&entryJSON)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	entity := Entity{
		Name:    path,
		Entries: entries,
	}

	return &entity, nil
}

// ParseEntry parses a mv sub tree to an EntityEntry
func ParseEntry(entryJSON *interface{}) (*EntityEntry, error) {
	fields, ok := (*entryJSON).(map[string]interface{})
	if !ok {
		return nil, errors.New("entry in Entity is not object")
	}
	return &EntityEntry{
		ID:     setAndReadID(&fields),
		Fields: &fields,
	}, nil
}

// AppendEntityEntry appends a new EntityEntry to the Entity
func (e *Entity) AppendEntityEntry(entry *EntityEntry) {
	newArr := append(e.Entries, entry)
	e.Entries = newArr
}

// FindEntryByID finds an EntityEntry within an Entity by its ID
func (e *Entity) FindEntryByID(ID string) *EntityEntry {
	var found *EntityEntry
	for _, entry := range e.Entries {
		if entry.ID == ID {
			found = entry
			break
		}
	}
	return found
}

// UpdateEntityEntry updates an existing Entry in an Entity
func (e *Entity) UpdateEntityEntry(ID string, newEntry *EntityEntry) {
	existing := e.FindEntryByID(ID)
	if existing != nil {
		*existing = *newEntry
	}
}

// ToJSON converts the GlobalObject back to JSON to be saved
func (g *GlobalObject) ToJSON() ([]byte, error) {
	root := make(map[string][]FieldsMap)
	for _, entity := range g.Entities {
		root[entity.Name] = []FieldsMap{}
		for _, entry := range entity.Entries {
			root[entity.Name] = append(root[entity.Name], entry.Fields)
		}
	}
	return json.MarshalIndent(root, "", "  ")
}

func setAndReadID(fields FieldsMap) string {
	id, ok := (*fields)["id"].(string)
	if !ok {
		id = generateID()
		(*fields)["id"] = id
	}
	return id
}

func generateID() string {
	b := make([]byte, idLen)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}