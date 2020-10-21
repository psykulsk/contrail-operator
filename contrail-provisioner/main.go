package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"gopkg.in/yaml.v2"

	contrail "github.com/Juniper/contrail-go-api"
	"github.com/Juniper/contrail-operator/contrail-provisioner/analyticsnodes"
	"github.com/Juniper/contrail-operator/contrail-provisioner/confignodes"
	contrailTypes "github.com/Juniper/contrail-operator/contrail-provisioner/contrail-go-types"
	"github.com/Juniper/contrail-operator/contrail-provisioner/controlnodes"
	"github.com/Juniper/contrail-operator/contrail-provisioner/databasenodes"
	"github.com/Juniper/contrail-operator/contrail-provisioner/vrouternodes"
)

// APIServer struct contains API Server configuration
type APIServer struct {
	APIPort       string     `yaml:"apiPort,omitempty"`
	APIServerList []string   `yaml:"apiServerList,omitempty"`
	Encryption    encryption `yaml:"encryption,omitempty"`
}

type encryption struct {
	CA       string `yaml:"ca,omitempty"`
	Cert     string `yaml:"cert,omitempty"`
	Key      string `yaml:"key,omitempty"`
	Insecure bool   `yaml:"insecure,omitempty"`
}

type KeystoneAuthParameters struct {
	AdminUsername string     `yaml:"admin_user,omitempty"`
	AdminPassword string     `yaml:"admin_password,omitempty"`
	AuthUrl       string     `yaml:"auth_url,omitempty"`
	TenantName    string     `yaml:"tenant_name,omitempty"`
	Encryption    encryption `yaml:"encryption,omitempty"`
}

type EcmpHashingIncludeFields struct {
	HashingConfigured bool `json:"hashingConfigured,omitempty"`
	SourceIp          bool `json:"sourceIp,omitempty"`
	DestinationIp     bool `json:"destinationIp,omitempty"`
	IpProtocol        bool `json:"ipProtocol,omitempty"`
	SourcePort        bool `json:"sourcePort,omitempty"`
	DestinationPort   bool `json:"destinationPort,omitempty"`
}

type GlobalVrouterConfiguration struct {
	EcmpHashingIncludeFields   EcmpHashingIncludeFields `json:"ecmpHashingIncludeFields,omitempty"`
	EncapsulationPriorities    string                   `json:"encapPriority,omitempty"`
	VxlanNetworkIdentifierMode string                   `json:"vxlanNetworkIdentifierMode,omitempty"`
}

func nodeManager(nodesPtr *string, nodeType string, contrailClient *contrail.Client) {
	log.Printf("%s %s updated\n", nodeType, *nodesPtr)
	nodeYaml, err := ioutil.ReadFile(*nodesPtr)
	if err != nil {
		panic(err)
	}
	switch nodeType {
	case "control":
		var nodeList []*controlnodes.ControlNode
		err = yaml.Unmarshal(nodeYaml, &nodeList)
		if err != nil {
			panic(err)
		}
		if err = controlnodes.ReconcileControlNodes(contrailClient, nodeList); err != nil {
			panic(err)
		}
	case "analytics":
		var nodeList []*analyticsnodes.AnalyticsNode
		err = yaml.Unmarshal(nodeYaml, &nodeList)
		if err != nil {
			panic(err)
		}
		if err = analyticsnodes.ReconcileAnalyticsNodes(contrailClient, nodeList); err != nil {
			panic(err)
		}
	case "config":
		var nodeList []*confignodes.ConfigNode
		err = yaml.Unmarshal(nodeYaml, &nodeList)
		if err != nil {
			panic(err)
		}
		if err = confignodes.ReconcileConfigNodes(contrailClient, nodeList); err != nil {
			panic(err)
		}
	case "vrouter":
		var nodeList []*vrouternodes.VrouterNode
		err = yaml.Unmarshal(nodeYaml, &nodeList)
		if err != nil {
			panic(err)
		}
		if err = vrouternodes.ReconcileVrouterNodes(contrailClient, nodeList); err != nil {
			panic(err)
		}
	case "database":
		var nodeList []*databasenodes.DatabaseNode
		err = yaml.Unmarshal(nodeYaml, &nodeList)
		if err != nil {
			panic(err)
		}
		if err = databasenodes.ReconcileDatabaseNodes(contrailClient, nodeList); err != nil {
			panic(err)
		}
	}
}

func check(err error) {
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}

func main() {

	controlNodesPtr := flag.String("controlNodes", "/provision.yaml", "path to control nodes yaml file")
	configNodesPtr := flag.String("configNodes", "/provision.yaml", "path to config nodes yaml file")
	analyticsNodesPtr := flag.String("analyticsNodes", "/provision.yaml", "path to analytics nodes yaml file")
	vrouterNodesPtr := flag.String("vrouterNodes", "/provision.yaml", "path to vrouter nodes yaml file")
	databaseNodesPtr := flag.String("databaseNodes", "/provision.yaml", "path to database nodes yaml file")
	apiserverPtr := flag.String("apiserver", "/provision.yaml", "path to apiserver yaml file")
	keystoneAuthConfPtr := flag.String("keystoneAuthConf", "/provision.yaml", "path to keystone authentication configuration file")
	globalVrouterConfPtr := flag.String("globalVrouterConf", "/provision.yaml", "path to global vrouter configuration file")
	modePtr := flag.String("mode", "watch", "watch/run")
	flag.Parse()

	if *modePtr == "watch" {

		var apiServer APIServer
		apiServerYaml, err := ioutil.ReadFile(*apiserverPtr)
		if err != nil {
			panic(err)
		}
		err = yaml.Unmarshal(apiServerYaml, &apiServer)
		if err != nil {
			panic(err)
		}

		var keystoneAuthParameters *KeystoneAuthParameters = &KeystoneAuthParameters{}
		if _, err := os.Stat(*keystoneAuthConfPtr); err == nil {
			keystoneAuthParameters = getKeystoneAuthParametersFromFile(*keystoneAuthConfPtr)
		}

		var contrailClient *contrail.Client
		err = retry(5, 10*time.Second, func() (err error) {
			contrailClient, err = getAPIClient(&apiServer, keystoneAuthParameters)
			return

		})
		if err != nil {
			if !connectionError(err) {
				panic(err)
			}
		}

		globalVrouterConfiguration := &GlobalVrouterConfiguration{}
		if _, err := os.Stat(*globalVrouterConfPtr); err == nil {
			globalVrouterConfiguration = getGlobalVrouterConfigFromFile(*globalVrouterConfPtr)
		}
		globalVrouterConfFQName := []string{"default-global-system-config", "default-global-vrouter-config"}
		encapPriority := strings.Split(globalVrouterConfiguration.EncapsulationPriorities, ",")
		encapPriorityObj := &contrailTypes.EncapsulationPrioritiesType{Encapsulation: encapPriority}
		ecmpObj := globalVrouterConfiguration.EcmpHashingIncludeFields
		ecmpHashingIncludeFieldsObj := &contrailTypes.EcmpHashingIncludeFields{ecmpObj.HashingConfigured, ecmpObj.SourceIp, ecmpObj.DestinationIp, ecmpObj.IpProtocol, ecmpObj.SourcePort, ecmpObj.DestinationPort}
		GlobalVrouterConfig := &contrailTypes.GlobalVrouterConfig{}
		GlobalVrouterConfig.SetFQName("", globalVrouterConfFQName)
		GlobalVrouterConfig.SetEncapsulationPriorities(encapPriorityObj)
		GlobalVrouterConfig.SetEcmpHashingIncludeFields(ecmpHashingIncludeFieldsObj)
		GlobalVrouterConfig.SetVxlanNetworkIdentifierMode(globalVrouterConfiguration.VxlanNetworkIdentifierMode)
		if err = contrailClient.Create(GlobalVrouterConfig); err != nil {
			if !strings.Contains(err.Error(), "409 Conflict") {
				panic(err)
			}
			obj, err := contrailClient.FindByName("global-vrouter-config", strings.Join(globalVrouterConfFQName, ":"))
			if err != nil {
				panic(err)
			}
			obj.(*contrailTypes.GlobalVrouterConfig).SetEncapsulationPriorities(encapPriorityObj)
			obj.(*contrailTypes.GlobalVrouterConfig).SetEcmpHashingIncludeFields(ecmpHashingIncludeFieldsObj)
			obj.(*contrailTypes.GlobalVrouterConfig).SetVxlanNetworkIdentifierMode(globalVrouterConfiguration.VxlanNetworkIdentifierMode)
			if err = contrailClient.Update(obj); err != nil {
				panic(err)
			}
		}

		log.Println("start watcher")
		done := make(chan bool)

		if controlNodesPtr != nil {
			log.Println("initial control node run")
			_, err := os.Stat(*controlNodesPtr)
			if !os.IsNotExist(err) {
				nodeManager(controlNodesPtr, "control", contrailClient)
			} else if os.IsNotExist(err) {
				controlnodes.ReconcileControlNodes(contrailClient, []*controlnodes.ControlNode{})
			}
			log.Println("setting up control node watcher")
			watchFile := strings.Split(*controlNodesPtr, "/")
			watchPath := strings.TrimSuffix(*controlNodesPtr, watchFile[len(watchFile)-1])
			nodeWatcher, err := WatchFile(watchPath, time.Second, func() {
				log.Println("control node event")
				_, err := os.Stat(*controlNodesPtr)
				if !os.IsNotExist(err) {
					nodeManager(controlNodesPtr, "control", contrailClient)
				} else if os.IsNotExist(err) {
					controlnodes.ReconcileControlNodes(contrailClient, []*controlnodes.ControlNode{})
				}
			})
			check(err)

			defer func() {
				nodeWatcher.Close()
			}()
		}

		if vrouterNodesPtr != nil {
			log.Println("initial vrouter node run")
			_, err := os.Stat(*vrouterNodesPtr)
			if !os.IsNotExist(err) {
				nodeManager(vrouterNodesPtr, "vrouter", contrailClient)
			} else if os.IsNotExist(err) {
				vrouternodes.ReconcileVrouterNodes(contrailClient, []*vrouternodes.VrouterNode{})
			}
			log.Println("setting up vrouter node watcher")
			watchFile := strings.Split(*vrouterNodesPtr, "/")
			watchPath := strings.TrimSuffix(*vrouterNodesPtr, watchFile[len(watchFile)-1])
			nodeWatcher, err := WatchFile(watchPath, time.Second, func() {
				log.Println("vrouter node event")
				_, err := os.Stat(*vrouterNodesPtr)
				if !os.IsNotExist(err) {
					nodeManager(vrouterNodesPtr, "vrouter", contrailClient)
				} else if os.IsNotExist(err) {
					vrouternodes.ReconcileVrouterNodes(contrailClient, []*vrouternodes.VrouterNode{})
				}
			})
			check(err)

			defer func() {
				nodeWatcher.Close()
			}()
		}

		if analyticsNodesPtr != nil {
			log.Println("initial analytics node run")
			_, err := os.Stat(*analyticsNodesPtr)
			if !os.IsNotExist(err) {
				nodeManager(analyticsNodesPtr, "analytics", contrailClient)
			} else if os.IsNotExist(err) {
				analyticsnodes.ReconcileAnalyticsNodes(contrailClient, []*analyticsnodes.AnalyticsNode{})
			}
			log.Println("setting up analytics node watcher")
			watchFile := strings.Split(*analyticsNodesPtr, "/")
			watchPath := strings.TrimSuffix(*analyticsNodesPtr, watchFile[len(watchFile)-1])
			nodeWatcher, err := WatchFile(watchPath, time.Second, func() {
				log.Println("analytics node event")
				_, err := os.Stat(*analyticsNodesPtr)
				if !os.IsNotExist(err) {
					nodeManager(analyticsNodesPtr, "analytics", contrailClient)
				} else if os.IsNotExist(err) {
					analyticsnodes.ReconcileAnalyticsNodes(contrailClient, []*analyticsnodes.AnalyticsNode{})
				}
			})
			check(err)

			defer func() {
				nodeWatcher.Close()
			}()
		}

		if configNodesPtr != nil {
			log.Println("initial config node run")
			_, err := os.Stat(*configNodesPtr)
			if !os.IsNotExist(err) {
				nodeManager(configNodesPtr, "config", contrailClient)
			} else if os.IsNotExist(err) {
				confignodes.ReconcileConfigNodes(contrailClient, []*confignodes.ConfigNode{})
			}
			log.Println("setting up config node watcher")
			watchFile := strings.Split(*configNodesPtr, "/")
			watchPath := strings.TrimSuffix(*configNodesPtr, watchFile[len(watchFile)-1])
			nodeWatcher, err := WatchFile(watchPath, time.Second, func() {
				log.Println("config node event")
				_, err := os.Stat(*configNodesPtr)
				if !os.IsNotExist(err) {
					nodeManager(configNodesPtr, "config", contrailClient)
				} else if os.IsNotExist(err) {
					confignodes.ReconcileConfigNodes(contrailClient, []*confignodes.ConfigNode{})
				}
			})
			check(err)

			defer func() {
				nodeWatcher.Close()
			}()
		}

		if databaseNodesPtr != nil {
			log.Println("initial database node run")
			_, err := os.Stat(*databaseNodesPtr)
			if !os.IsNotExist(err) {
				nodeManager(databaseNodesPtr, "database", contrailClient)
			} else if os.IsNotExist(err) {
				databasenodes.ReconcileDatabaseNodes(contrailClient, []*databasenodes.DatabaseNode{})
			}
			log.Println("setting up database node watcher")
			watchFile := strings.Split(*databaseNodesPtr, "/")
			watchPath := strings.TrimSuffix(*databaseNodesPtr, watchFile[len(watchFile)-1])
			nodeWatcher, err := WatchFile(watchPath, time.Second, func() {
				log.Println("database node event")
				_, err := os.Stat(*databaseNodesPtr)
				if !os.IsNotExist(err) {
					nodeManager(databaseNodesPtr, "database", contrailClient)
				} else if os.IsNotExist(err) {
					databasenodes.ReconcileDatabaseNodes(contrailClient, []*databasenodes.DatabaseNode{})
				}
			})
			check(err)

			defer func() {
				nodeWatcher.Close()
			}()
		}

		<-done
	}

	if *modePtr == "run" {

		var apiServer APIServer

		apiServerYaml, err := ioutil.ReadFile(*apiserverPtr)
		if err != nil {
			panic(err)
		}
		err = yaml.Unmarshal(apiServerYaml, &apiServer)
		if err != nil {
			panic(err)
		}

		var keystoneAuthParameters *KeystoneAuthParameters = &KeystoneAuthParameters{}
		if _, err := os.Stat(*keystoneAuthConfPtr); err == nil {
			keystoneAuthParameters = getKeystoneAuthParametersFromFile(*keystoneAuthConfPtr)
		}

		contrailClient, err := getAPIClient(&apiServer, keystoneAuthParameters)
		if err != nil {
			panic(err.Error())
		}

		if controlNodesPtr != nil {
			var controlNodeList []*controlnodes.ControlNode
			controlNodeYaml, err := ioutil.ReadFile(*controlNodesPtr)
			if err != nil {
				panic(err)
			}
			err = yaml.Unmarshal(controlNodeYaml, &controlNodeList)
			if err != nil {
				panic(err)
			}
			err = retry(5, 10*time.Second, func() (err error) {
				err = controlnodes.ReconcileControlNodes(contrailClient, controlNodeList)
				return
			})
			if err != nil {
				panic(err)
			}
		}

		if configNodesPtr != nil {
			var configNodeList []*confignodes.ConfigNode
			configNodeYaml, err := ioutil.ReadFile(*configNodesPtr)
			if err != nil {
				panic(err)
			}
			err = yaml.Unmarshal(configNodeYaml, &configNodeList)
			if err != nil {
				panic(err)
			}
			if err = confignodes.ReconcileConfigNodes(contrailClient, configNodeList); err != nil {
				panic(err)
			}
		}

		if analyticsNodesPtr != nil {
			var analyticsNodeList []*analyticsnodes.AnalyticsNode
			analyticsNodeYaml, err := ioutil.ReadFile(*analyticsNodesPtr)
			if err != nil {
				panic(err)
			}
			err = yaml.Unmarshal(analyticsNodeYaml, &analyticsNodeList)
			if err != nil {
				panic(err)
			}
			if err = analyticsnodes.ReconcileAnalyticsNodes(contrailClient, analyticsNodeList); err != nil {
				panic(err)
			}
		}

		if vrouterNodesPtr != nil {
			var vrouterNodeList []*vrouternodes.VrouterNode
			vrouterNodeYaml, err := ioutil.ReadFile(*vrouterNodesPtr)
			if err != nil {
				panic(err)
			}
			err = yaml.Unmarshal(vrouterNodeYaml, &vrouterNodeList)
			if err != nil {
				panic(err)
			}
			if err = vrouternodes.ReconcileVrouterNodes(contrailClient, vrouterNodeList); err != nil {
				panic(err)
			}
		}

		if databaseNodesPtr != nil {
			var databaseNodeList []*databasenodes.DatabaseNode
			databaseNodeYaml, err := ioutil.ReadFile(*databaseNodesPtr)
			if err != nil {
				panic(err)
			}
			err = yaml.Unmarshal(databaseNodeYaml, &databaseNodeList)
			if err != nil {
				panic(err)
			}
			if err = databasenodes.ReconcileDatabaseNodes(contrailClient, databaseNodeList); err != nil {
				panic(err)
			}
		}

	}
}
func retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; ; i++ {
		err = f()
		if err == nil {
			return
		}
		if attempts != 0 {
			if i >= (attempts - 1) {
				break
			}
		}

		time.Sleep(sleep)

		log.Println("retrying after error:", err)
	}
	return err
}

func connectionError(err error) bool {
	if err == nil {
		log.Println("Ok")
		return false

	} else if netError, ok := err.(net.Error); ok && netError.Timeout() {
		log.Println("Timeout")
		return true
	}
	unwrappedError := errors.Unwrap(err)
	switch t := unwrappedError.(type) {
	case *net.OpError:
		if t.Op == "dial" {
			log.Println("Unknown host")
			return true
		} else if t.Op == "read" {
			log.Println("Connection refused")
			return true
		}

	case syscall.Errno:
		if t == syscall.ECONNREFUSED {
			log.Println("Connection refused")
			return true
		}

	default:
		log.Println(t)
	}
	return false
}

func getAPIClient(apiServerObj *APIServer, keystoneAuthParameters *KeystoneAuthParameters) (*contrail.Client, error) {
	var contrailClient *contrail.Client
	for _, apiServer := range apiServerObj.APIServerList {
		apiServerSlice := strings.Split(apiServer, ":")
		apiPortInt, err := strconv.Atoi(apiServerSlice[1])
		if err != nil {
			return contrailClient, err
		}
		log.Printf("api server %s:%d\n", apiServerSlice[0], apiPortInt)
		contrailClient := contrail.NewClient(apiServerSlice[0], apiPortInt)
		err = contrailClient.AddEncryption(apiServerObj.Encryption.CA, apiServerObj.Encryption.Key, apiServerObj.Encryption.Cert, true)
		if err != nil {
			return nil, err
		}
		if keystoneAuthParameters.AuthUrl != "" {
			setupAuthKeystone(contrailClient, keystoneAuthParameters)
		}
		//contrailClient.AddHTTPParameter(1)
		_, err = contrailClient.List("global-system-config")
		if err == nil {
			return contrailClient, nil
		}
	}
	return contrailClient, fmt.Errorf("%s", "cannot get api server")

}

func setupAuthKeystone(client *contrail.Client, keystoneAuthParameters *KeystoneAuthParameters) {
	// AddEncryption expected http url in older versions of contrail-go-api
	// https://github.com/Juniper/contrail-go-api/commit/4c876ba038a8ecec211376133375d467b6098202
	var authUrl string
	if strings.HasPrefix(keystoneAuthParameters.AuthUrl, "https") {
		authUrl = strings.Replace(keystoneAuthParameters.AuthUrl, "https", "http", 1)
	} else {
		authUrl = keystoneAuthParameters.AuthUrl
	}
	keystone := contrail.NewKeepaliveKeystoneClient(
		authUrl,
		keystoneAuthParameters.TenantName,
		keystoneAuthParameters.AdminUsername,
		keystoneAuthParameters.AdminPassword,
		"",
	)
	err := keystone.AddEncryption(
		keystoneAuthParameters.Encryption.CA,
		keystoneAuthParameters.Encryption.Key,
		keystoneAuthParameters.Encryption.Cert,
		keystoneAuthParameters.Encryption.Insecure)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	err = keystone.AuthenticateV3()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	client.SetAuthenticator(keystone)

}

func getKeystoneAuthParametersFromFile(authParamsFilePath string) *KeystoneAuthParameters {
	var keystoneAuthParameters *KeystoneAuthParameters
	keystoneAuthYaml, err := ioutil.ReadFile(authParamsFilePath)
	if err != nil {
		panic(err)
	}
	if err = yaml.Unmarshal(keystoneAuthYaml, &keystoneAuthParameters); err != nil {
		panic(err)
	}
	return keystoneAuthParameters
}

func getGlobalVrouterConfigFromFile(globalVrouterFilePath string) *GlobalVrouterConfiguration {
	var globalVrouterConfig *GlobalVrouterConfiguration
	globalVrouterJson, err := ioutil.ReadFile(globalVrouterFilePath)
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal([]byte(globalVrouterJson), &globalVrouterConfig); err != nil {
		panic(err)
	}
	return globalVrouterConfig
}
