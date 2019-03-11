package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/clbanning/mxj"

	"github.com/gorilla/mux"
)

type handlerFunc = func(w http.ResponseWriter, r *http.Request)

// CreateGetAllHandler creates handler GetAll route for Entity e
func CreateGetAllHandler(e *Entity) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ret := []FieldsMap{}
		for _, entry := range e.Entries {
			ret = append(ret, entry.Fields)
		}
		successResponse(ret, w)
	}
}

// CreateGetOneHandler creates handler GetOne route for Entity e
func CreateGetOneHandler(e *Entity) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		found := e.FindEntryByID(id)
		if found == nil {
			notFoundResponse(w)
			return
		}
		successResponse(found.Fields, w)
	}
}

// CreateCreateHandler creates handler Create route for Entity e
func CreateCreateHandler(e *Entity, onUpdateFn func()) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entryJSON, err := parseBody(r)
		if err != nil {
			badRequestResponse(w)
			fmt.Println(err)
			return
		}
		entry, err := ParseEntry(&entryJSON)
		if err != nil {
			badRequestResponse(w)
			fmt.Println(err)
			return
		}
		e.AppendEntityEntry(entry)
		successResponse(entry.Fields, w)
		defer onUpdateFn()
	}
}

// CreateUpdateHandler creates handler Update route for Entity e
func CreateUpdateHandler(e *Entity, onUpdateFn func()) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entryJSON, err := parseBody(r)
		if err != nil {
			badRequestResponse(w)
			fmt.Println(err)
			return
		}
		entry, err := ParseEntry(&entryJSON)
		entry.ID = mux.Vars(r)["id"]
		fields := entry.Fields
		(*fields)["id"] = mux.Vars(r)["id"]
		if err != nil {
			badRequestResponse(w)
			fmt.Println(err)
			return
		}
		e.UpdateEntityEntry(entry.ID, entry)
		successResponse(entry.Fields, w)
		defer onUpdateFn()
	}
}

func notFoundResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 page not found"))
}

func badRequestResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("400 bad request"))
}

func successResponse(object interface{}, w http.ResponseWriter) {
	backToJSON, _ := json.MarshalIndent(object, "", "  ")
	w.Header().Set("Content-Type", "application/json")
	w.Write(backToJSON)
}

func parseBody(r *http.Request) (interface{}, error) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	mv, err := mxj.NewMapJson(b)

	if err != nil {
		return nil, err
	}
	ret, err := mv.ValueForPath("")
	if err != nil {
		return nil, err
	}

	return ret, nil
}
