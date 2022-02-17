package ibctesting

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/modules/core/04-channel/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	cryptoenc "github.com/tendermint/tendermint/crypto/encoding"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
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
	suite.coordinator = NewCoordinator(suite.T(), 2)         // initializes 2 test chains
	suite.chainA = suite.coordinator.GetChain(GetChainID(0)) // convenience and readability
	suite.chainB = suite.coordinator.GetChain(GetChainID(1)) // convenience and readability
}

func (suite *KeeperTestSuite) TestMain() {
	path := NewPath(suite.chainA, suite.chainB) // clientID, connectionID, channelID empty
	suite.coordinator.Setup(path)               // clientID, connectionID, channelID filled
	suite.Require().Equal("07-tendermint-0", path.EndpointA.ClientID)
	suite.Require().Equal("connection-0", path.EndpointA.ConnectionID)
	suite.Require().Equal("channel-0", path.EndpointA.ChannelID)

	// create packet
	packet := channeltypes.NewPacket([]byte{byte(1)}, 1, path.EndpointA.ChannelConfig.PortID, path.EndpointB.ChannelID,
		path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, clienttypes.NewHeight(1, 0), 0)

	// send on endpointA
	err := path.EndpointA.SendPacket(packet)
	suite.Require().NoError(err)

	err = path.EndpointB.UpdateClient()
	suite.Require().NoError(err)

	// receive on endpointB
	err = path.EndpointB.RecvPacket(packet)
	suite.Require().NoError(err)

	parentStakingKeeper := suite.chainA.App.GetStakingKeeper()

	bondAmt := sdk.NewInt(1000000)

	delAddr := suite.chainA.SenderAccount.GetAddress()

	// Choose a validator, and get its address and data structure into the correct types
	tmValidator := suite.chainA.Vals.Validators[0]
	valAddr, err := sdk.ValAddressFromHex(tmValidator.Address.String())
	suite.Require().NoError(err)
	validator, found := parentStakingKeeper.GetValidator(suite.chainA.GetContext(), valAddr)
	suite.Require().True(found)

	// Print next validator hash
	fmt.Printf("Old Hash %x %x\n", suite.chainA.CurrentHeader.ValidatorsHash, suite.chainA.CurrentHeader.GetNextValidatorsHash())
	// fmt.Printf("Delegates to %v\n", tmValidator.PubKey)
	_, err = parentStakingKeeper.Delegate(suite.chainA.GetContext(), delAddr, bondAmt, stakingtypes.Unbonded, stakingtypes.Validator(validator), true)
	suite.Require().NoError(err)

	// retrieve validator updates
	valUpdates := suite.chainA.App.EndBlock(abci.RequestEndBlock{})
	// update validator set
	p, err := cryptoenc.PubKeyFromProto(valUpdates.ValidatorUpdates[0].PubKey)
	suite.Require().NoError(err)

	val := tmtypes.NewValidator(p, valUpdates.ValidatorUpdates[0].Power)
	valset := tmtypes.NewValidatorSet([]*tmtypes.Validator{val})

	// commit and next block
	suite.chainA.App.Commit()

	// BEGIN NEXT BLOCK

	oldValset := suite.chainA.Vals
	// set the last header to the current header
	// use nil trusted fields
	suite.chainA.LastHeader = suite.chainA.CurrentTMClientHeader()
	// suite.chainA.LastHeader.Header.NextValidatorsHash = valset.Hash()

	// NOTE: We need to get validators from counterparty at height: trustedHeight+1
	// since the last trusted validators for a header at height h
	// is the NextValidators at h+1 committed to in header h by
	// NextValidatorsHash

	// increment the current header
	suite.chainA.CurrentHeader = tmproto.Header{
		ChainID: suite.chainA.ChainID,
		Height:  suite.chainA.App.LastBlockHeight() + 1,
		AppHash: suite.chainA.App.LastCommitID().Hash,
		// NOTE: the time is increased by the coordinator to maintain time synchrony amongst
		// chains.
		Time:               suite.chainA.CurrentHeader.Time,
		ValidatorsHash:     oldValset.Hash(),
		NextValidatorsHash: valset.Hash(),
	}

	// update chain validator set manually
	suite.chainA.Vals = valset

	suite.chainA.App.BeginBlock(abci.RequestBeginBlock{Header: suite.chainA.CurrentHeader})
	fmt.Println(suite.chainA.CurrentHeader.Height)

	// END NEXT BLOCK
	fmt.Printf("New Last Header Hashes %x %x\n", suite.chainA.LastHeader.Header.ValidatorsHash, suite.chainA.LastHeader.Header.NextValidatorsHash)

	fmt.Printf("New Header Hashes %x %x\n", suite.chainA.CurrentHeader.ValidatorsHash, suite.chainA.CurrentHeader.GetNextValidatorsHash())

	// Update the second client chain;
	// commit the the first chain block
	err = path.EndpointB.UpdateClient()
	suite.Require().NoError(err)

	err = path.EndpointA.UpdateClient()
	suite.Require().NoError(err)

	err = path.EndpointB.UpdateClient()
	suite.Require().NoError(err)

}
