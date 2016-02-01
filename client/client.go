package client

import (
	"database/sql"
	"fmt"
	"github.com/asiainfoLDP/datafactory-servicebroker-mysql/model"
	"github.com/asiainfoLDP/datafactory-servicebroker-mysql/utils"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
)

const DATABASE_NAME = "db_name"

var (
	DB                                                *sql.DB
	DB_ADDR, DB_PORT, DB_DATABASE, DB_USER, DB_PASSWD string
)

type Client interface {
	CreateInstance(parameters interface{}) (string, error)
	BindInstance(instanceId string, parameters interface{}) (string, error)
	GetInstanceState(instanceId string) (string, error)
	DeleteInstance(instance *model.ServiceInstance) error
}

func init() {
	GetEnvs()
	var err error
	URL := fmt.Sprintf(`%s:%s@tcp(%s:%s)/%s`, DB_USER, DB_PASSWD, DB_ADDR, DB_PORT, DB_DATABASE)
	log.Println(URL)
	DB, err = sql.Open("mysql", URL)
	if err != nil {
		log.Fatalln("open failed ", err)
	}
	if err := DB.Ping(); err != nil {
		log.Fatalln("ping db fail", err)
	}
}

func SetCredential(crd model.Credential) error {
	log.Printf(`CREATE USER '%s'@'localhost' IDENTIFIED BY '%s';`, crd.Username, crd.Password)
	if _, err := DB.Exec(fmt.Sprintf(`CREATE USER '%s' IDENTIFIED BY '%s';`, crd.Username, crd.Password)); err != nil {
		log.Println("SetCredential create user err", err)
		return err
	}

	if _, err := DB.Exec(fmt.Sprintf(`GRANT ALL PRIVILEGES ON %s.* TO '%s'@'%%'`, crd.Database, crd.Username)); err != nil {
		log.Println("SetCredential grant err", err)
		return err
	}
	return nil
}

type virtualGuestProps struct {
	hostname                     string
	domain                       string
	startCpus                    int
	maxMemory                    int
	dataCenterName               string
	operatingSystemReferenceCode string
}

type SoftLayerClient struct {
	vgProps virtualGuestProps
}

func (client *SoftLayerClient) CreateInstance(parameters interface{}) (string, error) {

	dataBaseName := fmt.Sprintf("DB_%s", utils.GetUid())
	_, err := DB.Exec(fmt.Sprintf("CREATE DATABASE %s;", dataBaseName))
	if err != nil {
		log.Printf("CREATE DATABASE %s err: %s.", dataBaseName, err)
		return "", err
	}

	return dataBaseName, nil
}

func (client *SoftLayerClient) BindInstance(instanceId string, parameters interface{}) (string, error) {
	return "123", nil
}

func (client *SoftLayerClient) GetInstanceState(instanceId string) (string, error) {
	return "123", nil
}
func (client *SoftLayerClient) DeleteInstance(instance *model.ServiceInstance) error {

	dataBaseName := instance.Id
	log.Printf("%+v", instance)
	if m, ok := instance.Parameters.(map[string]interface{}); ok {
		dataBaseName = m[DATABASE_NAME].(string)
	}
	_, err := DB.Exec(fmt.Sprintf("DROP DATABASE %s;", dataBaseName))
	if err != nil {
		log.Printf("DROP DATABASE %s err: %s.", dataBaseName, err)
	}

	return nil
}
func GetEnvs() {
	DB_ADDR = os.Getenv("MYSQL_ADDR")
	fmt.Printf("ENV[MYSQL_ADDR] is %s", DB_ADDR)
	if DB_ADDR == "" {
		fmt.Println("ENV[MYSQL_ADDR] is null")
		os.Exit(1)
	}
	DB_PORT = os.Getenv("MYSQL_PORT")
	fmt.Printf("ENV[DB_PORT] is %s", DB_PORT)
	if DB_PORT == "" {
		fmt.Println("ENV[MYSQL_PORT] is null")
		os.Exit(1)
	}
	DB_DATABASE = os.Getenv("MYSQL_DATABASE")
	fmt.Printf("ENV[DB_DATABASE] is %s", DB_DATABASE)
	if DB_DATABASE == "" {
		fmt.Println("ENV[MYSQL_DATABASE] is null")
		os.Exit(1)
	}
	DB_USER = os.Getenv("MYSQL_USER")
	fmt.Printf("ENV[DB_USER] is %s", DB_USER)
	if DB_USER == "" {
		fmt.Println("ENV[MYSQL_USER] is null")
		os.Exit(1)
	}
	DB_PASSWD = os.Getenv("MYSQL_ENV_MYSQL_ROOT_PASSWORD")
	fmt.Printf("ENV[DB_PASSWDR] is %s", DB_PASSWD)
	if DB_PASSWD == "" {
		fmt.Println("ENV[MYSQL_PASSWD] is null")
		os.Exit(1)
	}
}
