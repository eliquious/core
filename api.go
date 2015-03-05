package core

import (
	"log"

	"github.com/gin-gonic/gin"
)

// ResourceHandler allows for CRUD operations on various resources such as Projects, Stories and Tasks.
type ResourceHandler interface {
	List(*gin.Context)
	Get(*gin.Context)
	Put(*gin.Context)
	Post(*gin.Context)
	Delete(*gin.Context)
}

// HTTPResource is represents all the HTTP resources for the API
type HTTPResource interface {
	Bind(KeyValueDatabase) error
}

// API encapsulates all the API functionality for adding all the resources
type API struct {
	router   *gin.Engine
	database KeyValueDatabase
}

// Initialize created the database connection for all the resources to use
func (api *API) Initialize(factory DatabaseConnectionFactory) error {

	// open db connection
	db, err := factory.Connect()
	if err != nil {
		log.Fatalln("Could not open database: ", err)
		return err
	}

	api.database = db
	return nil

}

// AddResources is used to add all the HTTPResource objects to the API
func (api *API) AddResources(resources ...HTTPResource) error {

	// add each resource
	for _, resource := range resources {
		if err := resource.Bind(api.database); err != nil {
			return err
		}
	}
	return nil
}
