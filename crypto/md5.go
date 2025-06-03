package crypto

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"

	"github.com/azhai/allgo/fsutil"
)

// Md5Sum 计算md5哈希值
func Md5Sum(data string) string {
	cipher := md5.Sum([]byte(data))
	return hex.EncodeToString(cipher[:])
}

// Md5File 计算文件的md5哈希值
func Md5File(filename string) (string, error) {
	fp, err := fsutil.OpenFile(filename, os.O_RDONLY)
	if err != nil {
		return "", err
	}
	defer fp.Close()
	hash := md5.New()
	if _, err = io.Copy(hash, fp); err != nil {
		return "", err
	}
	sum := hex.EncodeToString(hash.Sum(nil))
	return sum, nil
}
