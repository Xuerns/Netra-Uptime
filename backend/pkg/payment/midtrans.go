package payment

import (
	"os"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
)

var SnapClient snap.Client
var CoreApiClient coreapi.Client

func InitMidtrans() {
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")

	// Still on development
	env := midtrans.Sandbox

	SnapClient.New(serverKey, env)
	CoreApiClient.New(serverKey, env)
}