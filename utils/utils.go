package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
)

// EncodeStringHash can calculate the MD5 hash of a string.
func EncodeStringHash(str string) string {
	tmp := md5.New()
	tmp.Write([]byte(str))
	return hex.EncodeToString(tmp.Sum(nil))
}

// EncodeFileHash return file md5 code generation function is used to verify integrity.
func EncodeFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := md5.New()
	io.Copy(hash, file)
	key := hex.EncodeToString(hash.Sum(nil))
	return key, nil
}

// Function checkPathExists determine whether the path exists.
func CheckPathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// LocalFolderList could get all folders name and return.
func LocalFolderList(path string) ([]string, error) {
	var dirList []string
	err := filepath.Walk(path, func(_ string, f fs.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			dirList = append(dirList, f.Name())
			return nil
		}
		return nil
	})
	if err != nil {
		return nil, err
	} else {
		return dirList[1:], nil
	}
}

// LocalDirectoryList could get all files name and return.
func LocalFileList(path string) ([]string, error) {
	var fileList []string
	read, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, f := range read {
		if !f.IsDir() && f.Name()[0] != '.' {
			fileList = append(fileList, f.Name())
		}
	}
	return fileList, nil
}
