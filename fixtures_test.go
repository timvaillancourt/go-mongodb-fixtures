package fixtures

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	testVersionsDir        string
	testVersionPSMDB       string
	testVersionPSMDBStatic string = "3.4.13"
	testBSONMessage        string = "test123"
	testEnableDBTests      string = os.Getenv("TEST_ENABLE_DB_TESTS")
	testMongoDBPort        string = os.Getenv("TEST_MONGODB_PORT")
	testPSMDBPort          string = os.Getenv("TEST_PSMDB_PORT")
	testDBVersion          string = os.Getenv("TEST_DB_VERSION")
)

func TestVersionDir(t *testing.T) {
	assert.NotEmpty(t, versionsDir())
	assert.Equal(t, "versions", filepath.Base(versionsDir()))
}

func TestFlavourString(t *testing.T) {
	assert.Equal(t, "mongodb", MongoDB.String())
	assert.Equal(t, "psmdb", PerconaServerForMongoDB.String())
}

func TestFlavourDir(t *testing.T) {
	assert.Regexp(t, ".+/versions/"+PerconaServerForMongoDB.String()+"$", PerconaServerForMongoDB.Dir())
}

func TestVersions(t *testing.T) {
	assert.NotZero(t, Versions(MongoDB), "there must be one or more mongodb versions")
	psmdbVersions := Versions(PerconaServerForMongoDB)
	assert.NotZero(t, psmdbVersions, "there must be one or more psmdb versions")
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
	bytes, err := bson.Marshal(&TestDataBSON{Message: testBSONMessage})
	assert.NoError(t, err)
	assert.NoError(t, Write(testServerInfo, "test", bytes))
}

func TestLoad(t *testing.T) {
	defer os.RemoveAll(TestMongoDBFlavourMongoDb.Dir())
	testData := &TestDataBSON{}
	assert.NoError(t, Load(TestMongoDBFlavourMongoDb, testVersionPSMDB, "test", &testData))
	assert.Equal(t, testBSONMessage, testData.Message)
}

func TestIsVersionMatch(t *testing.T) {
	assert.True(t, IsVersionMatch(testVersionPSMDBStatic, "> 3"))
	assert.True(t, IsVersionMatch(testVersionPSMDBStatic, "> 3.4"))
	assert.True(t, IsVersionMatch(testVersionPSMDBStatic, "= 3.4.13"))
	assert.True(t, IsVersionMatch(testVersionPSMDBStatic, "!= 2"))

	assert.False(t, IsVersionMatch(testVersionPSMDBStatic, "< 3"))
	assert.False(t, IsVersionMatch(testVersionPSMDBStatic, "= 2.6.12"))
	assert.False(t, IsVersionMatch(testVersionPSMDBStatic, ".^^this##.should.break.the.parser.and.return.false.."))
	assert.False(t, IsVersionMatch(".^^this##.should.break.the.parser.and.return.false..", "> 3"))
}

func TestIsServerPSMDB(t *testing.T) {
	if testEnableDBTests != "true" {
		t.Skip("DB tests are disabled, skipping")
	}

	if testPSMDBPort == "" {
		t.Skip("TEST_PSMDB_PORT is not set, skipping")
	}
	psmdb, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:   []string{"localhost:" + testPSMDBPort},
		Direct:  true,
		Timeout: 30 * time.Second,
	})
	defer psmdb.Close()
	assert.NoError(t, err)

	if testMongoDBPort == "" {
		t.Skip("TEST_MONGODB_PORT is not set, skipping")
	}
	mongodb, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:   []string{"localhost:" + testMongoDBPort},
		Direct:  true,
		Timeout: 30 * time.Second,
	})
	defer mongodb.Close()
	assert.NoError(t, err)

	isPSMDB, err := isServerPSMDB(psmdb)
	assert.NoError(t, err, "isServerPSMDB() should return no error")
	assert.True(t, isPSMDB, "isServerPSMDB() should return true")

	isPSMDB, err = isServerPSMDB(mongodb)
	assert.NoError(t, err, "isServerPSMDB() should return no error")
	assert.False(t, isPSMDB, "isServerPSMDB() should return false")
}

func TestGetServerInfo(t *testing.T) {
	if testEnableDBTests != "true" {
		t.Skip("DB tests are disabled, skipping")
	}

	if testPSMDBPort == "" {
		t.Skip("TEST_PSMDB_PORT is not set, skipping")
	}
	psmdb, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:   []string{"localhost:" + testPSMDBPort},
		Direct:  true,
		Timeout: 30 * time.Second,
	})
	defer psmdb.Close()
	assert.NoError(t, err)

	if testMongoDBPort == "" {
		t.Skip("TEST_MONGODB_PORT is not set, skipping")
	}
	mongodb, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:   []string{"localhost:" + testMongoDBPort},
		Direct:  true,
		Timeout: 30 * time.Second,
	})
	defer mongodb.Close()

	serverInfo, err := GetServerInfo(psmdb)
	assert.NoError(t, err, ".GetServerInfo() should not return an error")
	assert.Equal(t, PerconaServerForMongoDB, serverInfo.Flavour, "server flavour is incorrect")
	if testDBVersion != "latest" {
		assert.Truef(t, strings.HasPrefix(serverInfo.Version, testDBVersion), "server version is incorrect. got %s, expected %s*", serverInfo.Version, testDBVersion)
	}

	serverInfo, err = GetServerInfo(mongodb)
	assert.NoError(t, err, ".GetServerInfo() should not return an error")
	assert.Equal(t, MongoDB, serverInfo.Flavour, "server flavour is incorrect")
	if testDBVersion != "latest" {
		assert.Truef(t, strings.HasPrefix(serverInfo.Version, testDBVersion), "server version is incorrect. got %s, expected %s*", serverInfo.Version, testDBVersion)
	}
}
