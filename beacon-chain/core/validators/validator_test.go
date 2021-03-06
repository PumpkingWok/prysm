package validators

import (
	"bytes"
	"math/big"
	"testing"

	pb "github.com/prysmaticlabs/prysm/proto/beacon/p2p/v1"
	"github.com/prysmaticlabs/prysm/shared/bitutil"
	"github.com/prysmaticlabs/prysm/shared/params"
)

func TestHasVoted(t *testing.T) {
	// Setting bit field to 11111111.
	pendingAttestation := &pb.AggregatedAttestation{
		AttesterBitfield: []byte{255},
	}

	for i := 0; i < len(pendingAttestation.AttesterBitfield); i++ {
		voted, err := bitutil.CheckBit(pendingAttestation.AttesterBitfield, i)
		if err != nil {
			t.Errorf("checking bit failed at index: %d with : %v", i, err)
		}

		if !voted {
			t.Error("validator voted but received didn't vote")
		}
	}

	// Setting bit field to 01010101.
	pendingAttestation = &pb.AggregatedAttestation{
		AttesterBitfield: []byte{85},
	}

	for i := 0; i < len(pendingAttestation.AttesterBitfield); i++ {
		voted, err := bitutil.CheckBit(pendingAttestation.AttesterBitfield, i)
		if err != nil {
			t.Errorf("checking bit failed at index: %d : %v", i, err)
		}

		if i%2 == 0 && voted {
			t.Error("validator didn't vote but received voted")
		}
		if i%2 == 1 && !voted {
			t.Error("validator voted but received didn't vote")
		}
	}
}

func TestInitialValidatorRegistry(t *testing.T) {
	validators := InitialValidatorRegistry()
	for _, validator := range validators {
		if validator.GetBalance() != params.BeaconConfig().DepositSize*params.BeaconConfig().Gwei {
			t.Fatalf("deposit size of validator is not expected %d", validator.GetBalance())
		}
		if validator.GetStatus() != uint64(params.Active) {
			t.Errorf("validator status is not active: %d", validator.GetStatus())
		}
	}
}

func TestAreAttesterBitfieldsValid(t *testing.T) {
	attestation := &pb.AggregatedAttestation{
		AttesterBitfield: []byte{'F'},
	}

	indices := []uint32{0, 1, 2, 3, 4, 5, 6, 7}

	isValid := AreAttesterBitfieldsValid(attestation, indices)
	if !isValid {
		t.Fatalf("expected validation to pass for bitfield %v and indices %v", attestation, indices)
	}
}

func TestAreAttesterBitfieldsValidFalse(t *testing.T) {
	attestation := &pb.AggregatedAttestation{
		AttesterBitfield: []byte{'F', 'F'},
	}

	indices := []uint32{0, 1, 2, 3, 4, 5, 6, 7}

	isValid := AreAttesterBitfieldsValid(attestation, indices)
	if isValid {
		t.Fatalf("expected validation to fail for bitfield %v and indices %v", attestation, indices)
	}
}

func TestAreAttesterBitfieldsValidZerofill(t *testing.T) {
	attestation := &pb.AggregatedAttestation{
		AttesterBitfield: []byte{'F'},
	}

	indices := []uint32{0, 1, 2, 3, 4, 5, 6}

	isValid := AreAttesterBitfieldsValid(attestation, indices)
	if !isValid {
		t.Fatalf("expected validation to pass for bitfield %v and indices %v", attestation, indices)
	}
}

func TestAreAttesterBitfieldsValidNoZerofill(t *testing.T) {
	attestation := &pb.AggregatedAttestation{
		AttesterBitfield: []byte{'E'},
	}

	var indices []uint32
	for i := uint32(0); i < uint32(params.BeaconConfig().TargetCommitteeSize)+1; i++ {
		indices = append(indices, i)
	}

	isValid := AreAttesterBitfieldsValid(attestation, indices)
	if isValid {
		t.Fatalf("expected validation to fail for bitfield %v and indices %v", attestation, indices)
	}
}

func TestProposerShardAndIndex(t *testing.T) {
	shardCommittees := []*pb.ShardAndCommitteeArray{
		{ArrayShardAndCommittee: []*pb.ShardAndCommittee{
			{Shard: 0, Committee: []uint32{0, 1, 2, 3, 4}},
			{Shard: 1, Committee: []uint32{5, 6, 7, 8, 9}},
		}},
		{ArrayShardAndCommittee: []*pb.ShardAndCommittee{
			{Shard: 2, Committee: []uint32{10, 11, 12, 13, 14}},
			{Shard: 3, Committee: []uint32{15, 16, 17, 18, 19}},
		}},
		{ArrayShardAndCommittee: []*pb.ShardAndCommittee{
			{Shard: 4, Committee: []uint32{20, 21, 22, 23, 24}},
			{Shard: 5, Committee: []uint32{25, 26, 27, 28, 29}},
		}},
	}
	if _, _, err := ProposerShardAndIndex(shardCommittees, 100, 0); err == nil {
		t.Error("ProposerShardAndIndex should have failed with invalid lcs")
	}
	shard, index, err := ProposerShardAndIndex(shardCommittees, 128, 65)
	if err != nil {
		t.Fatalf("ProposerShardAndIndex failed with %v", err)
	}
	if shard != 2 {
		t.Errorf("Invalid shard ID. Wanted 2, got %d", shard)
	}
	if index != 0 {
		t.Errorf("Invalid proposer index. Wanted 0, got %d", index)
	}
}

func TestValidatorIndex(t *testing.T) {
	var validators []*pb.ValidatorRecord
	for i := 0; i < 10; i++ {
		validators = append(validators, &pb.ValidatorRecord{Pubkey: []byte{}, Status: uint64(params.Active)})
	}
	if _, err := ValidatorIndex([]byte("100"), validators); err == nil {
		t.Fatalf("ValidatorIndex should have failed,  there's no validator with pubkey 100")
	}
	validators[5].Pubkey = []byte("100")
	index, err := ValidatorIndex([]byte("100"), validators)
	if err != nil {
		t.Fatalf("call ValidatorIndex failed: %v", err)
	}
	if index != 5 {
		t.Errorf("Incorrect validator index. Wanted 5, Got %v", index)
	}
}

func TestValidatorShard(t *testing.T) {
	var validators []*pb.ValidatorRecord
	for i := 0; i < 21; i++ {
		validators = append(validators, &pb.ValidatorRecord{Pubkey: []byte{}, Status: uint64(params.Active)})
	}
	shardCommittees := []*pb.ShardAndCommitteeArray{
		{ArrayShardAndCommittee: []*pb.ShardAndCommittee{
			{Shard: 0, Committee: []uint32{0, 1, 2, 3, 4, 5, 6}},
			{Shard: 1, Committee: []uint32{7, 8, 9, 10, 11, 12, 13}},
			{Shard: 2, Committee: []uint32{14, 15, 16, 17, 18, 19}},
		}},
	}
	validators[19].Pubkey = []byte("100")
	Shard, err := ValidatorShardID([]byte("100"), validators, shardCommittees)
	if err != nil {
		t.Fatalf("call ValidatorShard failed: %v", err)
	}
	if Shard != 2 {
		t.Errorf("Incorrect validator shard ID. Wanted 2, Got %v", Shard)
	}

	validators[19].Pubkey = []byte{}
	if _, err := ValidatorShardID([]byte("100"), validators, shardCommittees); err == nil {
		t.Fatalf("ValidatorShard should have failed, there's no validator with pubkey 100")
	}

	validators[20].Pubkey = []byte("100")
	if _, err := ValidatorShardID([]byte("100"), validators, shardCommittees); err == nil {
		t.Fatalf("ValidatorShard should have failed, validator indexed at 20 is not in the committee")
	}
}

func TestValidatorSlotAndResponsibility(t *testing.T) {
	var validators []*pb.ValidatorRecord
	for i := 0; i < 61; i++ {
		validators = append(validators, &pb.ValidatorRecord{Pubkey: []byte{}, Status: uint64(params.Active)})
	}
	shardCommittees := []*pb.ShardAndCommitteeArray{
		{ArrayShardAndCommittee: []*pb.ShardAndCommittee{
			{Shard: 0, Committee: []uint32{0, 1, 2, 3, 4, 5, 6}},
			{Shard: 1, Committee: []uint32{7, 8, 9, 10, 11, 12, 13}},
			{Shard: 2, Committee: []uint32{14, 15, 16, 17, 18, 19}},
		}},
		{ArrayShardAndCommittee: []*pb.ShardAndCommittee{
			{Shard: 3, Committee: []uint32{20, 21, 22, 23, 24, 25, 26}},
			{Shard: 4, Committee: []uint32{27, 28, 29, 30, 31, 32, 33}},
			{Shard: 5, Committee: []uint32{34, 35, 36, 37, 38, 39}},
		}},
		{ArrayShardAndCommittee: []*pb.ShardAndCommittee{
			{Shard: 6, Committee: []uint32{40, 41, 42, 43, 44, 45, 46}},
			{Shard: 7, Committee: []uint32{47, 48, 49, 50, 51, 52, 53}},
			{Shard: 8, Committee: []uint32{54, 55, 56, 57, 58, 59}},
		}},
	}
	if _, _, err := ValidatorSlotAndRole([]byte("100"), validators, shardCommittees); err == nil {
		t.Fatalf("ValidatorSlot should have failed, there's no validator with pubkey 100")
	}

	validators[59].Pubkey = []byte("100")
	slot, _, err := ValidatorSlotAndRole([]byte("100"), validators, shardCommittees)
	if err != nil {
		t.Fatalf("call ValidatorSlot failed: %v", err)
	}
	if slot != 2 {
		t.Errorf("Incorrect validator slot ID. Wanted 1, Got %v", slot)
	}

	validators[60].Pubkey = []byte("101")
	if _, _, err := ValidatorSlotAndRole([]byte("101"), validators, shardCommittees); err == nil {
		t.Fatalf("ValidatorSlot should have failed, validator indexed at 60 is not in the committee")
	}
}

func TestTotalActiveValidatorDeposit(t *testing.T) {
	var validators []*pb.ValidatorRecord
	for i := 0; i < 10; i++ {
		validators = append(validators, &pb.ValidatorRecord{Balance: 1e9, Status: uint64(params.Active)})
	}

	expectedTotalDeposit := new(big.Int)
	expectedTotalDeposit.SetString("10000000000", 10)

	totalDeposit := TotalActiveValidatorDeposit(validators)
	if expectedTotalDeposit.Cmp(new(big.Int).SetUint64(totalDeposit)) != 0 {
		t.Fatalf("incorrect total deposit calculated %d", totalDeposit)
	}

	totalDepositETH := TotalActiveValidatorDepositInEth(validators)
	if totalDepositETH != 10 {
		t.Fatalf("incorrect total deposit in ETH calculated %d", totalDepositETH)
	}
}

func TestVotedBalanceInAttestation(t *testing.T) {
	var validators []*pb.ValidatorRecord
	defaultBalance := uint64(1e9)
	for i := 0; i < 100; i++ {
		validators = append(validators, &pb.ValidatorRecord{Balance: defaultBalance, Status: uint64(params.Active)})
	}

	// Calculating balances with zero votes by attesters.
	attestation := &pb.AggregatedAttestation{
		AttesterBitfield: []byte{0},
	}

	indices := []uint32{4, 8, 10, 14, 30}
	expectedTotalBalance := uint64(len(indices)) * defaultBalance

	totalBalance, voteBalance, err := VotedBalanceInAttestation(validators, indices, attestation)

	if err != nil {
		t.Fatalf("unable to get voted balances in attestation %v", err)
	}

	if totalBalance != expectedTotalBalance {
		t.Errorf("incorrect total balance calculated %d", totalBalance)
	}

	if voteBalance != 0 {
		t.Errorf("incorrect vote balance calculated %d", voteBalance)
	}

	// Calculating balances with 3 votes by attesters.

	newAttestation := &pb.AggregatedAttestation{
		AttesterBitfield: []byte{224}, // 128 + 64 + 32
	}

	expectedTotalBalance = uint64(len(indices)) * defaultBalance

	totalBalance, voteBalance, err = VotedBalanceInAttestation(validators, indices, newAttestation)

	if err != nil {
		t.Fatalf("unable to get voted balances in attestation %v", err)
	}

	if totalBalance != expectedTotalBalance {
		t.Errorf("incorrect total balance calculated %d", totalBalance)
	}

	if voteBalance != defaultBalance*3 {
		t.Errorf("incorrect vote balance calculated %d", voteBalance)
	}

}

func TestAddValidatorRegistry(t *testing.T) {
	var existingValidatorRegistry []*pb.ValidatorRecord
	for i := 0; i < 10; i++ {
		existingValidatorRegistry = append(existingValidatorRegistry, &pb.ValidatorRecord{Status: uint64(params.Active)})
	}

	// Create a new validator.
	validators := AddPendingValidator(existingValidatorRegistry, []byte{'A'}, []byte{'C'}, uint64(params.PendingActivation))

	// The newly added validator should be indexed 10.
	if validators[10].Status != uint64(params.PendingActivation) {
		t.Errorf("Newly added validator should be pending")
	}
	if validators[10].Balance != uint64(params.BeaconConfig().DepositSize*params.BeaconConfig().Gwei) {
		t.Errorf("Incorrect deposit size")
	}

	// Set validator 6 to withdrawn
	existingValidatorRegistry[5].Status = uint64(params.Withdrawn)
	validators = AddPendingValidator(existingValidatorRegistry, []byte{'E'}, []byte{'F'}, uint64(params.PendingActivation))

	// The newly added validator should be indexed 5.
	if validators[5].Status != uint64(params.PendingActivation) {
		t.Errorf("Newly added validator should be pending")
	}
	if validators[5].Balance != uint64(params.BeaconConfig().DepositSize*params.BeaconConfig().Gwei) {
		t.Errorf("Incorrect deposit size")
	}
}

func TestChangeValidatorRegistry(t *testing.T) {
	existingValidatorRegistry := []*pb.ValidatorRecord{
		{Pubkey: []byte{1}, Status: uint64(params.PendingActivation), Balance: uint64(params.BeaconConfig().DepositSize * params.BeaconConfig().Gwei), LatestStatusChangeSlot: params.BeaconConfig().MinWithdrawalPeriod},
		{Pubkey: []byte{2}, Status: uint64(params.PendingExit), Balance: uint64(params.BeaconConfig().DepositSize * params.BeaconConfig().Gwei), LatestStatusChangeSlot: params.BeaconConfig().MinWithdrawalPeriod},
		{Pubkey: []byte{3}, Status: uint64(params.PendingActivation), Balance: uint64(params.BeaconConfig().DepositSize * params.BeaconConfig().Gwei), LatestStatusChangeSlot: params.BeaconConfig().MinWithdrawalPeriod},
		{Pubkey: []byte{4}, Status: uint64(params.PendingExit), Balance: uint64(params.BeaconConfig().DepositSize * params.BeaconConfig().Gwei), LatestStatusChangeSlot: params.BeaconConfig().MinWithdrawalPeriod},
		{Pubkey: []byte{5}, Status: uint64(params.PendingActivation), Balance: uint64(params.BeaconConfig().DepositSize * params.BeaconConfig().Gwei), LatestStatusChangeSlot: params.BeaconConfig().MinWithdrawalPeriod},
		{Pubkey: []byte{6}, Status: uint64(params.PendingExit), Balance: uint64(params.BeaconConfig().DepositSize * params.BeaconConfig().Gwei), LatestStatusChangeSlot: params.BeaconConfig().MinWithdrawalPeriod},
		{Pubkey: []byte{7}, Status: uint64(params.PendingWithdraw), Balance: uint64(params.BeaconConfig().DepositSize * params.BeaconConfig().Gwei)},
		{Pubkey: []byte{8}, Status: uint64(params.PendingWithdraw), Balance: uint64(params.BeaconConfig().DepositSize * params.BeaconConfig().Gwei)},
		{Pubkey: []byte{9}, Status: uint64(params.Penalized), Balance: uint64(params.BeaconConfig().DepositSize * params.BeaconConfig().Gwei)},
		{Pubkey: []byte{10}, Status: uint64(params.Penalized), Balance: uint64(params.BeaconConfig().DepositSize * params.BeaconConfig().Gwei)},
		{Pubkey: []byte{11}, Status: uint64(params.Active), Balance: uint64(params.BeaconConfig().DepositSize * params.BeaconConfig().Gwei)},
		{Pubkey: []byte{12}, Status: uint64(params.Active), Balance: uint64(params.BeaconConfig().DepositSize * params.BeaconConfig().Gwei)},
		{Pubkey: []byte{13}, Status: uint64(params.Active), Balance: uint64(params.BeaconConfig().DepositSize * params.BeaconConfig().Gwei)},
		{Pubkey: []byte{14}, Status: uint64(params.Active), Balance: uint64(params.BeaconConfig().DepositSize * params.BeaconConfig().Gwei)},
	}

	validators := ChangeValidatorRegistry(params.BeaconConfig().MinWithdrawalPeriod+1, 50*10e9, existingValidatorRegistry)

	if validators[0].Status != uint64(params.Active) {
		t.Errorf("Wanted status Active. Got: %d", validators[0].Status)
	}
	if validators[0].Balance != uint64(params.BeaconConfig().DepositSize*params.BeaconConfig().Gwei) {
		t.Error("Failed to set validator balance")
	}
	if validators[1].Status != uint64(params.PendingWithdraw) {
		t.Errorf("Wanted status PendingWithdraw. Got: %d", validators[1].Status)
	}
	if validators[1].LatestStatusChangeSlot != params.BeaconConfig().MinWithdrawalPeriod+1 {
		t.Errorf("Failed to set validator lastest status change slot")
	}
	if validators[2].Status != uint64(params.Active) {
		t.Errorf("Wanted status Active. Got: %d", validators[2].Status)
	}
	if validators[2].Balance != uint64(params.BeaconConfig().DepositSize*params.BeaconConfig().Gwei) {
		t.Error("Failed to set validator balance")
	}
	if validators[3].Status != uint64(params.PendingWithdraw) {
		t.Errorf("Wanted status PendingWithdraw. Got: %d", validators[3].Status)
	}
	if validators[3].LatestStatusChangeSlot != params.BeaconConfig().MinWithdrawalPeriod+1 {
		t.Errorf("Failed to set validator lastest status change slot")
	}
	// Reach max validation rotation case, this validator couldn't be rotated.
	if validators[5].Status != uint64(params.PendingExit) {
		t.Errorf("Wanted status PendingExit. Got: %d", validators[5].Status)
	}
	if validators[7].Status != uint64(params.Withdrawn) {
		t.Errorf("Wanted status Withdrawn. Got: %d", validators[7].Status)
	}
	if validators[8].Status != uint64(params.Withdrawn) {
		t.Errorf("Wanted status Withdrawn. Got: %d", validators[8].Status)
	}
}

func TestValidatorMinDeposit(t *testing.T) {
	minDeposit := params.BeaconConfig().MinOnlineDepositSize * params.BeaconConfig().Gwei
	currentSlot := uint64(99)
	validators := []*pb.ValidatorRecord{
		{Status: uint64(params.Active), Balance: uint64(minDeposit) + 1},
		{Status: uint64(params.Active), Balance: uint64(minDeposit)},
		{Status: uint64(params.Active), Balance: uint64(minDeposit) - 1},
	}
	newValidatorRegistry := CheckValidatorMinDeposit(validators, currentSlot)
	if newValidatorRegistry[0].Status != uint64(params.Active) {
		t.Error("Validator should be active")
	}
	if newValidatorRegistry[1].Status != uint64(params.Active) {
		t.Error("Validator should be active")
	}
	if newValidatorRegistry[2].Status != uint64(params.PendingExit) {
		t.Error("Validator should be pending exit")
	}
	if newValidatorRegistry[2].LatestStatusChangeSlot != currentSlot {
		t.Errorf("Validator's lastest status change slot should be %d got %d", currentSlot, newValidatorRegistry[2].LatestStatusChangeSlot)
	}
}

func TestMinEmptyValidator(t *testing.T) {
	validators := []*pb.ValidatorRecord{
		{Status: uint64(params.Active)},
		{Status: uint64(params.Withdrawn)},
		{Status: uint64(params.Active)},
	}
	if minEmptyValidator(validators) != 1 {
		t.Errorf("Min vaidator index should be 1")
	}

	validators[1].Status = uint64(params.Active)
	if minEmptyValidator(validators) != -1 {
		t.Errorf("Min vaidator index should be -1")
	}
}

func TestDeepCopyValidatorRegistry(t *testing.T) {
	var validators []*pb.ValidatorRecord
	defaultValidator := &pb.ValidatorRecord{
		Pubkey:                 []byte{'k', 'e', 'y'},
		RandaoCommitmentHash32: []byte{'r', 'a', 'n', 'd', 'a', 'o'},
		Balance:                uint64(1e9),
		Status:                 uint64(params.Active),
		LatestStatusChangeSlot: 10,
	}
	for i := 0; i < 100; i++ {
		validators = append(validators, defaultValidator)
	}

	newValidatorSet := CopyValidatorRegistry(validators)

	defaultValidator.Pubkey = []byte{'n', 'e', 'w', 'k', 'e', 'y'}
	defaultValidator.RandaoCommitmentHash32 = []byte{'n', 'e', 'w', 'r', 'a', 'n', 'd', 'a', 'o'}
	defaultValidator.Balance = uint64(2e9)
	defaultValidator.Status = uint64(params.PendingExit)
	defaultValidator.LatestStatusChangeSlot = 5

	if len(newValidatorSet) != len(validators) {
		t.Fatalf("validator set length is unequal, copy of set failed: %d", len(newValidatorSet))
	}

	for i, validator := range newValidatorSet {
		if bytes.Equal(validator.Pubkey, defaultValidator.Pubkey) {
			t.Errorf("validator with index %d was unable to have their pubkey copied correctly %v", i, validator.Pubkey)
		}

		if bytes.Equal(validator.RandaoCommitmentHash32, defaultValidator.RandaoCommitmentHash32) {
			t.Errorf("validator with index %d was unable to have their randao commitment copied correctly %v", i, validator.RandaoCommitmentHash32)
		}

		if validator.Balance == defaultValidator.Balance {
			t.Errorf("validator with index %d was unable to have their balance copied correctly %d", i, validator.Balance)
		}

		if validator.Status == defaultValidator.Status {
			t.Errorf("validator with index %d was unable to have their status copied correctly %d", i, validator.Status)
		}

		if validator.LatestStatusChangeSlot == defaultValidator.LatestStatusChangeSlot {
			t.Errorf("validator with index %d was unable to have their lastest status change slot copied correctly %d", i, validator.LatestStatusChangeSlot)
		}
	}

}
