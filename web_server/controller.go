package web_server

import (
	"errors"
	"fmt"
	"net/http"
	client"github.com/asiainfoLDP/broker_mysql/client"
	model "github.com/asiainfoLDP/broker_mysql/model"
	utils "github.com/asiainfoLDP/broker_mysql/utils"
	"log"
)

const (
	DEFAULT_POLLING_INTERVAL_SECONDS = 10
)

type Controller struct {
	cloudName   string
	cloudClient client.Client

	instanceMap   map[string]*model.ServiceInstance
	bindingMap    map[string]*model.ServiceBinding
	credentialMap map[string]*model.Credential
}

func CreateController(cloudName string, instanceMap map[string]*model.ServiceInstance, bindingMap map[string]*model.ServiceBinding, credentialMap map[string]*model.Credential) (*Controller, error) {
	cloudClient, err := createCloudClient(cloudName)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not create cloud: %s client, message: %s", cloudName, err.Error()))
	}

	return &Controller{
		cloudName:     cloudName,
		cloudClient:   cloudClient,
		instanceMap:   instanceMap,
		bindingMap:    bindingMap,
		credentialMap: credentialMap,
	}, nil
}

func (c *Controller) Catalog(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get Service Broker Catalog...")

	var catalog model.Catalog
	catalogFileName := "catalog.json"

	if c.cloudName == utils.AWS {
		catalogFileName = "catalog.AWS.json"
	} else if c.cloudName == utils.SOFTLAYER || c.cloudName == utils.SL {
		catalogFileName = "catalog.SoftLayer.json"
	}

	err := utils.ReadAndUnmarshal(&catalog, conf.CatalogPath, catalogFileName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	utils.WriteResponse(w, http.StatusOK, catalog)
}

func (c *Controller) CreateServiceInstance(w http.ResponseWriter, r *http.Request) {
	user, password, err := utils.ParseBasicAuth(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	if user == "" || password == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	fmt.Println("Create Service Instance...")

	var instance model.ServiceInstance

	err = utils.ProvisionDataFromRequest(r, &instance)
	if err != nil {
		log.Printf(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	instanceId, err := c.cloudClient.CreateInstance(instance.Parameters)
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		return
	}

	instance.InternalId = instanceId
	instance.DashboardUrl = "http://dashbaord_url"
	instance.Id = utils.ExtractVarsFromRequest(r, "service_instance_guid")
	instance.LastOperation = &model.LastOperation{
		State:                    "in progress",
		Description:              "creating service instance...",
		AsyncPollIntervalSeconds: DEFAULT_POLLING_INTERVAL_SECONDS,
	}

	c.instanceMap[instance.Id] = &instance

	crd := model.Credential{
		Uri:      fmt.Sprintf("mysql://%s:%s@%s:3306/%s", user, password, "mysqlhost", instanceId),
		Username: user,
		Password: password,
		Host:     "mysqlhost",
		Port:     3306,
		Database: instanceId,
	}

	if err := client.SetCredential(crd); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.credentialMap[instance.Id] = &crd

	err = utils.MarshalAndRecord(c.instanceMap, conf.DataPath, conf.ServiceInstancesFileName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = utils.MarshalAndRecord(c.credentialMap, conf.DataPath, conf.ServicdCredentialsFileName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := model.CreateServiceInstanceResponse{
		DashboardUrl:  instance.DashboardUrl,
		LastOperation: instance.LastOperation,
	}

	utils.WriteResponse(w, http.StatusAccepted, response)
}

func (c *Controller) GetServiceInstance(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get Service Instance State....")

	instanceId := utils.ExtractVarsFromRequest(r, "service_instance_guid")
	instance := c.instanceMap[instanceId]
	if instance == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	state, err := c.cloudClient.GetInstanceState(instance.InternalId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if state == "pending" {
		instance.LastOperation.State = "in progress"
		instance.LastOperation.Description = "creating service instance..."
	} else if state == "running" {
		instance.LastOperation.State = "succeeded"
		instance.LastOperation.Description = "successfully created service instance"
	} else {
		instance.LastOperation.State = "failed"
		instance.LastOperation.Description = "failed to create service instance"
	}

	response := model.CreateServiceInstanceResponse{
		DashboardUrl:  instance.DashboardUrl,
		LastOperation: instance.LastOperation,
	}
	utils.WriteResponse(w, http.StatusOK, response)
}

func (c *Controller) RemoveServiceInstance(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Remove Service Instance...")

	instanceId := utils.ExtractVarsFromRequest(r, "service_instance_guid")
	instance := c.instanceMap[instanceId]
	if instance == nil {
		w.WriteHeader(http.StatusGone)
		return
	}

	err := c.cloudClient.DeleteInstance(instance)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.deleteCredentials(instanceId)

	delete(c.instanceMap, instanceId)
	utils.MarshalAndRecord(c.instanceMap, conf.DataPath, conf.ServiceInstancesFileName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = c.deleteAssociatedBindings(instanceId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	utils.WriteResponse(w, http.StatusOK, "{}")
}

func (c *Controller) Bind(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Bind Service Instance...")

	bindingId := utils.ExtractVarsFromRequest(r, "service_binding_guid")
	instanceId := utils.ExtractVarsFromRequest(r, "service_instance_guid")

	instance := c.instanceMap[instanceId]
	if instance == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var binding model.ServiceBinding
	err := utils.ProvisionDataFromRequest(r, &binding)

	c.bindingMap[bindingId] = &model.ServiceBinding{
		Id:                bindingId,
		ServiceId:         instance.ServiceId,
		ServicePlanId:     instance.PlanId,
		ServiceInstanceId: instance.Id,
	}

	//
	//	ip, userName, privateKey, err := c.cloudClient.InjectKeyPair(instance.InternalId)
	//	if err != nil {
	//		w.WriteHeader(http.StatusInternalServerError)
	//		return
	//	}
	//
	//	credential := model.Credential{
	//		Uri 	:
	//		Username string `json:username`
	//		Password string `json:password`
	//		Host	 string `json:host`
	//		Port 	 int    `json:port`
	//		Database string `json:database`
	//	}
	//
	//	response := model.CreateServiceBindingResponse{
	//		Credentials: credential,
	//	}
	//
	//
	//
	err = utils.MarshalAndRecord(c.bindingMap, conf.DataPath, conf.ServiceBindingsFileName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if instance, ok := c.credentialMap[instanceId]; ok {
		utils.WriteResponse(w, http.StatusOK, instance)
	} else {
		log.Println("no found instance", http.StatusNotFound)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

//curl -X DELETE  'http://username:password@broker-url/v2/service_instances/:instance_id/
func (c *Controller) UnBind(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Unbind Service Instance...")

	bindingId := utils.ExtractVarsFromRequest(r, "service_binding_guid")
	instanceId := utils.ExtractVarsFromRequest(r, "service_instance_guid")
	instance := c.instanceMap[instanceId]
	if instance == nil {
		w.WriteHeader(http.StatusGone)
		return
	}

	delete(c.bindingMap, bindingId)
	if err := utils.MarshalAndRecord(c.bindingMap, conf.DataPath, conf.ServiceBindingsFileName); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	utils.WriteResponse(w, http.StatusOK, "{}")
}

// Private instance methods

func (c *Controller) deleteAssociatedBindings(instanceId string) error {
	for id, binding := range c.bindingMap {
		if binding.ServiceInstanceId == instanceId {
			delete(c.bindingMap, id)
		}
	}

	return utils.MarshalAndRecord(c.bindingMap, conf.DataPath, conf.ServiceBindingsFileName)
}

func (c *Controller) deleteCredentials(instanceId string) error {
	_, err := client.DB.Exec(fmt.Sprintf("DROP DATABASE %s;", c.credentialMap[instanceId].Database))
	if err != nil {
		log.Printf("DROP DATABASE %s err: %s.", c.credentialMap[instanceId].Database, err)
		return err
	}
	delete(c.credentialMap, instanceId)
	return utils.MarshalAndRecord(c.bindingMap, conf.DataPath, conf.ServicdCredentialsFileName)
}

// Private methods

func createCloudClient(cloudName string) (client.Client, error) {
	switch cloudName {
	case utils.AWS:
		return nil, nil

	case utils.SOFTLAYER, utils.SL, utils.SQL:
		return new(client.SoftLayerClient), nil
	}

	return nil, errors.New(fmt.Sprintf("Invalid cloud name: %s", cloudName))
}
