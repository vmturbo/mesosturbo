package vmturbocommunicator

import (
	"fmt"

	"github.com/golang/glog"
	vmtmeta "github.com/vmturbo/mesosturbo/communicator/metadata"
	vmtapi "github.com/vmturbo/mesosturbo/communicator/vmtapi"
	comm "github.com/vmturbo/vmturbo-go-sdk/communicator"
	"github.com/vmturbo/vmturbo-go-sdk/sdk"
)

type VMTCommunicator struct {
	// TODO mesos client ?
	tmpClient map[string]string
	meta      *vmtmeta.VMTMeta
	wsComm    *comm.WebSocketCommunicator
	// TODO etcdStorage storage.Storage
}

func NewVMTCommunicator(client map[string]string, vmtMetadata *vmtmeta.VMTMeta) *VMTCommunicator {
	return &VMTCommunicator{
		tmpClient: client,
		meta:      vmtMetadata,
	}
}

func (vmtcomm *VMTCommunicator) Run() {
	vmtcomm.Init()
	vmtcomm.RegisterMesos()
}

// Init() intialize the VMTCommunicator, creating websocket communicator and server message handler.
func (vmtcomm *VMTCommunicator) Init() {
	wsCommunicator := &comm.WebSocketCommunicator{
		VmtServerAddress: vmtcomm.meta.ServerAddress,
		LocalAddress:     vmtcomm.meta.LocalAddress,
		ServerUsername:   vmtcomm.meta.WebSocketUsername,
		ServerPassword:   vmtcomm.meta.WebSocketPassword,
	}
	vmtcomm.wsComm = wsCommunicator

	// First create the message handler for mesos
	mesosMsgHandler := &MesosServerMessageHandler{
		//	mesosClient: vmtcomm.tmpClient,
		meta:   vmtcomm.meta,
		wsComm: wsCommunicator,
		//	etcdStorage: vmtcomm.etcdStorage,
	}
	wsCommunicator.ServerMsgHandler = mesosMsgHandler
	return
}

// Register Mesos target onto server and start listen to websocket.
func (vmtcomm *VMTCommunicator) RegisterMesos() {
	// 1. Construct the account definition for Mesos.
	acctDefProps := createAccountDefMesos()

	// 2. Build supply chain.
	templateDtos := createSupplyChain()
	glog.V(3).Infof("Supply chain for Mesos is created.")

	// 3. construct the mesos ,  mesosProbe is the only probe supported.
	probeType := "Mesos" //vmtcomm.meta.TargetType
	probeCat := "Container"
	mesosProbe := comm.NewProbeInfoBuilder(probeType, probeCat, templateDtos, acctDefProps).Create()

	// 4. Add mesosProbe to probe, and that's the only probe supported in this client.
	var probes []*comm.ProbeInfo
	probes = append(probes, mesosProbe)

	// 5. Create mediation container
	containerInfo := &comm.ContainerInfo{
		Probes: probes,
	}
	fmt.Printf("Send registration message: %+v", containerInfo)
	vmtcomm.wsComm.RegisterAndListen(containerInfo)
}

// TODO, rephrase comment.
// create account definition for mesos, which is used later to create mesos probe.
// The return type is a list of ProbeInfo_AccountDefProp.
// For a valid definition, targetNameIdentifier, username and password should be contained.
func createAccountDefMesos() []*comm.AccountDefEntry {
	var acctDefProps []*comm.AccountDefEntry

	// target id
	targetIDAcctDefEntry := comm.NewAccountDefEntryBuilder("targetIdentifier", "Address",
		"IP of the mesos master", ".*", comm.AccountDefEntry_OPTIONAL, false).Create()
	acctDefProps = append(acctDefProps, targetIDAcctDefEntry)

	// username
	usernameAcctDefEntry := comm.NewAccountDefEntryBuilder("username", "Username",
		"Username of the mesos master", ".*", comm.AccountDefEntry_OPTIONAL, false).Create()
	acctDefProps = append(acctDefProps, usernameAcctDefEntry)

	// password
	passwdAcctDefEntry := comm.NewAccountDefEntryBuilder("password", "Password",
		"Password of the mesos master", ".*", comm.AccountDefEntry_OPTIONAL, true).Create()
	acctDefProps = append(acctDefProps, passwdAcctDefEntry)

	return acctDefProps
}

// also include mesos supply chain explanation
// check if task becomes vApp or App
func createSupplyChain() []*sdk.TemplateDTO {
	glog.V(3).Infof(".......... Now use builder to create a supply chain ..........")

	fakeKey := "fake"

	slaveSupplyChainNodeBuilder := sdk.NewSupplyChainNodeBuilder()
	slaveSupplyChainNodeBuilder = slaveSupplyChainNodeBuilder.
		Entity(sdk.EntityDTO_VIRTUAL_MACHINE).
		Selling(sdk.CommodityDTO_CPU_ALLOCATION, fakeKey).
		Selling(sdk.CommodityDTO_MEM_ALLOCATION, fakeKey).
		Selling(sdk.CommodityDTO_VCPU, fakeKey).
		Selling(sdk.CommodityDTO_VMEM, fakeKey).
		Selling(sdk.CommodityDTO_APPLICATION, fakeKey)

	glog.V(3).Infof(".......... slave supply chain node builder is created ..........")

	// Container Supplychain builder
	containerSupplyChainNodeBuilder := sdk.NewSupplyChainNodeBuilder()
	containerSupplyChainNodeBuilder = containerSupplyChainNodeBuilder.
		Entity(sdk.EntityDTO_CONTAINER).
		Selling(sdk.CommodityDTO_CPU_ALLOCATION, fakeKey).
		Selling(sdk.CommodityDTO_MEM_ALLOCATION, fakeKey)

	cpuAllocationType := sdk.CommodityDTO_CPU_ALLOCATION
	cpuAllocationTemplateComm := &sdk.TemplateCommodity{
		Key:           &fakeKey,
		CommodityType: &cpuAllocationType,
	}
	memAllocationType := sdk.CommodityDTO_MEM_ALLOCATION
	memAllocationTemplateComm := &sdk.TemplateCommodity{
		Key:           &fakeKey,
		CommodityType: &memAllocationType,
	}
	//vmpmaccessType := sdk.CommodityDTO_VMPM_ACCESS

	containerSupplyChainNodeBuilder = containerSupplyChainNodeBuilder.
		Provider(sdk.EntityDTO_VIRTUAL_MACHINE, sdk.Provider_LAYERED_OVER).
		Buys(*cpuAllocationTemplateComm).
		Buys(*memAllocationTemplateComm)
	glog.V(3).Infof(".......... container supply chain node builder is created ..........")

	// Application supplychain builder
	appSupplyChainNodeBuilder := sdk.NewSupplyChainNodeBuilder()
	appSupplyChainNodeBuilder = appSupplyChainNodeBuilder.
		Entity(sdk.EntityDTO_APPLICATION).
		Selling(sdk.CommodityDTO_TRANSACTION, fakeKey)
	// Buys CpuAllocation/MemAllocation from Pod
	appCpuAllocationTemplateComm := &sdk.TemplateCommodity{
		Key:           &fakeKey,
		CommodityType: &cpuAllocationType,
	}
	appMemAllocationTemplateComm := &sdk.TemplateCommodity{
		Key:           &fakeKey,
		CommodityType: &memAllocationType,
	}
	appSupplyChainNodeBuilder = appSupplyChainNodeBuilder.Provider(sdk.EntityDTO_CONTAINER, sdk.Provider_LAYERED_OVER).Buys(*appCpuAllocationTemplateComm).Buys(*appMemAllocationTemplateComm)
	// Buys VCpu and VMem from VM
	vCpuType := sdk.CommodityDTO_VCPU
	appVCpu := &sdk.TemplateCommodity{
		Key:           &fakeKey,
		CommodityType: &vCpuType,
	}
	vMemType := sdk.CommodityDTO_VMEM
	appVMem := &sdk.TemplateCommodity{
		Key:           &fakeKey,
		CommodityType: &vMemType,
	}
	appCommType := sdk.CommodityDTO_APPLICATION
	appAppComm := &sdk.TemplateCommodity{
		Key:           &fakeKey,
		CommodityType: &appCommType,
	}
	appSupplyChainNodeBuilder = appSupplyChainNodeBuilder.Provider(sdk.EntityDTO_VIRTUAL_MACHINE, sdk.Provider_HOSTING).Buys(*appVCpu).Buys(*appVMem).Buys(*appAppComm)

	// Link from Pod to VM
	vmContainerExtLinkBuilder := sdk.NewExternalEntityLinkBuilder()
	vmContainerExtLinkBuilder.Link(sdk.EntityDTO_CONTAINER, sdk.EntityDTO_VIRTUAL_MACHINE, sdk.Provider_LAYERED_OVER).
		Commodity(cpuAllocationType, true).
		Commodity(memAllocationType, true).
		ProbeEntityPropertyDef(sdk.SUPPLYCHAIN_CONSTANT_IP_ADDRESS, "IP Address where the Container is running").
		ExternalEntityPropertyDef(sdk.VM_IP)

	vmContainerExternalLink := vmContainerExtLinkBuilder.Build()

	// Link from Application to VM
	vmAppExtLinkBuilder := sdk.NewExternalEntityLinkBuilder()
	vmAppExtLinkBuilder.Link(sdk.EntityDTO_APPLICATION, sdk.EntityDTO_VIRTUAL_MACHINE, sdk.Provider_HOSTING).
		Commodity(vCpuType, true).
		Commodity(vMemType, true).
		Commodity(appCommType, true).
		ProbeEntityPropertyDef(sdk.SUPPLYCHAIN_CONSTANT_IP_ADDRESS, "IP Address where the Application is running").
		ExternalEntityPropertyDef(sdk.VM_IP)

	vmAppExternalLink := vmAppExtLinkBuilder.Build()

	supplyChainBuilder := sdk.NewSupplyChainBuilder()
	//	supplyChainBuilder.Top(vAppSupplyChainNodeBuilder)
	supplyChainBuilder.Top(appSupplyChainNodeBuilder)
	supplyChainBuilder.ConnectsTo(vmAppExternalLink)
	supplyChainBuilder.Entity(containerSupplyChainNodeBuilder)
	supplyChainBuilder.ConnectsTo(vmContainerExternalLink)
	supplyChainBuilder.Entity(slaveSupplyChainNodeBuilder)

	return supplyChainBuilder.Create()
}

// Send action response to vmt server.
func (vmtcomm *VMTCommunicator) SendActionReponse(state sdk.ActionResponseState, progress, messageID int32, description string) {
	// 1. build response
	response := &comm.ActionResponse{
		ActionResponseState: &state,
		Progress:            &progress,
		ResponseDescription: &description,
	}

	// 2. built action result.
	result := &comm.ActionResult{
		Response: response,
	}

	// 3. Build Client message
	clientMsg := comm.NewClientMessageBuilder(messageID).SetActionResponse(result).Create()

	vmtcomm.wsComm.SendClientMessage(clientMsg)
}

// Code is very similar to those in serverMsgHandler.
func (vmtcomm *VMTCommunicator) DiscoverTarget() {
	vmtUrl := vmtcomm.wsComm.VmtServerAddress

	extCongfix := make(map[string]string)
	extCongfix["Username"] = vmtcomm.meta.OpsManagerUsername
	extCongfix["Password"] = vmtcomm.meta.OpsManagerPassword
	vmturboApi := vmtapi.NewVmtApi(vmtUrl, extCongfix)

	// Discover mesos target.
	vmturboApi.DiscoverTarget(vmtcomm.meta.NameOrAddress)
}
