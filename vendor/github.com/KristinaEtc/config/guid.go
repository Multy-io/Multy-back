package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	uuid "github.com/satori/go.uuid"
)

const uuidFile = "local.conf"

const chmodDir = 0755
const chmodFile = 0644

// getUUIDFromFile read and validate uuid from file
func getUUIDFromFile(filename string) (string, error) {

	defaultLogF("[config] guid: uuidPath=[%s]\n", filename)

	//Reading uuid
	u, err := ioutil.ReadFile(filename)
	if err != nil {
		defaultLogF("[config]: guid: Could not read uuid from [%s]: [%s]\n", filename, err.Error())
		return "", err
	}

	// Validating uuid
	uValidated, err := uuid.FromString(string(u))
	if err != nil {
		defaultLogF("[config]: guid: Could not validate uuid [%s] from file [%s]\n", string(u), filename)
		return "", err
	}
	return uValidated.String(), nil
}

// GetUUID read or create UUID from local.conf file
func GetUUID(DirWithUUID string) string {

	// making full path to local.conf
	//
	fpath, err := GetPathForDir(DirWithUUID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] [config]: guid:  Wrong directory for uuid file: [%s]; return uuid without saving in file\n", err.Error())
	}
	filename := filepath.Join(fpath, uuidFile)
	// trying to take uuid from file
	//
	u, err := getUUIDFromFile(filename)
	if err == nil {
		defaultLogF("[config] guid: uuid=[%s]\n", u)
		return u
	}

	fmt.Fprintf(os.Stderr, "[ERROR] [config]: guid: Could not get uuid from file [%s]; creating new\n", err.Error())

	// trying to create new file with uuid
	//
	u = uuid.NewV4().String()

	exist, err := Exists(fpath)
	if err != nil || !exist {
		defaultLogF("[config]: guid: Directory [%s] doesn't exist; creating...\n", DirWithUUID)
		err = os.MkdirAll(DirWithUUID, chmodDir)
		if err != nil {
			defaultLogF("[config]: guid: Could not create uuid directory [%s]: [%s]; return uuid without saving in file\n", DirWithUUID, err.Error())
		}
	}

	err = ioutil.WriteFile(filename, []byte(u), chmodFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] [config]: guid: Could not create uuid in file [%s]: [%s]; return uuid without saving in file\n", filename, err.Error())
	}

	defaultLogF("[config] guid: uuid=[%s]\n", u)
	return u
}
