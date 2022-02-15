package ibctesting

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// KeeperTestSuite is a testing suite to test keeper functions.
type KeeperTestSuite struct {
	suite.Suite

	coordinator *Coordinator

	// testing chains used for convenience and readability
	chainA *TestChain
	chainB *TestChain
}

// TestKeeperTestSuite runs all the tests within this package.
func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// SetupTest creates a coordinator with 2 test chains.
func (suite *KeeperTestSuite) SetupTest() {
	// suite.coordinator = NewCoordinator(suite.T(), 2)         // initializes 2 test chains
	// suite.chainA = suite.coordinator.GetChain(GetChainID(0)) // convenience and readability
	// suite.chainB = suite.coordinator.GetChain(GetChainID(1)) // convenience and readability
}

func NewTransferPath(chainA, chainB *TestChain) *Path {
	path := NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = TransferPort
	path.EndpointB.ChannelConfig.PortID = TransferPort

	return path
}

func (suite *KeeperTestSuite) TestDelegate(t *testing.T) {
	suite.coordinator = NewCoordinator(suite.T(), 2) // initializes 2 test chains
	// suite.chainA = suite.coordinator.GetChain(GetChainID(0)) // convenience and readability
	// suite.chainB = suite.coordinator.GetChain(GetChainID(1)) // c
	// path := NewPath(suite.chainA, suite.chainB)
	// // clientID, connectionID, channelID empty
	// suite.coordinator.Setup(path) // clientID, connectionID, channelID filled
	// suite.Require().Equal("07-tendermint-0", path.EndpointA.ClientID)
	// suite.Require().Equal("connection-0", path.EndpointA.ClientID)
	// suite.Require().Equal("channel-0", path.EndpointA.ClientID)

	// 	// create packet 1
	// 	packet1 := NewPacket() // NewPacket would construct your packet

	// 	// send on endpointA
	// 	path.EndpointA.SendPacket(packet1)

	// 	// receive on endpointB
	// 	path.EndpointB.RecvPacket(packet1)

	// 	// acknowledge the receipt of the packet
	// 	path.EndpointA.AcknowledgePacket(packet1, ack)

	// 	// we can also relay
	// 	packet2 := NewPacket()

	// 	path.EndpointA.SendPacket(packet2)

	// 	path.Relay(packet2, expectedAck)

	// 	// if needed we can update our clients
	// 	path.EndpointB.UpdateClient()
	// }
}
