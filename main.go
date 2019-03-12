package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var globalObject GlobalObject

func main() {
	// Define flags
	store := flag.String("store", "./db.json", "JSON file to store/retrieve data from")
	port := flag.Int("port", 9001, "the port to run the server on")
	flag.Parse()

	dbData, err := loadDataFromFile(*store)
	if err != nil {
		log.Fatalf("Could not read from file %s\n", *store)
	}

	r := mux.NewRouter()
	buildRoutes(*store, dbData, r)

	portStr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Server running on port %d\n", *port)
	log.Fatal(http.ListenAndServe(portStr, r))
}

func buildRoutes(storeFilePath string, data []byte, r *mux.Router) *mux.Router {

	globalObject, err := ParseEntities(data)

	if err != nil {
		log.Fatal(err)
	}

	onUpdateFn := saveDataToFile(storeFilePath, globalObject)

	for _, entity := range globalObject.Entities {
		multipleRoute := fmt.Sprintf("/%s", entity.Name)
		singleRoute := fmt.Sprintf("/%s/{id}", entity.Name)
		r.HandleFunc(multipleRoute, CreateGetAllHandler(entity)).Methods("GET")
		r.HandleFunc(singleRoute, CreateGetOneHandler(entity)).Methods("GET")
		r.HandleFunc(multipleRoute, CreateCreateHandler(entity, onUpdateFn)).Methods("POST")
		r.HandleFunc(singleRoute, CreateUpdateHandler(entity, onUpdateFn)).Methods("PUT")
	}

	return r
}

func saveDataToFile(filePath string, g *GlobalObject) func() {
	return func() {
		data, err := g.ToJSON()
		if err != nil {
			fmt.Printf("Could not write to file\n%s\n", err)
		}
		ioutil.WriteFile(filePath, data, 0644)
	}
}

func loadDataFromFile(filePath string) ([]byte, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return data, nil
}
