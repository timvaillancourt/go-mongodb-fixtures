package main

import (
	"flag"
	"log"

	fixtures "github.com/timvaillancourt/go-mongodb-fixtures"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type serverCommand struct {
	Name  string
	Value string
	Db    string
	Coll  string
	Query *mgo.Query
}

var (
	mongodbUri     = flag.String("uri", "mongodb://localhost:27017", "mongodb server uri")
	serverCommands = []serverCommand{
		{Name: "currentOp", Value: "1", Db: "admin"},
		{Name: "getCmdLineOpts", Value: "1", Db: "admin"},
		{Name: "hostInfo", Value: "1", Db: "admin"},
		{Name: "isMaster", Value: "1", Db: "admin"},
		{Name: "listCollections", Value: "1", Db: "admin"},
		{Name: "listDatabases", Value: "1", Db: "admin"},
		{Name: "replSetGetConfig", Value: "1", Db: "admin"},
		{Name: "replSetGetStatus", Value: "1", Db: "admin"},
		{Name: "serverStatus", Value: "1", Db: "admin"},
		{Name: "top", Value: "1", Db: "admin"},
	}
)

func main() {
	flag.Parse()

	session, err := mgo.Dial(*mongodbUri)
	if err != nil {
		log.Fatalf("cannot get db connection: %s", err.Error())
	}
	defer session.Close()

	info, err := fixtures.GetServerInfo(session)
	if err != nil {
		log.Fatalf("cannot get db version: %s", err.Error())
	}

	log.Printf("Connected to a %s instance with version %s", info.Flavour, info.Version)

	var data bson.Raw
	for _, cmd := range serverCommands {
		log.Printf("Running command on db %s: '{%s: \"%s\"}'\n", cmd.Db, cmd.Name, cmd.Value)

		err = session.DB(cmd.Db).Run(bson.D{{cmd.Name, cmd.Value}}, &data)
		if err != nil {
			log.Fatalf("Database command %s failed: %s", cmd.Name, err)
		}

		err = fixtures.Write(info, cmd.Name, data.Data)
		if err != nil {
			log.Fatalf("Failed to write %s bson fixture: %s", cmd.Name, err)
		}
	}
}
