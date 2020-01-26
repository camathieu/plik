package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"time"

	"github.com/root-gg/plik/server/metadata"
	"github.com/root-gg/utils"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
 * User input is only safe in document field !!!
 * Keys with ( '.', '$', ... ) may be interpreted
 */

// Ensure Mongo Metadata Backend implements metadata.Backend interface
var _ metadata.Backend = (*Backend)(nil)

// Config object
type Config struct {
	ConnectionString string
	Database         string
	UploadCollection string
	UserCollection   string
}

// NewConfig configures the backend
// from config passed as argument
func NewConfig(params map[string]interface{}) (c *Config) {
	c = new(Config)
	c.ConnectionString = "mongodb://127.0.0.1:27017"
	c.Database = "plik"
	c.UploadCollection = "uploads"
	c.UserCollection = "tokens"
	utils.Assign(c, params)
	return
}

// Backend object
type Backend struct {
	config           *Config
	client           *mongo.Client
	database         *mongo.Database
	uploadCollection *mongo.Collection
	userCollection   *mongo.Collection
}

// NewBackend instantiate a new MongoDB Metadata Backend
// from configuration passed as argument
func NewBackend(config *Config) (b *Backend, err error) {
	b = new(Backend)
	b.config = config

	// TODO use logger or remove
	fmt.Printf("connecting to mongodb %s", b.config.ConnectionString)

	opts := options.Client().ApplyURI(b.config.ConnectionString).SetWriteConcern(writeconcern.New(writeconcern.WMajority()))
	b.client, err = mongo.NewClient(opts)
	if err != nil {
		return nil, err
	}

	ctx, cancel := newContext()
	defer cancel()

	err = b.client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	// TODO use logger or remove
	fmt.Printf("connected to mongodb %s", b.config.ConnectionString)

	b.database = b.client.Database(b.config.Database)
	b.uploadCollection = b.database.Collection(b.config.UploadCollection)
	b.userCollection = b.database.Collection(b.config.UserCollection)

	return b, nil
}

func newContext() (context.Context, func()) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}
