package fixtures

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	version "github.com/hashicorp/go-version"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MongoDBFlavour string

const (
	MongoDB                 MongoDBFlavour = "mongodb"
	PerconaServerForMongoDB MongoDBFlavour = "psmdb"
)

func (mf MongoDBFlavour) String() string {
	return string(mf)
}

func (mf MongoDBFlavour) Dir() string {
	return filepath.Join(versionsDir(), mf.String())
}

type ServerInfo struct {
	Version string
	Flavour MongoDBFlavour
}

func isServerPSMDB(session *mgo.Session) (bool, error) {
	resp := struct {
		Ok int `bson:ok"`
	}{}
	err := session.Run(bson.M{"getParameter": 1, "profilingRateLimit": true}, &resp)
	if err != nil || resp.Ok != 1 {
		return false, err
	}
	return true, nil
}

func GetServerInfo(session *mgo.Session) (*ServerInfo, error) {
	info := &ServerInfo{
		Flavour: MongoDB,
	}

	buildInfo, err := session.BuildInfo()
	if err != nil {
		return info, err
	}
	info.Version = buildInfo.Version
	if strings.Contains(buildInfo.Version, "-") {
		versionElems := strings.SplitN(buildInfo.Version, "-", 2)
		info.Version = versionElems[0]
		isPSMDB, err := isServerPSMDB(session)
		if err != nil {
			return info, err
		} else if isPSMDB {
			info.Flavour = PerconaServerForMongoDB
		}
	}
	return info, nil
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

func Write(serverInfo *ServerInfo, command string, data []byte) error {
	versionDir := filepath.Join(serverInfo.Flavour.Dir(), serverInfo.Version)
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
