package blockchain

import (
	"fmt"
	"strings"
	"os"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/status"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/cryptosuite"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
	identityconfigmsp "github.com/hyperledger/fabric-sdk-go/pkg/msp"
	contextAPI "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	contextImpl "github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/pkg/errors"
)

// FabricSetup implementation
type FabricSetup struct {
	OrdererAdmin string
	OrdererOrgName   string
	ConfigFile      string
	OrdererID       string
	OrgmspID        string
	ChannelID       string
	ChannelListID[]       string
	ChainCodeID     string
	ChainCodeListID[]     string
	initialized     bool
//	ChannelConfig   string
//	ChannelListConfig[]   string
	ChaincodeGoPath string
	ChaincodePath   string
	Org_nowAdmin        string	
	Org_nowUser       string
	Org4Admin        string
	Org_now        string
	client          *channel.Client
	org_nowResMgmt     *resmgmt.Client
	orgTestPeernow    fab.Peer
	org_nowAdminClientContext contextAPI.ClientProvider
	sdk             *fabsdk.FabricSDK
	event           *event.Client
}


// CleanupTestPath removes the contents of a state store.
func CleanupTestPath(storePath string) {
	err := os.RemoveAll(storePath)
	if err != nil {
		//if t == nil {
		//	panic(fmt.Sprintf("Cleaning up directory '%s' failed: %v", storePath, err))
		//}
		//t.Fatalf("Cleaning up directory '%s' failed: %v", storePath, err)
		//panic(fmt.Sprintf("Cleaning up directory '%s' failed: %v", storePath, err))
		fmt.Println("Cleaning up directory '%s' failed: %v", storePath, err)
	}
	fmt.Println("Cleaning up directory '%s' successful", storePath)
}

// CleanupUserData removes user data.
func CleanupUserData(sdk *fabsdk.FabricSDK) {
	var keyStorePath, credentialStorePath string

	configBackend, err := sdk.Config()
	if err != nil {
		// if an error is returned from Config, it means configBackend was nil, in this case simply hard code
		// the keyStorePath and credentialStorePath to the default values
		// This case is mostly happening due to configless test that is not passing a ConfigProvider to the SDK
		// which makes configBackend = nil.
		// Since configless test uses the same config values as the default ones (config_test.yaml), it's safe to
		// hard code these paths here
		keyStorePath = "/tmp/app-1g4p0-store"
		credentialStorePath = "/tmp/app-1g4p0-msp"
		fmt.Println("Cleaning up if directory keyStorePath:%s, credentialStorePath:%s", keyStorePath,credentialStorePath)
	} else {
		cryptoSuiteConfig := cryptosuite.ConfigFromBackend(configBackend)
		identityConfig, err := identityconfigmsp.ConfigFromBackend(configBackend)
		if err != nil {
			fmt.Println("CleanupUserData removes user data fail:%s", err)
			//t.Fatal(err)
		}

		keyStorePath = cryptoSuiteConfig.KeyStorePath()
		credentialStorePath = identityConfig.CredentialStorePath()
		fmt.Println("Cleaning up else directory keyStorePath:%s, credentialStorePath:%s", keyStorePath,credentialStorePath)
	}

	CleanupTestPath(keyStorePath)
	CleanupTestPath(credentialStorePath)
}

func WaitForOrdererConfigUpdate(client *resmgmt.Client, channelID string, genesis bool, lastConfigBlock uint64) uint64 {

	blockNum, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			chConfig, err := client.QueryConfigFromOrderer(channelID, resmgmt.WithOrdererEndpoint("orderer.hf.srezd.io"))
			if err != nil {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), err.Error(), nil)
			}

			currentBlock := chConfig.BlockNumber()
			if currentBlock <= lastConfigBlock && !genesis {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("Block number was not incremented [%d, %d]", currentBlock, lastConfigBlock), nil)
			}
			return &currentBlock, nil
		},
	)
	fmt.Println("WaitForOrdererConfigUpdate: %s", err)
	//require.NoError(err)
	return *blockNum.(*uint64)
}


// Initialize reads the configuration file and sets up the client, chain and event hub
func (setup *FabricSetup) Initialize() error {

	// Add parameters for the initialization
	if setup.initialized {
		return errors.New("sdk already initialized")
	}

	// Initialize the SDK with the configuration file
	sdk, err := fabsdk.New(config.FromFile(setup.ConfigFile))
	if err != nil {
		return errors.WithMessage(err, "failed to create SDK")
	}
	setup.sdk = sdk
	fmt.Println("SDK created")

	org_nowMspClient, err := mspclient.New(setup.sdk.Context(), mspclient.WithOrg(setup.Org_now))
	if err != nil {
		return errors.WithMessage(err, "failed to create org_nowMspClient")
	}else{
		fmt.Println("org_nowMspClient successful")
	}

	org_nowAdminIdentity, err := org_nowMspClient.GetSigningIdentity(setup.Org_nowAdmin)
	if err != nil {
		return errors.WithMessage(err, "failed to get org_nowAdminIdentity")
	}else{
		fmt.Println("org_nowAdminIdentity successly.")
	}
	//资源管理客户负责管理通道
	//ctx, err := ctxProvider()
	ctx, err := sdk.Context(fabsdk.WithUser(setup.Org_nowAdmin), fabsdk.WithOrg(setup.Org_now))()
	if err != nil {
		return errors.WithMessage(err, "context creation failed")
	}

	org_nowPeers, ok := ctx.EndpointConfig().PeersConfig(setup.Org_now)
	if ok {
		fmt.Println("org_now ok:ture")
	}else{
		fmt.Println("org_now ok:false")
	}
	orgTestPeernow, err := ctx.InfraProvider().CreatePeerFromConfig(&fab.NetworkPeer{PeerConfig: org_nowPeers[0]})
	if err != nil {
		return errors.WithMessage(err, "CreatePeerFromConfig org_nowPeers failed")
	}else{
		fmt.Println("orgTestPeernow:",orgTestPeernow)
	}
	setup.orgTestPeernow = orgTestPeernow

	//删除加密套件store的所有私钥以及终端用户（utils.go）清除环境和用户数据
	CleanupUserData(setup.sdk)
	defer CleanupUserData(setup.sdk)

	setup.org_nowAdminClientContext = setup.sdk.Context(fabsdk.WithUser(setup.Org_nowAdmin), fabsdk.WithOrg(setup.Org_now))
	if err != nil {
		return errors.WithMessage(err, "failed to load org_nowRes Admin identity")
	}
	org_nowResMgmtClient, err := resmgmt.New(setup.org_nowAdminClientContext)
	if err != nil {
		return errors.WithMessage(err, "failed to org_nowResMgmtClient client from Admin identity")
	}
	setup.org_nowResMgmt = org_nowResMgmtClient
	fmt.Println("Ressource org_nowResMgmt client created")

	configQueryClient, err := resmgmt.New(setup.org_nowAdminClientContext)
	if err != nil {
		return errors.WithMessage(err, "failed to configQueryClient client from Admin identity")
	}else{
		fmt.Println("configQueryClient successful")
	}
	
	channelHasInstall := false
	// 查询已经存在的channel
	channelRes, err := setup.org_nowResMgmt.QueryChannels(resmgmt.WithTargets(setup.orgTestPeernow))
	//channelRes, err := setup.org_nowResMgmt.QueryChannels(resmgmt.WithTargetEndpoints("peer0.org4.hf.srezd.io"))
	if err != nil {
		return errors.WithMessage(err, "failed to Query channel")
	}

	if channelRes != nil {
		for _, channel := range channelRes.Channels {
			if strings.EqualFold(setup.ChannelID, channel.ChannelId) {
				channelHasInstall = true
			}
		}
	}

	fmt.Println("channelHasInstall:", channelHasInstall)

	if !channelHasInstall {

		// 如果没有同名的channel 再安装
		var lastConfigBlock uint64
		lastConfigBlock = WaitForOrdererConfigUpdate(configQueryClient,setup.ChannelID, false, lastConfigBlock)
		req := resmgmt.SaveChannelRequest{ChannelID: setup.ChannelID, ChannelConfigPath: os.Getenv("GOPATH") + "/src/github.com/srezd/app-1g4p0/fixtures/artifacts/org4.srezd.anchors.tx", 
		SigningIdentities: []msp.SigningIdentity{org_nowAdminIdentity}}
		txID, err := setup.org_nowResMgmt.SaveChannel(req, resmgmt.WithOrdererEndpoint(setup.OrdererID))
		if err != nil || txID.TransactionID == "" {
			return errors.WithMessage(err, "org_nowResMgmt failed to save channel")
		}
		fmt.Println("org_nowResMgmt Channel created")
		if err = setup.org_nowResMgmt.JoinChannel(setup.ChannelID, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint(setup.OrdererID)); err != nil {
			return errors.WithMessage(err, "org_nowResMgmt failed to make admin join channel")
		}else{
			fmt.Println("org_nowResMgmt Channel joined successful")
		}
	} else {
		fmt.Println("Channel already exist")
	}
	fmt.Println("Initialization Successful")
	setup.initialized = true

	return nil
}
func isCCInstalled(orgID string, resMgmt *resmgmt.Client, ccName, ccVersion string, peers []fab.Peer) bool {
	installedOnAllPeers := true
	for _, peer := range peers {
		resp, err := resMgmt.QueryInstalledChaincodes(resmgmt.WithTargets(peer))
		if err != nil {
			fmt.Println("QueryInstalledChaincodes for peer url fail")
			return false
		}else{
			fmt.Println("QueryInstalledChaincodes for peer url successful")
		}
		found := false
		for _, ccInfo := range resp.Chaincodes {
			if ccInfo.Name == ccName && ccInfo.Version == ccVersion {
				found = true
				break
			}
		}
		if !found {
			installedOnAllPeers = false
		}
	}
	return installedOnAllPeers
}

func DiscoverLocalPeers(ctxProvider contextAPI.ClientProvider, expectedPeers int) ([]fab.Peer, error) {
	ctx, err := contextImpl.NewLocal(ctxProvider)
	if err != nil {
		return nil, errors.Wrap(err, "error creating local context")
	}

	discoveredPeers, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			peers, err := ctx.LocalDiscoveryService().GetPeers()
			if err != nil {
				return nil, errors.Wrapf(err, "error getting peers for MSP [%s]", ctx.Identifier().MSPID)
			}
			if len(peers) < expectedPeers {
				return nil, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("Expecting %d peers but got %d", expectedPeers, len(peers)), nil)
			}
			return peers, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return discoveredPeers.([]fab.Peer), nil
}

func queryInstalledCC(orgID string, resMgmt *resmgmt.Client, ccName, ccVersion string, peers []fab.Peer) bool {
	installed, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			ok := isCCInstalled(orgID, resMgmt, ccName, ccVersion, peers)
			if !ok {
				return &ok, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("Chaincode [%s:%s] is not installed on all peers in Org_now", ccName, ccVersion), nil)
			}
			return &ok, nil
		},
	)
	if err != nil {
		fmt.Println("queryInstalledCC for peer fail")
		return *(installed).(*bool)
	}else{
		fmt.Println("queryInstalledCC for peer successful")
	}
	return *(installed).(*bool)
}

func queryInstantiatedCC(orgID string, resMgmt *resmgmt.Client, channelID, ccName, ccVersion string, peers []fab.Peer) bool {
	instantiated, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			ok := isCCInstantiated(resMgmt, channelID, ccName, ccVersion, peers)
			if !ok {
				return &ok, status.New(status.TestStatus, status.GenericTransient.ToInt32(), fmt.Sprintf("Did NOT find instantiated chaincode [%s:%s] on one or more peers in [%s].", ccName, ccVersion, orgID), nil)
			}
			return &ok, nil
		},
	)
	if err != nil {
		fmt.Println("queryInstantiatedCC for peer fail")
		return *(instantiated).(*bool)
	}else{
		fmt.Println("queryInstantiatedCC for peer successful")
	}
	return *(instantiated).(*bool)
}

func isCCInstantiated(resMgmt *resmgmt.Client, channelID, ccName, ccVersion string, peers []fab.Peer) bool {
	installedOnAllPeers := true
	for _, peer := range peers {
		chaincodeQueryResponse, err := resMgmt.QueryInstantiatedChaincodes(channelID, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithTargets(peer))		
		if err != nil {
			fmt.Println("QueryInstantiatedChaincodes chaincodeQueryResponse for peer fail")
		}else{
			fmt.Println("QueryInstantiatedChaincodes chaincodeQueryResponse for peer successful")
		}
		found := false
		for _, chaincode := range chaincodeQueryResponse.Chaincodes {
			if chaincode.Name == ccName && chaincode.Version == ccVersion {
				found = true
				break
			}
		}
		if !found {
			installedOnAllPeers = false
		}
	}
	return installedOnAllPeers
}

func (setup *FabricSetup) InstallAndInstantiateCC() error {

	// Create the chaincode package that will be sent to the peers
	ccPkg, err := packager.NewCCPackage(setup.ChaincodePath, setup.ChaincodeGoPath)
	if err != nil {
		return errors.WithMessage(err, "failed to create chaincode package")
	}else{
		fmt.Println("ccPkg created")
	}
	
	org_nowPeers, err := DiscoverLocalPeers(setup.org_nowAdminClientContext, 1)
	if err != nil {
		return errors.WithMessage(err, "failed to org_nowPeers")
	}else{
		fmt.Println("org_nowPeers finally")
	}
	for _, peer := range org_nowPeers {
		fmt.Println("org_nowPeers:",peer)
	}
	

	// Install example cc to org peers
	installCCReq := resmgmt.InstallCCRequest{Name: setup.ChainCodeID, Path: setup.ChaincodePath, Version: "0", Package: ccPkg}
	_, err = setup.org_nowResMgmt.InstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {			
		return errors.WithMessage(err, "org_nowResMgmt failed to install chaincode")
	}else{		
		fmt.Println("Chaincode org_nowResMgmt install success")
	}
	installed := queryInstalledCC("Org4", setup.org_nowResMgmt,setup.ChainCodeID, "0", org_nowPeers)
	if installed {
		fmt.Println("org_nowResMgmt ccHasInstall:ture")
	}else{
		fmt.Println("org_nowResMgmt ccHasInstall:false")
	}
	//ccPolicy := cauthdsl.SignedByAnyMember([]string{"org1.hf.srezd.io","org2.hf.srezd.io","org3.hf.srezd.io"})
	ccPolicy, err := cauthdsl.FromString("AND ('org1.hf.srezd.io.member','org2.hf.srezd.io.member','org3.hf.srezd.io.member','org4.hf.srezd.io.member')")
	if err != nil {
		//return resmgmt.InstantiateCCResponse{}, errors.Wrapf(err, "error creating CC policy [%s]", ccPolicyStr)
		return errors.WithMessage(err, "org_nowResMgmt failed to InstantiateCCResponse ccPolicy")
	}else{
		fmt.Println("ccPolicy InstantiateCCResponse successful")
	}
	request := resmgmt.InstantiateCCRequest{Name: setup.ChainCodeID, Path: setup.ChaincodeGoPath,Version: "0",Args: [][]byte{[]byte("init"), []byte("init")}, Policy: ccPolicy}
	resp, err := setup.org_nowResMgmt.InstantiateCC(setup.ChannelID, request)
	if err != nil || resp.TransactionID == "" {
		//return errors.WithMessage(err, "failed to instantiate the chaincode")
	} else {
		fmt.Println("Chaincode instantiate success")
	}
	found := queryInstantiatedCC("Org4", setup.org_nowResMgmt, setup.ChannelID, setup.ChainCodeID,"0", org_nowPeers)
	if found {
		fmt.Println("org_nowResMgmt app-1g4p0 found:ture")
	}else{
		fmt.Println("org_nowResMgmt app-1g4p0 found:false")
	}


	org_nowChannelClientContext := setup.sdk.ChannelContext(setup.ChannelID, fabsdk.WithUser(setup.Org_nowUser), fabsdk.WithOrg(setup.Org_now))
	setup.client, err = channel.New(org_nowChannelClientContext)
	if err != nil {
		return errors.WithMessage(err, "Failed to create new channel client for Org_now user")
	}else{
		fmt.Println("Successful to create new channel client for Org_now user")
	}
	setup.event, err = event.New(org_nowChannelClientContext)
	if err != nil {
		return errors.WithMessage(err, "failed to create new event client for Org_now user")
	}else{
		fmt.Println("Successful to create new event client for Org_now user")
	}

	fmt.Println("Chaincode app-1g4p0 Installation & Instantiation Successful")
	return nil
}

func (setup *FabricSetup) CloseSDK() {
	setup.sdk.Close()
}
