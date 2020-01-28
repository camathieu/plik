package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/root-gg/plik/server/metadata"
	"github.com/root-gg/utils"

	"go.mongodb.org/mongo-driver/mongo/readpref"
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
	TimeoutInSeconds int
}

// NewConfig configures the backend
// from config passed as argument
func NewConfig(params map[string]interface{}) (c *Config) {
	c = new(Config)
	c.ConnectionString = "mongodb://127.0.0.1:27017"
	c.Database = "plik"
	c.UploadCollection = "uploads"
	c.UserCollection = "tokens"
	c.TimeoutInSeconds = 5
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
	fmt.Printf("connecting to %s\n", b.config.ConnectionString)

	// Create client
	opts := options.Client().ApplyURI(b.config.ConnectionString)
	b.client, err = mongo.NewClient(opts)
	if err != nil {
		return nil, err
	}

	// Connect to mongodb replica set cluster

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = b.client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = b.client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		return nil, err
	}

	// TODO use logger or remove
	fmt.Printf("connected to %s\n", b.config.ConnectionString)

	b.database = b.client.Database(b.config.Database)
	b.uploadCollection = b.database.Collection(b.config.UploadCollection)
	b.userCollection = b.database.Collection(b.config.UserCollection)

	return b, nil
}

func (b *Backend) newContext() (context.Context, func()) {
	return context.WithTimeout(context.Background(), time.Duration(b.config.TimeoutInSeconds) * time.Second)
}

func runTransactionWithRetry(sctx mongo.SessionContext, txnFn func(mongo.SessionContext) error) error {
	for {
		err := txnFn(sctx) // Performs transaction.
		if err == nil {
			return nil
		}

		// If transient error, retry the whole transaction
		if cmdErr, ok := err.(mongo.CommandError); ok && cmdErr.HasErrorLabel("TransientTransactionError") {
			continue
		}
		return err
	}
}

func commitWithRetry (sctx mongo.SessionContext) error {
	for {
		err := sctx.CommitTransaction(sctx)
		switch e := err.(type) {
		case nil:
			return nil
		case mongo.CommandError:
			// Can retry commit
			if e.HasErrorLabel("UnknownTransactionCommitResult") {
				continue
			}
			return e
		default:
			return e
		}
	}
}
