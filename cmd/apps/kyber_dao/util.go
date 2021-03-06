package kyberdao

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/spf13/cobra"
	"github.com/tranvictor/ethutils"
	"github.com/tranvictor/jarvis/config"
	"github.com/tranvictor/jarvis/util"
)

func PrintENV() {
	fmt.Printf("--ENV-----------------------------------------------------------------------------------------------\n")
	fmt.Printf("Dao contract: %s\n", DaoContract)
	fmt.Printf("Staking contract: %s\n", StakingContract)
	fmt.Printf("FeeHandler contract: %s\n", FeeHandler)
	fmt.Printf("Campaign creator: %s\n", CampaignCreator)
	fmt.Printf("Epoch duration: %d\n", EpochDuration)
	fmt.Printf("KNC: %s\n", KNCContract)
	fmt.Printf("----------------------------------------------------------------------------------------------------\n")
}

func optionStr(options []*big.Int) string {
	opStrs := []string{}
	for _, o := range options {
		opStrs = append(opStrs, hexutil.EncodeBig(o))
	}
	return fmt.Sprintf("[%s]", strings.Join(opStrs, ","))
}

func PromptBRROption(prompter string) *big.Int {
	opStr := util.PromptInputWithValidation(prompter, func(str string) error {
		parts := strings.Split(str, ",")
		if len(parts) != 2 {
			return fmt.Errorf("Please input exactly 2 number in bps separated by a comma")
		}
		rebate, err := strconv.ParseUint(strings.Trim(parts[0], " "), 10, 64)
		if err != nil {
			return fmt.Errorf("Can't parse %s to number, %w", parts[0], err)
		}
		reward, err := strconv.ParseUint(strings.Trim(parts[1], " "), 10, 64)
		if err != nil {
			return fmt.Errorf("Can't parse %s to number, %w", parts[1], err)
		}
		if rebate+reward > 10000 {
			return fmt.Errorf("Can't have rebate + reward > 10000 bps")
		}
		return nil
	})

	parts := strings.Split(opStr, ",")
	rebate, _ := strconv.ParseUint(strings.Trim(parts[0], " "), 10, 64)
	reward, _ := strconv.ParseUint(strings.Trim(parts[1], " "), 10, 64)
	rebateBig := big.NewInt(int64(rebate))
	temp := big.NewInt(0).Lsh(rebateBig, 128)
	return big.NewInt(0).Add(temp, big.NewInt(int64(reward)))
}

func PrintCampaignInformation(cmd *cobra.Command, info *CampaignRelatedInfo, currentBlock uint64) {
	fmt.Printf("----------------------------------------------------------------------------------------------------\n")
	fmt.Printf("Current block: %d\n", currentBlock)
	fmt.Printf("Campaign ID: %d\n", info.Campaign.ID)
	fmt.Printf("Type: %s\n", info.Campaign.Type())
	fmt.Printf("Duration: block %d -> %d, %d blocks\n",
		info.Campaign.StartBlock.Uint64(),
		info.Campaign.EndBlock.Uint64(),
		info.Campaign.EndBlock.Uint64()-info.Campaign.StartBlock.Uint64())
	if currentBlock < info.Campaign.StartBlock.Uint64() {
		timeLeft := util.CalculateTimeDurationFromBlock(config.Network, currentBlock, info.Campaign.StartBlock.Uint64())
		fmt.Printf("Start in: %s\n", timeLeft.String())
	} else {
		timeLeft := util.CalculateTimeDurationFromBlock(config.Network, currentBlock, info.Campaign.EndBlock.Uint64())
		if timeLeft == 0 {
			fmt.Printf("Time left: ENDED")
			if !info.Campaign.HasWinningOption() {
				fmt.Printf(" - No winning option")
			}
			fmt.Printf("\n")
		} else {
			fmt.Printf("Time left: %s\n", timeLeft.String())
		}
	}
	if len(info.Campaign.LinkStr()) == 0 {
		fmt.Printf("For more information: No link is provided.\n")
	} else {
		fmt.Printf("For more information: %s\n", info.Campaign.LinkStr())
	}
	fmt.Printf("\n%d options to vote (total vote: %d):\n", len(info.Campaign.Options), info.Campaign.TotalPoints)
	for i, op := range info.Campaign.Options {
		fmt.Printf("    %d. %s\n", i+1, info.Campaign.VerboseOption(op, uint64(i+1), info.VotedID))
	}
	fmt.Printf("\n")
}

func PrintStakeInformation(cmd *cobra.Command, info *StakeRelatedInfo) {
	isCurrentEpoch := info.CurrentEpoch == info.Epoch

	if info.CurrentEpoch == info.Epoch {
		cmd.Printf("Working with Epoch: %d (current epoch)\n", info.Epoch)
	} else if info.CurrentEpoch > info.Epoch {
		cmd.Printf("Working with Epoch: %d (epoch in the past, current epoch: %d)\n", info.Epoch, info.CurrentEpoch)
	} else {
		cmd.Printf("Working with Epoch: %d (epoch in the future). Abort!\n", info.Epoch)
		return
	}

	cmd.Printf("Staker: %s\n", util.VerboseAddress(info.Staker))
	stakef := ethutils.BigToFloat(info.Stake, 18)
	cmd.Printf("Your stake: %f KNC (%s)\n", stakef, info.Stake)

	balancef := ethutils.BigToFloat(info.Balance, 18)
	if isCurrentEpoch {
		cmd.Printf("Your pending stake (can withdraw without any penalty): %f KNC (%s)\n", ethutils.BigToFloat(info.PendingStake, 18), info.PendingStake)
	}
	cmd.Printf("Available KNC to stake: %f KNC (%s)\n", balancef, info.Balance)

	if strings.ToLower(info.Staker) == strings.ToLower(info.Representative) {
		cmd.Printf("Your representative: None\n")
	} else {
		cmd.Printf("Your representative: %s (Contact him to get your reward if you have some)\n", util.VerboseAddress(info.Representative))
	}
	cmd.Printf("Stake other people delegated to you: %f KNC\n", ethutils.BigToFloat(info.DelegatedStake, 18))
}

func percentage(a, n *big.Int) float64 {
	if n == nil || n.Cmp(big.NewInt(0)) == 0 {
		return 0
	}
	af := ethutils.BigToFloat(a, 18)
	nf := ethutils.BigToFloat(n, 18)
	return af / nf * 100
}
