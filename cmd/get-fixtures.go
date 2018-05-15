package main

import (
	"flag"
	"log"
	"strings"

	fixtures "github.com/timvaillancourt/go-mongodb-fixtures"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type serverCommand struct {
	Name  string
	Value string
	Db    string
	Coll  string
	Query string
}

var (
	mongodbUri     = flag.String("uri", "mongodb://localhost:27017", "mongodb server uri")
	serverCommands = []serverCommand{
		serverCommand{Name: "currentOp", Value: "1", Db: "admin"},
		serverCommand{Name: "getCmdLineOpts", Value: "1", Db: "admin"},
		serverCommand{Name: "hostInfo", Value: "1", Db: "admin"},
		serverCommand{Name: "isMaster", Value: "1", Db: "admin"},
		serverCommand{Name: "listCollections", Value: "1", Db: "admin"},
		serverCommand{Name: "listDatabases", Value: "1", Db: "admin"},
		serverCommand{Name: "replSetGetConfig", Value: "1", Db: "admin"},
		serverCommand{Name: "replSetGetStatus", Value: "1", Db: "admin"},
		serverCommand{Name: "serverStatus", Value: "1", Db: "admin"},
		serverCommand{Name: "top", Value: "1", Db: "admin"},
	}
)

func serverVersion(session *mgo.Session) (string, error) {
	buildInfo, err := session.BuildInfo()
	if err != nil {
		return "", err
	}
	if strings.Contains(buildInfo.Version, "-") {
		version := strings.SplitN(buildInfo.Version, "-", 2)
		return version[0], nil
	}
	return buildInfo.Version, nil
}

func serverFlavour(session *mgo.Session) (fixtures.MongoDBFlavour, error) {
	return fixtures.PerconaServerForMongoDB, nil
}

func main() {
	flag.Parse()

	session, err := mgo.Dial(*mongodbUri)
	if err != nil {
		log.Fatalf("cannot get db connection: %s", err.Error())
	}

	version, err := serverVersion(session)
	if err != nil {
		log.Fatalf("cannot get db version: %s", err.Error())
	}

	flavour, err := serverFlavour(session)
	if err != nil {
		log.Fatalf("cannot get db flavour: %s", err.Error())
	}

	var data bson.Raw
	for _, cmd := range serverCommands {
		log.Printf("Running command on db %s: '{%s: \"%s\"}'\n", cmd.Db, cmd.Name, cmd.Value)

		err = session.DB(cmd.Db).Run(bson.D{{cmd.Name, cmd.Value}}, &data)
		if err != nil {
			panic(err)
		}

		err = fixtures.Write(flavour, version, cmd.Name, data.Data)
		if err != nil {
			panic(err)
		}
	}
}
