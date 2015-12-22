package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"encoding/base64"
	"github.com/gorilla/mux"
	"log"
	"regexp"
	"strings"

	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"io"
)
const encodeStd = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789ac"
var (
	REG_BASIC_AUTH = regexp.MustCompile(`^Basic (.+)$`)
	URLEncoding = base64.NewEncoding(encodeStd)
)

func ReadAndUnmarshal(object interface{}, dir string, fileName string) error {
	path := dir + string(os.PathSeparator) + fileName

	bytes, err := ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, object)
	if err != nil {
		return err
	}

	return nil
}

func MarshalAndRecord(object interface{}, dir, fileName string) error {
	log.Println(object, dir, fileName)
	MkDir(dir)
	path := dir + string(os.PathSeparator) + fileName

	bytes, err := json.MarshalIndent(object, "", " ")
	if err != nil {
		return err
	}

	return WriteFile(path, bytes)
}

func WriteResponse(w http.ResponseWriter, code int, object interface{}) {
	data, err := json.Marshal(object)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	fmt.Fprintf(w, string(data))
}

func ProvisionDataFromRequest(r *http.Request, object interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, object)
	if err != nil {
		return err
	}
	return nil
}

func ExtractVarsFromRequest(r *http.Request, varName string) string {
	return mux.Vars(r)[varName]
}

func ReadFile(path string) (content []byte, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}
	content = bytes

	return
}

func WriteFile(path string, content []byte) error {
	err := ioutil.WriteFile(path, content, 0700)
	if err != nil {
		return err
	}

	return nil
}

func GetPath(paths []string) string {
	workDirectory, _ := os.Getwd()

	if len(paths) == 0 {
		return workDirectory
	}

	resultPath := workDirectory

	for _, path := range paths {
		resultPath += string(os.PathSeparator)
		resultPath += path
	}

	return resultPath
}

func Exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func MkDir(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0700)
		if err != nil {
			return false
		}
	}

	return true
}

func ParseBasicAuth(r *http.Request) (string, string, error) {
	s := r.Header.Get("Authorization")

	match := REG_BASIC_AUTH.FindAllStringSubmatch(s, -1)
	if match == nil {
		return "", "", fmt.Errorf("bad auth header %s\n", s)
	}

	s = match[0][1]

	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", "", err
	}

	s = string(b)

	ary := strings.Split(s, ":")
	if len(ary) != 2 {
		return "", "", fmt.Errorf("bad auth string %s\n", s)
	}

	return ary[0], ary[1], nil
}

func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func GetGuid() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return GetMd5String(URLEncoding.EncodeToString(b))
}

func base64Encode(src []byte) string {
	return URLEncoding.EncodeToString(src)
}

func GetUid() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64Encode(b)[:12]
}
