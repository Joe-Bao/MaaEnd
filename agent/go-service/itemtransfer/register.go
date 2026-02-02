package itemtransfer

import "github.com/MaaXYZ/maa-framework-go/v3"

func Register() {

	maa.AgentServerRegisterCustomRecognition("LocateItemInRepository", &RepoLocate{})
	maa.AgentServerRegisterCustomRecognition("LocateItemInBackpack", &BackpackLocate{})
	maa.AgentServerRegisterCustomAction("LeftClickWithCtrlDown", &LeftClickWithCtrlDown{})
	// maa.AgentServerRegisterCustomRecognition("LocateItemFromBackpack")
	// maa.AgentServerRegisterCustomAction("TransferItemToRepository")
}
