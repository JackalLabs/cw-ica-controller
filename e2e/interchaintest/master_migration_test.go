package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	logger "github.com/JackalLabs/storage-outpost/e2e/interchaintest/logger"
	"github.com/JackalLabs/storage-outpost/e2e/interchaintest/testsuite"
	"github.com/JackalLabs/storage-outpost/e2e/interchaintest/types"
	outpostfactory "github.com/JackalLabs/storage-outpost/e2e/interchaintest/types/outpostfactory"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
)

// SetupMigrationTestSuite starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
// sets up the contract and does the channel handshake for the contract test suite.
func (s *FactoryTestSuite) SetupMigrationTestSuite(ctx context.Context, encoding string) {
	// This starts the chains, relayer, creates the user accounts, and creates the ibc clients and connections.
	s.SetupSuite(ctx, chainSpecs)

	logger.InitLogger()

	// Upload the outpost's wasm module on Wasmd
	outpostCodeID, err := s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/storage_outpost.wasm")
	s.Require().NoError(err)

	// codeId is string and needs to be converted to uint64
	s.OutpostContractCodeId, err = strconv.ParseInt(outpostCodeID, 10, 64)
	s.Require().NoError(err)

	factoryCodeId, err := s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/outpost_factory.wasm")
	s.Require().NoError(err)

	instantiateMsg := outpostfactory.InstantiateMsg{StorageOutpostCodeId: int(s.OutpostContractCodeId)}
	// this is the outpost factory
	outpostfactoryContractAddr, err := s.ChainA.InstantiateContract(ctx, s.UserA.KeyName(), factoryCodeId, toString(instantiateMsg), false, "--gas", "500000", "--admin", s.UserA.KeyName())
	s.Require().NoError(err)
	s.FactoryAddress = outpostfactoryContractAddr

	// Jackal Labs account will be the admin of the outpost factory
	factoryContractInfoRes, infoErr := testsuite.GetContractInfo(ctx, s.ChainA, outpostfactoryContractAddr)
	s.Require().NoError(infoErr)
	s.Require().Equal(factoryContractInfoRes.Admin, s.UserA.FormattedAddress())
	logger.LogInfo(fmt.Sprintf("contract Info is: %s", factoryContractInfoRes))

	// TODO: wrapping the encoding with 'TxEncoding' is not needed anymore because 'Proto3Json'
	// is not the recommended encoding type for the ICA channel
	// we should just use an optional string
	proto3Encoding := outpostfactory.TxEncoding(encoding)

	// Create UserA's outpost
	createOutpostMsg := outpostfactory.ExecuteMsg{
		CreateOutpost: &outpostfactory.ExecuteMsg_CreateOutpost{
			Salt: nil,
			ChannelOpenInitOptions: outpostfactory.ChannelOpenInitOptions{
				ConnectionId:             s.ChainAConnID,
				CounterpartyConnectionId: s.ChainBConnID,
				TxEncoding:               &proto3Encoding,
			},
		},
	}

	res, err := s.ChainA.ExecuteContract(ctx, s.UserA.KeyName(), outpostfactoryContractAddr, toString(createOutpostMsg), "--gas", "500000")
	s.Require().NoError(err)
	// Confirm that UserA's outpost is administered by the factory
	outpostAddressFromEvent := logger.ParseOutpostAddressFromEvent(res.Events)
	outpostContractInfoRes, outpostInfoErr := testsuite.GetContractInfo(ctx, s.ChainA, outpostAddressFromEvent)
	s.Require().NoError(outpostInfoErr)
	s.Require().Equal(outpostContractInfoRes.Admin, outpostfactoryContractAddr)
	logger.LogInfo(fmt.Sprintf("outpostContractInfo is: %s", outpostContractInfoRes))

	// Save user A's outpost in the Factory suite for later use
	s.Contract = types.NewIcaContract(types.NewContract(outpostAddressFromEvent, outpostCodeID, s.ChainA))

	// Confirm UserA is the owner of the outpost they just made
	ownerQueryRes, ownerError := testsuite.GetOutpostOwner(ctx, s.ChainA, outpostAddressFromEvent)
	s.Require().NoError(ownerError)
	var outpostOwner string
	if err := json.Unmarshal(ownerQueryRes.Data, &outpostOwner); err != nil {
		log.Fatalf("Error parsing response data: %v", err)
	}
	s.Require().Equal(s.UserA.FormattedAddress(), outpostOwner)

	// We know that the outpost we just made emitted an event showing its address
	// We can now query the mapping inside of 'outpost factory' to confirm that we mapped the correct address
	// Query for the relevant addresses to ensure everything exists
	OutpostAddressFromMap, addressErr := testsuite.GetOutpostAddressFromFactoryMap(ctx, s.ChainA, outpostfactoryContractAddr, s.UserA.FormattedAddress())
	s.Require().NoError(addressErr)
	var mappedOutpostAddress string
	if err := json.Unmarshal(OutpostAddressFromMap.Data, &mappedOutpostAddress); err != nil {
		log.Fatalf("Error parsing response data: %v", err)
	}
	s.Require().Equal(outpostAddressFromEvent, mappedOutpostAddress)

	// TODO: Confirm that outpost still works to post a key

}

func (s *FactoryTestSuite) TestMasterMigration() {
	ctx := context.Background()

	// This starts the chains, relayer, creates the user accounts, creates the ibc clients and connections,
	// sets up the contract and does the channel handshake for the contract test suite.
	s.SetupMigrationTestSuite(ctx, icatypes.EncodingProtobuf) // NOTE: canined's ibc-go is outdated and does not support proto3json

	// Store v2 of the outpost
	newOutpostCodeId, err := s.ChainA.StoreContract(ctx, s.UserA.KeyName(), "../../artifacts/v2/storage_outpost_v2.wasm")
	s.Require().NoError(err)
	fmt.Println(newOutpostCodeId)
	logger.LogInfo(fmt.Sprintf("new outpost code id is: %s", newOutpostCodeId))

	migrateOutpostMsg := outpostfactory.ExecuteMsg{
		MigrateOutpost: &outpostfactory.ExecuteMsg_MigrateOutpost{
			OutpostOwner:     s.UserA.FormattedAddress(),
			NewOutpostCodeId: newOutpostCodeId,
		},
	}

	res, err := s.ChainA.ExecuteContract(ctx, s.UserA.KeyName(), s.FactoryAddress, toString(migrateOutpostMsg), "--gas", "500000")
	s.Require().NoError(err)
	fmt.Println(res)

	// Query contractinfo of userA's outpost to see that it points to code ID 3

	outpostInfoRes, err := testsuite.GetContractInfo(ctx, s.ChainA, s.Contract.Address)
	logger.LogInfo(fmt.Sprintf("codeID is: %d", outpostInfoRes.ContractInfo.CodeID))

	newOutpostCodeIdUint, err := strconv.ParseUint(newOutpostCodeId, 10, 64)
	s.Require().NoError(err)
	s.Require().Equal(newOutpostCodeIdUint, outpostInfoRes.ContractInfo.CodeID)

	fmt.Println("END OF TEST")

	time.Sleep(time.Duration(10) * time.Hour)

}