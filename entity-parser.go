package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

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

// ParseEntities parses a JSON string tree to GlobalObject
func ParseEntities(dbJSON []byte) (*GlobalObject, error) {
	var root map[string]interface{}
	err := json.Unmarshal(dbJSON, &root)
	if err != nil {
		return nil, err
	}
	entities := []*Entity{}
	for k, v := range root {
		parsed, err := ParseEntity(v, k)
		if err != nil {
			return nil, err
		}
		entities = append(entities, parsed)
	}

	g := GlobalObject{
		Entities: entities,
	}
	return &g, nil
}

// ParseEntity parses a JSON sub tree to an Entity
func ParseEntity(entriesJSON interface{}, path string) (*Entity, error) {
	entriesArray, ok := entriesJSON.([]interface{})
	if !ok {
		return nil, errors.New("could not read Entity " + path)

	}
	entries := []*EntityEntry{}

	for _, entryJSON := range entriesArray {
		entry, err := ParseEntry(entryJSON)
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

// ParseEntry parses a JSON sub tree to an EntityEntry
func ParseEntry(entryJSON interface{}) (*EntityEntry, error) {
	fields, ok := entryJSON.(map[string]interface{})
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
func (e *Entity) FindEntryByID(ID string) (*EntityEntry, error) {
	for _, entry := range e.Entries {
		if entry.ID == ID {
			return entry, nil
		}
	}
	return nil, errors.New("Entry not found")
}

// UpdateEntityEntry updates an existing Entry in an Entity
func (e *Entity) UpdateEntityEntry(ID string, newEntry *EntityEntry) error {
	existing, err := e.FindEntryByID(ID)
	if err != nil {
		return err
	}
	*existing = *newEntry
	return nil
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
	rand.Seed(time.Now().UTC().UnixNano())
	b := make([]byte, 16)
	rand.Read(b)
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid
}
