package fixtures

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	version "github.com/hashicorp/go-version"
	"gopkg.in/mgo.v2/bson"
)

type MongoDBFlavour string

const (
	MongoDBCommunity        MongoDBFlavour = "mongodb"
	PerconaServerForMongoDB MongoDBFlavour = "psmdb"
)

func (mf MongoDBFlavour) String() string {
	return string(mf)
}

func (mf MongoDBFlavour) Dir() string {
	return filepath.Join(versionsDir(), mf.String())
}

func versionsDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	return filepath.Join(filepath.Dir(filename), "versions")
}

func Load(flavour MongoDBFlavour, versionStr, command string, out interface{}) error {
	filePath := filepath.Join(flavour.Dir(), versionStr, command+".bson")
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	return bson.Unmarshal(bytes, out)
}

func Write(flavour MongoDBFlavour, versionStr, command string, data []byte) error {
	versionDir := filepath.Join(flavour.Dir(), versionStr)
	if _, err := os.Stat(versionDir); os.IsNotExist(err) {
		err = os.MkdirAll(versionDir, 0755)
		if err != nil {
			return err
		}
	}
	filePath := filepath.Join(versionDir, command+".bson")
	return ioutil.WriteFile(filePath, data, 0644)
}

func Versions(flavour MongoDBFlavour) []string {
	var versions []string
	subdirs, err := ioutil.ReadDir(flavour.Dir())
	if err != nil {
		return versions
	}
	for _, subdir := range subdirs {
		if subdir.IsDir() {
			versions = append(versions, subdir.Name())
		}
	}
	return versions
}

func VersionsFilter(flavour MongoDBFlavour, filter string) []string {
	var versions []string
	for _, versionStr := range Versions(flavour) {
		if IsVersionMatch(versionStr, filter) {
			versions = append(versions, versionStr)
		}
	}
	return versions
}

func IsVersionMatch(versionStr, filter string) bool {
	constraints, err := version.NewConstraint(filter)
	if err != nil {
		return false
	}
	v, err := version.NewVersion(versionStr)
	if err != nil {
		return false
	}
	return constraints.Check(v)
}
