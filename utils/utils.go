package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

// StringToMD5 can calculate the MD5 hash of a string.
func StringToMD5(str string) string {
	tmp := md5.New()
	tmp.Write([]byte(str))
	return hex.EncodeToString(tmp.Sum(nil))
}

// FileToMD5 file md5 code generation function is used to verify integrity.
func FileToMD5(path string) (string, error) {
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
