package fixtures

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2/bson"
)

var (
	testVersionsDir  string
	testVersionPSMDB string
)

func TestVersionDir(t *testing.T) {
	testVersionsDir = versionsDir()
	assert.Equal(t, "versions", filepath.Base(testVersionsDir))
}

func TestFlavourString(t *testing.T) {
	assert.Equal(t, "mongodb", MongoDB.String())
	assert.Equal(t, "psmdb", PerconaServerForMongoDB.String())
}

func TestFlavourDir(t *testing.T) {
	assert.Regexp(t, ".+/versions/"+PerconaServerForMongoDB.String()+"$", PerconaServerForMongoDB.Dir())
}

func TestVersions(t *testing.T) {
	assert.NotZero(t, Versions(MongoDB))
	psmdbVersions := Versions(PerconaServerForMongoDB)
	assert.NotZero(t, psmdbVersions)
	testVersionPSMDB = psmdbVersions[0]
}

func TestVersionsFilter(t *testing.T) {
	assert.Len(t, VersionsFilter(PerconaServerForMongoDB, "= "+testVersionPSMDB), 1)
	assert.NotZero(t, VersionsFilter(PerconaServerForMongoDB, "> 1.0"))
	assert.Zero(t, VersionsFilter(PerconaServerForMongoDB, "> 5.0"))
}

type TestDataBSON struct {
	Message string `bson:"msg"`
}

type TestMongoDBFlavour string

var TestMongoDBFlavourMongoDb TestMongoDBFlavour = "mongodb"

func (f TestMongoDBFlavour) String() string {
	return string(f)
}

func (f TestMongoDBFlavour) Dir() string {
	return "/tmp/test-go-mongodb-fixtures"
}

func TestWrite(t *testing.T) {
	testServerInfo := &ServerInfo{
		Version: testVersionPSMDB,
		Flavour: TestMongoDBFlavourMongoDb,
	}
	bytes, err := bson.Marshal(&TestDataBSON{Message: "test123"})
	assert.NoError(t, err)
	assert.NoError(t, Write(testServerInfo, "test", bytes))
}

func TestLoad(t *testing.T) {
	defer os.RemoveAll(TestMongoDBFlavourMongoDb.Dir())
	testData := &TestDataBSON{}
	assert.NoError(t, Load(TestMongoDBFlavourMongoDb, testVersionPSMDB, "test", &testData))
	assert.Equal(t, "test123", testData.Message)
}

func TestIsVersionMatch(t *testing.T) {
	assert.True(t, IsVersionMatch("3.4.13", "> 3"))
	assert.True(t, IsVersionMatch("3.4.13", "> 3.4"))
	assert.True(t, IsVersionMatch("3.4.13", "= 3.4.13"))
	assert.True(t, IsVersionMatch("3.4.13", "!= 2"))
	assert.False(t, IsVersionMatch("3.4.13", "< 3"))
	assert.False(t, IsVersionMatch("3.4.13", "= 2.6.12"))
}
