package pnap

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"

	"github.com/pkg/errors"

	"github.com/PNAP/go-sdk-helper-bmc/command/bmcapi/server"
	"github.com/PNAP/go-sdk-helper-bmc/receiver"

	"github.com/PNAP/go-sdk-helper-bmc/command/billingapi/product"
	helperdto "github.com/PNAP/go-sdk-helper-bmc/dto"
	jwt "github.com/golang-jwt/jwt/v4"
	bmcapiclient "github.com/phoenixnap/go-sdk-bmc/bmcapi/v3"
)

// Driver is the implementation of BaseDriver interface
type Driver struct {
	*drivers.BaseDriver
	client *receiver.BMCSDK

	//APIToken           string
	//UserAgentPrefix    string
	ID                   string
	Status               string
	Name                 string
	ServerDescription    string
	PrivateIPAddresses   []string
	PublicIPAddresses    []string
	ServerOs             string
	ServerType           string
	ServerLocation       string
	CPU                  string
	RAM                  string
	Storage              string
	ClientIdentifier     string
	ClientSecret         string
	BearerToken          string
	ProvisionedOn        *time.Time
	ServerPrivateNetwork string
	PrivateNetworking    bool
	ServerGateway        string
	UserDataFile         string
}

const (
	defaultOS                = "ubuntu/bionic"
	defaultType              = "s1.c1.medium"
	defaultLocation          = "PHX"
	defaultprivateNetworking = false
)

// NewDriver creates and returns a new instance of the PNAP driver
func NewDriver() *Driver {
	return &Driver{
		ServerOs:          defaultOS,
		ServerType:        defaultType,
		ServerLocation:    defaultLocation,
		PrivateNetworking: defaultprivateNetworking,

		BaseDriver: &drivers.BaseDriver{},
	}
}

// getClient creates the pnap API Client
func (d *Driver) getClient() (*receiver.BMCSDK, error) {
	if d.client == nil {
		var pnapClient (receiver.BMCSDK)

		var confErr error

		configuration := helperdto.Configuration{}
		configuration.TokenURL = "https://auth.phoenixnap.com/auth/realms/BMC/protocol/openid-connect/token"
		configuration.ApiHostName = "https://api.phoenixnap.com/"
		configuration.UserAgent = "PNAP-Rancher-Node-Driver/0.5.0"
		configuration.PoweredBy = "PNAP-Rancher-Node-Driver"

		log.Infof("Client id *** %s ", d.ClientIdentifier)
		//log.Infof("Client id *** %s ", d.ClientSecret)
		if d.BearerToken != "" && d.isTokenValid() {
			log.Info("Token auth with BMC API will be performed..")
			configuration.BearerToken = d.BearerToken
			pnapClient = receiver.NewBMCSDKWithTokenAuthentication(configuration)
			return &pnapClient, nil
		} else if (d.ClientIdentifier != "") && (d.ClientSecret != "") {
			//pnapClient = client.NewPNAPClient(d.ClientIdentifier, d.ClientSecret)
			log.Info("Cloud credentials will be used for authentication..")
			//log.Infof("Client id %s ", d.ClientIdentifier)
			configuration.ClientID = d.ClientIdentifier
			configuration.ClientSecret = d.ClientSecret
			pnapClient = receiver.NewBMCSDK(configuration)
		} else {
			//log.Info("Default config auth will be performed..")
			pnapClient, confErr = receiver.NewBMCSDKWithDefaultConfig(configuration)
			if confErr != nil {
				return nil, errors.Wrap(confErr, "PNAP API client can not be created")

			}
		}

		d.client = &pnapClient
	}
	return d.client, nil
}

// GetCreateFlags returns the mcnflag.Flag slice representing the flags
// that can be set, their descriptions and defaults.
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "PNAP_SERVER_OS",
			Name:   "pnap-server-os",
			Usage:  "The server’s OS ID used when the server was created (e.g., ubuntu/bionic, centos/centos7).",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "PNAP_SERVER_LOCATION",
			Name:   "pnap-server-location",
			Usage:  "Server Location ID. Cannot be changed once a server is created",
		},
		mcnflag.StringFlag{
			EnvVar: "PNAP_SERVER_TYPE",
			Name:   "pnap-server-type",
			Usage:  "Server type ID. Cannot be changed once a server is created",
		},
		mcnflag.StringFlag{
			EnvVar: "PNAP_SERVER_DESCRIPTION",
			Name:   "pnap-server-description",
			Usage:  "Server description",
		},
		mcnflag.StringFlag{
			EnvVar: "PNAP_SERVER_HOSTNAME",
			Name:   "pnap-server-hostname",
			Usage:  "Server hostname",
		},
		mcnflag.StringFlag{
			EnvVar: "PNAP_CLIENT_ID",
			Name:   "pnap-client-identifier",
			Usage:  "Client ID from Application Credentials",
		},
		mcnflag.StringFlag{
			EnvVar: "PNAP_CLIENT_SECRET",
			Name:   "pnap-client-secret",
			Usage:  "Client Secret from Application Credentials",
		},
		mcnflag.StringFlag{
			EnvVar: "PNAP_CLIENT_TOKEN",
			Name:   "pnap-client-token",
			Usage:  "Client Token generated by Authentication Service",
		},
		mcnflag.StringFlag{
			EnvVar: "PNAP_SERVER_PRIVATE_NETWORK",
			Name:   "pnap-server-private-network",
			Usage:  "Private Network ID",
		},
		mcnflag.StringFlag{
			EnvVar: "PNAP_SERVER_GATEWAY",
			Name:   "pnap-server-gateway",
			Usage:  "Server Gateway",
		},
		mcnflag.BoolFlag{
			EnvVar: "PNAP_PRIVATE_NETWORKING",
			Name:   "pnap-private-networking",
			Usage:  "Defines whether to use private network for communication.",
		},
		mcnflag.StringFlag{
			EnvVar: "PNAP_USERDATA",
			Name:   "pnap-userdata",
			Usage:  "path to file with cloud-init user data",
		},
	}
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "pnap"
}

// SetConfigFromFlags configures the driver with the object that was returned
// by RegisterCreateFlags
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Name = flags.String("pnap-server-hostname")
	d.ServerLocation = flags.String("pnap-server-location")
	d.ServerOs = flags.String("pnap-server-os")
	d.ServerType = flags.String("pnap-server-type")
	d.ServerDescription = flags.String("pnap-server-description")
	d.ClientIdentifier = flags.String("pnap-client-identifier")
	d.ClientSecret = flags.String("pnap-client-secret")
	d.BearerToken = flags.String("pnap-client-token")
	d.PrivateNetworking = flags.Bool("pnap-private-networking")
	d.ServerPrivateNetwork = flags.String("pnap-server-private-network")
	d.ServerGateway = flags.String("pnap-server-gateway")
	d.UserDataFile = flags.String("pnap-userdata")

	return nil
}
func (d *Driver) createSSHKey() (string, error) {
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return "", err
	}

	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return "", err
	}

	return string(publicKey), nil
}

// publicSSHKeyPath is always SSH Key Path appended with ".pub"
func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

// Create a host using the driver's config
func (d *Driver) Create() error {
	log.Info("Creating pnap machine instance...")
	//log.Infof("Driver params host:%s clientID:%s type:%s os:%s", d.Name, d.ClientIdentifier, d.ServerType, d.ServerOs)
	publicKey, err := d.createSSHKey()
	if err != nil {
		return err
	}

	client, err := d.getClient()
	if err != nil {
		return err
	}

	request := &bmcapiclient.ServerCreate{}
	request.Hostname = d.MachineName
	var desc = d.ServerDescription
	if len(desc) > 0 {
		request.Description = &desc
	}
	request.Os = d.ServerOs
	request.Type = d.ServerType
	request.Location = d.ServerLocation

	request.SshKeys = []string{strings.TrimSpace(publicKey)}

	query := &helperdto.Query{}

	networkConfigurationObject := bmcapiclient.NetworkConfiguration{}
	request.NetworkConfiguration = &networkConfigurationObject
	networkConfigurationObject.GatewayAddress = &d.ServerGateway

	if d.ServerPrivateNetwork != "" && d.PrivateNetworking {
		var networkType = "PRIVATE_ONLY"
		request.NetworkType = &networkType
	}

	if d.ServerPrivateNetwork != "" {
		privateNetworkConfigurationObject := bmcapiclient.PrivateNetworkConfiguration{}
		var confType = "USER_DEFINED"
		privateNetworkConfigurationObject.ConfigurationType = &confType
		serPrivateNets := make([]bmcapiclient.ServerPrivateNetwork, 1)
		serverPrivateNetworkObject := bmcapiclient.ServerPrivateNetwork{}
		serverPrivateNetworkObject.Id = d.ServerPrivateNetwork

		serPrivateNets[0] = serverPrivateNetworkObject
		privateNetworkConfigurationObject.PrivateNetworks = serPrivateNets
		networkConfigurationObject.PrivateNetworkConfiguration = &privateNetworkConfigurationObject

	}

	//var userdata = []byte("#cloud-config\r\nusers:\r\n - name: root\r\n   ssh_authorized_keys:\r\n    - " + string(publicKey))

	var userdata string
	if b64, err := d.Base64UserData(); err != nil {
		return err
	} else {
		userdata = b64
	}
	cloudInitObject := bmcapiclient.OsConfigurationCloudInit{}
	cloudInitObject.UserData = &userdata
	dtoOsConfiguration := bmcapiclient.OsConfiguration{}
	dtoOsConfiguration.CloudInit = &cloudInitObject
	request.OsConfiguration = &dtoOsConfiguration

	//b, _ := json.MarshalIndent(request, "", "  ")
	//log.Info("request object is" + string(b))
	requestCommand := server.NewCreateServerCommand(*client, *request, *query)

	response, err := requestCommand.Execute()

	if err != nil {
		return err
	} else {
		/* response := &dto.LongServer{}
		response.FromBytes(resp) */
		d.ID = response.Id
		d.Name = d.MachineName
		//d.MachineName = (response.ID)
		d.PrivateIPAddresses = response.PrivateIpAddresses
		d.PublicIPAddresses = response.PublicIpAddresses
		//d.IPAddress = response.PublicIpAddresses[0]
		d.GetIP()
		d.RAM = response.Ram
		d.Storage = response.Storage
		d.CPU = response.Cpu
		d.ProvisionedOn = response.ProvisionedOn
	}

	if err := d.waitForStatus(state.Running); err != nil {
		return fmt.Errorf("wait for machine running failed: %s", err)
	}

	return nil
}

func (d *Driver) Base64UserData() (userdata string, err error) {
	if d.UserDataFile != "" {
		buf, ioerr := ioutil.ReadFile(d.UserDataFile)
		if ioerr != nil {
			log.Warnf("failed to read user data file %q: %s", d.UserDataFile, ioerr)
			err = fmt.Errorf("unable to read --pnap-userdata file")
			return
		}
		userdata = base64.StdEncoding.EncodeToString(buf)
	}
	return
}

// GetSSHHostname returns hostname for use with ssh
func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// GetIP returns IP to use in communication
func (d *Driver) GetIP() (string, error) {
	log.Debug("pnap.GetIP()")

	if d.IPAddress == "" {

		if d.PrivateNetworking {
			if len(d.PrivateIPAddresses) > 0 {
				d.IPAddress = d.PrivateIPAddresses[0]
			} else {
				return "", fmt.Errorf("private ip address not found on server, please check configuration")
			}

		} else {
			if len(d.PublicIPAddresses) > 0 {
				d.IPAddress = d.PublicIPAddresses[0]
			} else {
				return "", fmt.Errorf("public ip address not found on server, please check configuration")
			}
		}

	}

	return d.IPAddress, nil
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {

	d.setTokenToEmptySTring()
	client, err := d.getClient()
	if err != nil {
		return state.Error, err
	}
	requestCommand := server.NewGetServerCommand(*client, d.ID)
	response, err := requestCommand.Execute()

	if err != nil {
		return state.Error, err
	} else {
		//d.ID = (response.Id)
		d.Status = response.Status
		d.ProvisionedOn = response.ProvisionedOn

		switch d.Status {
		case "powered-on":
			return state.Running, nil
		case "creating",
			"resetting",
			"rebooting":
			return state.Starting, nil
		case "powered-off":
			return state.Stopped, nil
		}
		return state.None, nil
	}
}

// GetURL returns a Docker compatible host URL for connecting to this host
// e.g. tcp://1.2.3.4:2376
func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", nil
	}

	return fmt.Sprintf("tcp://%s:%d", ip, 2376), nil
}

func (d *Driver) waitForStatus(a state.State) error {
	for {
		//log.Infof("Waiting for Machine %s...", a.String())
		act, err := d.GetState()
		if err != nil {
			return errors.Wrap(err, "Could not get Server state.")
		}

		if act == a {
			log.Infof("Created pnap machine reached state %s.", a.String())
			break
		} else if act == state.Error {
			return errors.Wrap(err, "Server state could not be retrived.")
		}

		log.Infof("Waiting for Machine %s...", a.String())
		time.Sleep(10 * time.Second)
	}
	return nil
}

// Kill stops a host forcefully
func (d *Driver) Kill() error {
	log.Info("Killing pnap machine instance...")
	d.setTokenToEmptySTring()
	client, err := d.getClient()
	if err != nil {
		return err
	}

	//var requestCommand command.Executor
	requestCommand := server.NewDeleteServerCommand(*client, d.ID)
	_, err1 := requestCommand.Execute()
	if err1 != nil {
		return err1
	}
	if err := d.waitForStatus(state.Stopped); err != nil {
		return fmt.Errorf("wait for machine stopping failed: %s", err)
	}
	return err
}

// Remove a host
func (d *Driver) Remove() error {
	log.Infof("Removing pnap machine instance with id %s", d.ID)
	d.setTokenToEmptySTring()
	if d.ID == "" {
		return nil
	}
	client, err := d.getClient()
	if err != nil {
		return err
	}

	requestCommand := server.NewDeleteServerCommand(*client, d.ID)
	resp, err1 := requestCommand.Execute()
	if err1 != nil {
		return err1
	}

	log.Infof("Removing pnap machine instance with id %s returned result %s", d.ID, resp.Result)

	return nil
}

// Restart a host. This may just call Stop(); Start() if the provider does not
// have any special restart behaviour.
func (d *Driver) Restart() error {
	log.Info("Rebooting pnap machine instance...")
	d.setTokenToEmptySTring()
	client, err := d.getClient()
	if err != nil {
		return err
	}

	requestCommand := server.NewRebootServerCommand(*client, d.ID)
	_, err1 := requestCommand.Execute()
	if err1 != nil {
		return err1
	}
	if err := d.waitForStatus(state.Running); err != nil {
		return fmt.Errorf("wait for machine reboot failed: %s", err)
	}
	return err
}

// Start a host
func (d *Driver) Start() error {
	log.Info("Starting pnap machine instance...")
	d.setTokenToEmptySTring()
	client, err := d.getClient()
	if err != nil {
		return err
	}

	requestCommand := server.NewPowerOnServerCommand(*client, d.ID)
	_, err1 := requestCommand.Execute()
	if err1 != nil {
		return err1
	}
	if err := d.waitForStatus(state.Running); err != nil {
		return fmt.Errorf("wait for machine to start failed: %s", err)
	}
	return err
}

// Stop a host gracefully
func (d *Driver) Stop() error {
	log.Info("Stopping pnap machine instance...")
	d.setTokenToEmptySTring()
	client, err := d.getClient()
	if err != nil {
		return err
	}

	requestCommand := server.NewShutDownServerCommand(*client, d.ID)
	_, err1 := requestCommand.Execute()
	if err1 != nil {
		return err1
	}
	if err := d.waitForStatus(state.Stopped); err != nil {
		return fmt.Errorf("wait for machine to shut down failed: %s", err)
	}
	return err
}

/* func run(command command.Executor) error {
	resp, err := command.Execute()
	if err != nil {
		return err
	}
	code := resp.StatusCode
	if code != 200 {
		response := &dto.ErrorMessage{}
		response.FromBytes(resp)
		return fmt.Errorf("API Returned Code: %v, Message: %v, Validation Errors: %v", code, response.Message, response.ValidationErrors)
	}
	return nil
} */

// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
func (d *Driver) PreCreateCheck() error {
	//////////////////////////////////////////////////////////////////////////////////////////////////////////////
	//log.Infof("Driver params private network:%s server private network:%s gateway:%s os:%s", d.PrivateNetworking, d.ServerPrivateNetwork, d.ServerGateway, d.ServerOs)
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////
	if d.ServerLocation == "" {
		log.Info("Location has not been set, will be used PHX as default location.")
		d.ServerLocation = defaultLocation
	}
	if d.ServerType == "" {
		log.Info("Type has not been set, will be used s1.c1.medium as default type.")
		d.ServerType = defaultType
	}
	if d.ServerOs == "" {
		log.Info("OS has not been set, will be used ubuntu/bionic as default type.")
		d.ServerOs = defaultOS
	}

	client, err := d.getClient()
	if err != nil {
		return err
	}
	query := helperdto.ProductAvailabilityQuery{}
	proCod := make([]string, 1)
	proCod[0] = fmt.Sprint(d.ServerType)
	query.ProductCode = proCod

	location := make([]string, 1)
	location[0] = fmt.Sprint(d.ServerLocation)
	query.Location = location

	category := make([]string, 1)
	category[0] = "SERVER"
	query.ProductCategory = category

	query.MinQuantity = 1

	requestCommand := product.NewGetProductAvailabilityCommand(*client, query)
	resp, err := requestCommand.Execute()
	if err != nil {
		return err
	}

	if len(resp) <= 0 {
		return fmt.Errorf("no servers of type: %s available in location: %s", d.ServerType, d.ServerLocation)
	}

	if len(resp[0].LocationAvailabilityDetails) < 1 || resp[0].LocationAvailabilityDetails[0].AvailableQuantity < 1 {
		return fmt.Errorf("no servers of type: %s available in location: %s", d.ServerType, d.ServerLocation)
	}

	return nil
}

// GetSSHUsername returns username for use with ssh
func (d *Driver) GetSSHUsername() string {

	if strings.Contains(d.ServerOs, "ubuntu") {
		d.SSHUser = "ubuntu"
	} else if strings.Contains(d.ServerOs, "centos") {
		d.SSHUser = "centos"
	} else if strings.Contains(d.ServerOs, "windows") {
		d.SSHUser = "Admin"
	}

	return d.SSHUser
}

// setTokenToEmptySTring invalidates token.
// Token is definitelly expired after one hour, and this method enables other ways of authentication.
func (d *Driver) setTokenToEmptySTring() {

	if d.ProvisionedOn != nil && d.ProvisionedOn.Add(time.Minute*60).Before(time.Now()) {
		log.Info("Bearer token invalidated.")
		d.BearerToken = ""
	}
}

func (d *Driver) isTokenValid() bool {
	token, _, errjj := new(jwt.Parser).ParseUnverified(d.BearerToken, jwt.MapClaims{})
	if errjj != nil {
		log.Info("Error happened when validating token expiration ", errjj)
		return false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Info("Can't convert token's claims to standard claims.")
		return false
	}

	var tm time.Time
	switch iat := claims["exp"].(type) {
	case float64:
		tm = time.Unix(int64(iat), 0)
	case json.Number:
		v, _ := iat.Int64()
		tm = time.Unix(v, 0)
	}
	var now = time.Now()
	if tm.Before(now) {
		log.Info("Token expired. Use cloud credentials.")
		return false
	} else {
		//log.Info("Token is not expired. Try to use it for authentication.")
		return true
	}
}
