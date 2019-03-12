package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestParseEntry(t *testing.T) {
	bookWithID := map[string]interface{}{
		"id":    "1234",
		"title": "The history of everything",
		"author": map[string]interface{}{
			"firstName": "John",
			"lastName":  "Smith",
		},
	}
	bookWithoutID := map[string]interface{}{
		"title": "The history of everything",
		"author": map[string]interface{}{
			"firstName": "John",
			"lastName":  "Smith",
		},
	}

	emptyMap := map[string]interface{}{}

	tt := []struct {
		name       string
		data       map[string]interface{}
		shouldFail bool
	}{
		{name: "map with id", data: bookWithID, shouldFail: false},
		{name: "map without id", data: bookWithoutID, shouldFail: false},
		{name: "empty map", data: emptyMap, shouldFail: false},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			entry, err := ParseEntry(tc.data)
			if err != nil {
				if !tc.shouldFail {
					t.Fatalf("expected map to be parsed to EntityEntry\nerr: %s", err)
				} else {
					return
				}
			}

			if tc.shouldFail {
				t.Fatalf("expected failure for %s", tc.name)
			}

			if tc.data["id"] != nil {
				if entry.ID != tc.data["id"] {
					t.Errorf("expected ID %v to match input of %v", entry.ID, tc.data["id"])
				}
			}

			fields := entry.Fields
			if entry.ID != (*fields)["id"] {
				t.Errorf("expected ID \"%s\" in entity to match id \"%s\" in fields", entry.ID, (*fields)["id"])
			}

			eq := reflect.DeepEqual(tc.data, *fields)
			if !eq {
				t.Errorf("expected entry.Fields to be equivalent to input map")
			}
		})
	}
}

func TestParseEntity(t *testing.T) {
	books := []interface{}{
		map[string]interface{}{"id": "book-0", "title": "book 0", "author": "xxx yyy"},
		map[string]interface{}{"id": "book-1", "title": "book 1", "author": "xx yy"},
	}
	empty := []interface{}{}

	tt := []struct {
		name       string
		data       []interface{}
		shouldFail bool
		entityName string
	}{
		{name: "happy path", data: books, entityName: "books", shouldFail: false},
		{name: "empty name", data: books, entityName: "", shouldFail: false}, // Maybe this should fail?
		{name: "empty slice", data: empty, entityName: "empty", shouldFail: false},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			entity, err := ParseEntity(tc.data, tc.entityName)

			if err != nil {
				if !tc.shouldFail {
					t.Fatalf("expected to parse input slice to Entity\nerr: %s", err)
				} else {
					return
				}
			}

			if tc.shouldFail {
				t.Fatalf("expected conversion to fail for %s", tc.name)
			}

			if entity.Name != tc.entityName {
				t.Errorf("expected entity Name \"%s\" to match input of \"%s\"", entity.Name, tc.entityName)
			}

			if len(entity.Entries) != len(tc.data) {
				t.Errorf("expected entity Entries to contain %d entries, instead got %d", len(tc.data), len(entity.Entries))
			}
		})
	}
}

func TestParseEntities(t *testing.T) {
	happyData := []byte(`{
		"books": [{
			"id": "book-0",
			"title": "book 0 title",
			"author": "john smith"
		},{
			"id": "book-1",
			"title": "book 1 title",
			"author": "john smith II"
		}],
		"authors": [{
			"id": "auth-1",
			"firstname": "John",
			"lastname": "Smith"
		}]
	}`)

	tt := []struct {
		name             string
		data             []byte
		entitiesLen      int
		shouldFail       bool
		entityNames      []string
		entityEntriesLen []int // Entries in each entity
	}{
		{
			name:             "happy path",
			data:             happyData,
			entitiesLen:      2,
			shouldFail:       false,
			entityNames:      []string{"books", "authors"},
			entityEntriesLen: []int{2, 1},
		},
		{
			name:             "invalid json",
			data:             []byte(`{fdasf}`),
			entitiesLen:      0,
			shouldFail:       true,
			entityNames:      []string{},
			entityEntriesLen: []int{},
		},
		{
			name:             "empty json",
			data:             []byte(`{}`),
			entitiesLen:      0,
			shouldFail:       false,
			entityNames:      []string{},
			entityEntriesLen: []int{},
		},
		{
			name:             "no entries",
			data:             []byte(`{"books": [], "authors": []}`),
			entitiesLen:      2,
			shouldFail:       false,
			entityNames:      []string{"books", "authors"},
			entityEntriesLen: []int{0, 0},
		},
		{
			name:             "entity object instead of array",
			data:             []byte(`{"books": {}, "authors": []}`),
			entitiesLen:      2,
			shouldFail:       true,
			entityNames:      []string{"books", "authors"},
			entityEntriesLen: []int{0, 0},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			g, err := ParseEntities(tc.data)

			if err != nil {
				if !tc.shouldFail {
					t.Fatalf("expected to parse input JSON to Entities\nerr: %s", err)
				} else {
					return
				}
			}

			if tc.shouldFail {
				t.Fatalf("expected conversion to fail for %s", tc.name)
			}

			if len(g.Entities) != tc.entitiesLen {
				t.Errorf("expected Entities to contain %d entries, instead got %d", tc.entitiesLen, len(g.Entities))
			}

			for idx, e := range g.Entities {
				if e.Name != tc.entityNames[idx] {
					t.Errorf("expected entity name to be %s, instead got %s", tc.entityNames[idx], e.Name)
				}
				if len(e.Entries) != tc.entityEntriesLen[idx] {
					t.Errorf("expected entity %s to contain %d entries, instead got %d", e.Name, tc.entityEntriesLen[idx], len(e.Entries))
				}
			}
		})
	}

}

func TestToJSON(t *testing.T) {
	happyData := []byte(`{
		"books": [{
			"id": "book-0",
			"title": "book 0 title",
			"author": "john smith"
		},{
			"id": "book-1",
			"title": "book 1 title",
			"author": "john smith II"
		}],
		"authors": [{
			"id": "auth-1",
			"firstname": "John",
			"lastname": "Smith"
		}]
	}`)

	tt := []struct {
		name       string
		data       []byte
		shouldFail bool
	}{
		{name: "happy path", data: happyData, shouldFail: false},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			g, err := ParseEntities(tc.data)

			if err != nil {
				t.Fatalf("expected conversion not to fail\nerr: %s", err)
			}

			convertedJSON, err := g.ToJSON()

			if err != nil {
				if !tc.shouldFail {
					t.Fatalf("expected conversion to JSON not to fail\nerr: %s", err)
				} else {
					return
				}
			}

			if tc.shouldFail {
				t.Fatalf("expected conversion to JSON to fail for %s", tc.name)
			}

			var expected map[string]interface{}
			var actual map[string]interface{}
			json.Unmarshal(tc.data, &expected)
			json.Unmarshal(convertedJSON, &actual)

			eq := reflect.DeepEqual(expected, actual)

			if !eq {
				t.Error("expected output JSON to be equivalent to input")
			}

		})
	}
}
