package model

type ServiceBinding struct {
	Id                string `json:"id"`
	ServiceId         string `json:"service_id"`
	AppId             string `json:"app_id"`
	ServicePlanId     string `json:"service_plan_id"`
	PrivateKey        string `json:"private_key"`
	ServiceInstanceId string `json:"service_instance_id"`
}

type CreateServiceBindingResponse struct {
	// SyslogDrainUrl string      `json:"syslog_drain_url, omitempty"`
	Credentials interface{} `json:"credentials"`
}

type Credential struct {
	Uri      string `json:uri`
	Username string `json:username`
	Password string `json:password`
	Host     string `json:host`
	Port     int    `json:port`
	Database string `json:database`
}
