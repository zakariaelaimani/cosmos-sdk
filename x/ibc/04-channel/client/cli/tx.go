package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

const (
	FlagOrdered    = "ordered"
	FlagIBCVersion = "ibc-version"
	FlagNode1      = "node1"
	FlagNode2      = "node2"
	FlagFrom1      = "from1"
	FlagFrom2      = "from2"
	FlagChainId2   = "chain-id2"
)

// GetTxCmd returns the transaction commands for IBC Connections
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	ics04ChannelTxCmd := &cobra.Command{
		Use:   "connection",
		Short: "IBC connection transaction subcommands",
	}

	ics04ChannelTxCmd.AddCommand(client.PostCommands(
		GetMsgChannelOpenInitCmd(storeKey, cdc),
		GetMsgChannelOpenTryCmd(storeKey, cdc),
		GetMsgChannelOpenAckCmd(storeKey, cdc),
		GetMsgChannelOpenConfirmCmd(storeKey, cdc),
		GetMsgChannelCloseInitCmd(storeKey, cdc),
		GetMsgChannelCloseConfirmCmd(storeKey, cdc),
		// GetCmdHandshake(storeKey, cdc),
	)...)

	return ics04ChannelTxCmd
}

// TODO: module needs to pass the capability key (i.e store key)

// GetMsgChannelOpenInitCmd returns the command to create a MsgChannelOpenInit transaction
func GetMsgChannelOpenInitCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-init [port-id] [channel-id] [counterparty-port-id] [counterparty-channel-id] [connection-hops]",
		Short: "Creates and sends a ChannelOpenInit message",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			portID := args[0]
			channelID := args[1]
			counterpartyPortID := args[2]
			counterpartyChannelID := args[3]
			hops := strings.Split(args[4], "/")
			order := channelOrder()
			version := viper.GetString(FlagIBCVersion)

			msg := types.NewMsgChannelOpenInit(
				portID, channelID, version, order, hops,
				counterpartyPortID, counterpartyChannelID, cliCtx.GetFromAddress(),
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().Bool(FlagOrdered, true, "Pass flag for opening ordered channels")
	cmd.Flags().String(FlagIBCVersion, "1.0.0", "supported IBC version")

	return cmd
}

// GetMsgChannelOpenTryCmd returns the command to create a MsgChannelOpenTry transaction
func GetMsgChannelOpenTryCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-try [port-id] [channel-id] [counterparty-port-id] [counterparty-channel-id] [connection-hops] [/path/to/proof-init.json] [proof-height]",
		Short: "Creates and sends a ChannelOpenTry message",
		Args:  cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			portID := args[0]
			channelID := args[1]
			counterpartyPortID := args[2]
			counterpartyChannelID := args[3]
			hops := strings.Split(args[4], "/")
			order := channelOrder()
			version := viper.GetString(FlagIBCVersion) // TODO: diferenciate between channel and counterparty versions

			var proof commitment.ProofI
			if err := cdc.UnmarshalJSON([]byte(args[5]), &proof); err != nil {
				fmt.Fprintf(os.Stderr, "failed to unmarshall input into struct, checking for file...")
				contents, err := ioutil.ReadFile(args[5])
				if err != nil {
					return fmt.Errorf("error opening proof file: %v", err)
				}
				if err := cdc.UnmarshalJSON(contents, &proof); err != nil {
					return fmt.Errorf("error unmarshalling proof file: %v", err)
				}
			}

			proofHeight, err := strconv.ParseInt(args[6], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgChannelOpenTry(
				portID, channelID, version, order, hops,
				counterpartyPortID, counterpartyChannelID, version,
				proof, uint64(proofHeight), cliCtx.GetFromAddress(),
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd.Flags().Bool(FlagOrdered, true, "Pass flag for opening ordered channels")
	cmd.Flags().String(FlagIBCVersion, "1.0.0", "supported IBC version")

	return cmd
}

// GetMsgChannelOpenAckCmd returns the command to create a MsgChannelOpenAck transaction
func GetMsgChannelOpenAckCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-ack [port-id] [channel-id] [/path/to/proof-try.json] [proof-height]",
		Short: "Creates and sends a ChannelOpenAck message",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			portID := args[0]
			channelID := args[1]
			version := viper.GetString(FlagIBCVersion) // TODO: diferenciate between channel and counterparty versions

			var proof commitment.ProofI
			if err := cdc.UnmarshalJSON([]byte(args[2]), &proof); err != nil {
				fmt.Fprintf(os.Stderr, "failed to unmarshall input into struct, checking for file...")
				contents, err := ioutil.ReadFile(args[2])
				if err != nil {
					return fmt.Errorf("error opening proof file: %v", err)
				}
				if err := cdc.UnmarshalJSON(contents, &proof); err != nil {
					return fmt.Errorf("error unmarshalling proof file: %v", err)
				}
			}

			proofHeight, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgChannelOpenAck(
				portID, channelID, version, proof, uint64(proofHeight), cliCtx.GetFromAddress(),
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
	cmd.Flags().String(FlagIBCVersion, "1.0.0", "supported IBC version")
	return cmd
}

// GetMsgChannelOpenConfirmCmd returns the command to create a MsgChannelOpenConfirm transaction
func GetMsgChannelOpenConfirmCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "open-confirm [port-id] [channel-id] [/path/to/proof-ack.json] [proof-height]",
		Short: "Creates and sends a ChannelOpenConfirm message",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			portID := args[0]
			channelID := args[1]

			var proof commitment.ProofI
			if err := cdc.UnmarshalJSON([]byte(args[2]), &proof); err != nil {
				fmt.Fprintf(os.Stderr, "failed to unmarshall input into struct, checking for file...")
				contents, err := ioutil.ReadFile(args[2])
				if err != nil {
					return fmt.Errorf("error opening proof file: %v", err)
				}
				if err := cdc.UnmarshalJSON(contents, &proof); err != nil {
					return fmt.Errorf("error unmarshalling proof file: %v", err)
				}
			}

			proofHeight, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgChannelOpenConfirm(
				portID, channelID, proof, uint64(proofHeight), cliCtx.GetFromAddress(),
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetMsgChannelCloseInitCmd returns the command to create a MsgChannelCloseInit transaction
func GetMsgChannelCloseInitCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "close-init [port-id] [channel-id]",
		Short: "Creates and sends a ChannelCloseInit message",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			portID := args[0]
			channelID := args[1]

			msg := types.NewMsgChannelCloseInit(portID, channelID, cliCtx.GetFromAddress())
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetMsgChannelCloseConfirmCmd returns the command to create a MsgChannelCloseConfirm transaction
func GetMsgChannelCloseConfirmCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "close-confirm [port-id] [channel-id] [/path/to/proof-init.json] [proof-height]",
		Short: "Creates and sends a ChannelCloseConfirm message",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			portID := args[0]
			channelID := args[1]

			var proof commitment.ProofI
			if err := cdc.UnmarshalJSON([]byte(args[2]), &proof); err != nil {
				fmt.Fprintf(os.Stderr, "failed to unmarshall input into struct, checking for file...")
				contents, err := ioutil.ReadFile(args[2])
				if err != nil {
					return fmt.Errorf("error opening proof file: %v", err)
				}
				if err := cdc.UnmarshalJSON(contents, &proof); err != nil {
					return fmt.Errorf("error unmarshalling proof file: %v", err)
				}
			}

			proofHeight, err := strconv.ParseInt(args[3], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgChannelCloseConfirm(
				portID, channelID, proof, uint64(proofHeight), cliCtx.GetFromAddress(),
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// func GetCmdHandshake(storeKey string, cdc *codec.Codec) *cobra.Command {
// 	cmd := &cobra.Command{
// 		Use:   "handshake",
// 		Short: "initiate connection handshake between two chains",
// 		Long: strings.TrimSpace(
// 			fmt.Sprintf(`initialize a connection on chain A with a given counterparty chain B:

// Example:
// $ %s tx ibc channel handshake [client-id] [port-id] [chan-id] [conn-id] [cp-client-id] [cp-port-id] [cp-chain-id] [cp-conn-id]
// 		`, version.ClientName),
// 		Args:  cobra.ExactArgs(6),
// 		// Args: []string{portid1, chanid1, connid1, portid2, chanid2, connid2}
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			// --chain-id values for each chain
// 			cid1 := viper.GetString(flags.FlagChainID)
// 			cid2 := viper.GetString(FlagChainId2)

// 			// --from values for each wallet
// 			from1 := viper.GetString(FlagFrom1)
// 			from2 := viper.GetString(FlagFrom2)

// 			// --node values for each RPC
// 			node1 := viper.GetString(FlagNode1)
// 			node2 := viper.GetString(FlagNode2)

// 			// client IDs
// 			clientid1 := args[0]
// 			clientid2 := args[4]

// 			// port IDs
// 			portid1 := args[1]
// 			portid2 := args[5]

// 			// channel IDs
// 			chanid1 := args[2]
// 			chanid2 := args[6]

// 			// connection IDs
// 			connid1 := args[3]
// 			connid2 := args[7]

// 			// Create txbldr, clictx, querier for cid1
// 			viper.Set(flags.FlagChainID, cid1)
// 			txBldr1 := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
// 			ctx1 := context.NewCLIContextIBC(from1, cid1, node1).WithCodec(cdc).
// 				WithBroadcastMode(flags.BroadcastBlock)

// 			// Create txbldr, clictx, querier for cid2
// 			viper.Set(flags.FlagChainID, cid2)
// 			txBldr2 := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
// 			ctx2 := context.NewCLIContextIBC(from2, cid2, node2).WithCodec(cdc).
// 				WithBroadcastMode(flags.BroadcastBlock)

// 			// get passphrase for key from1
// 			passphrase1, err := keys.GetPassphrase(from1)
// 			if err != nil {
// 				return err
// 			}

// 			// get passphrase for key from2
// 			passphrase2, err := keys.GetPassphrase(from2)
// 			if err != nil {
// 				return err
// 			}

// 			// Create channel objects
// 			// chan1 := types.NewChannel(types.State(byte(0)), types.UNORDERED, types.NewCounterparty(portid2, chanid2), []string{connid2}, "v1.0.0")
// 			// chan2 := types.NewChannel(types.State(byte(0)), types.UNORDERED, types.NewCounterparty(portid1, chanid1), []string{connid1}, "v1.0.0")

// 			// TODO: check state and if not Idle continue existing process
// 			viper.Set(flags.FlagChainID, cid1)
// 			msgOpenInit := types.NewMsgChannelOpenInit(portid1, chanid1, "v1.0.0", channelOrder(), []string{connid2}, portid2, chanid2, ctx1.GetFromAddress())
// 			fmt.Printf("%v <- %-14v", cid1, msgOpenInit.Type())
// 			res, err := utils.CompleteAndBroadcastTx(txBldr1, ctx1, []sdk.Msg{msgOpenInit}, passphrase1)
// 			if err != nil || !res.IsOK() {
// 				fmt.Println(res)
// 				return err
// 			}
// 			fmt.Printf(" [OK] txid(%v) portid(%v) chanid(%v)\n", res.TxHash, portid1, chanid1)

// 			// Another block has to be passed after msginit is commited
// 			// to retrieve the correct proofs
// 			time.Sleep(8 * time.Second)

// 			header, err := tendermint.GetHeader(ctx1)
// 			if err != nil {
// 				return err
// 			}

// 			viper.Set(flags.FlagChainID, cid2)
// 			msgUpdateClient := ibcclient.NewMsgUpdateClient(clientid2, header, ctx2.GetFromAddress())
// 			fmt.Printf("%v <- %-14v", cid2, msgUpdateClient.Type())
// 			res, err = utils.CompleteAndBroadcastTx(txBldr2, ctx2, []sdk.Msg{msgUpdateClient}, passphrase2)
// 			if err != nil || !res.IsOK() {
// 				fmt.Println(res)
// 				return err
// 			}
// 			fmt.Printf(" [OK] txid(%v) client(%v)\n", res.TxHash, clientid2)

// 			// viper.Set(flags.FlagChainID, cid1)
// 			// proofs, err := queryProofs(ctx1.WithHeight(header.Height-1), connID1, storeKey)
// 			// if err != nil {
// 			// 	return err
// 			// }

// 			// csProof, err := queryConsensusStateProof(ctx2.WithHeight(header.Height-1), clientID2)
// 			// if err != nil {
// 			// 	return err
// 			// }

// 			msgOpenTry := channel.NewMsgOpenTry(portid2, chanid2, chan2, []commitment.Proof{pchan, pstate}, uint64(header.Height), ctx2.GetFromAddress())
// 			fmt.Printf("%v <- %-14v", cid2, msgOpenTry.Type())
// 			res, err = utils.CompleteAndBroadcastTx(txBldr2, ctx2, []sdk.Msg{msgOpenTry}, passphrase2)
// 			if err != nil || !res.IsOK() {
// 				fmt.Println(res)
// 				return err
// 			}
// 			fmt.Printf(" [OK] txid(%v) portid(%v) chanid(%v)\n", res.TxHash, portid2, chanid2)

// 			// Another block has to be passed after msginit is commited
// 			// to retrieve the correct proofs
// 			time.Sleep(8 * time.Second)

// 			header, err = getHeader(ctx2)
// 			if err != nil {
// 				return err
// 			}

// 			viper.Set(flags.FlagChainID, cid1)
// 			msgUpdateClient = client.NewMsgUpdateClient(clientid1, header, ctx1.GetFromAddress())
// 			fmt.Printf("%v <- %-14v", cid1, msgUpdateClient.Type())
// 			res, err = utils.CompleteAndBroadcastTx(txBldr1, ctx1, []sdk.Msg{msgUpdateClient}, passphrase1)
// 			if err != nil || !res.IsOK() {
// 				fmt.Println(res)
// 				return err
// 			}
// 			fmt.Printf(" [OK] txid(%v) client(%v)\n", res.TxHash, clientid1)

// 			q2 = state.NewCLIQuerier(ctx2.WithHeight(header.Height - 1))
// 			_, pchan, err = obj2.ChannelCLI(q2)
// 			if err != nil {
// 				return err
// 			}
// 			_, pstate, err = obj2.StageCLI(q2)
// 			if err != nil {
// 				return err
// 			}

// 			msgOpenAck := channel.NewMsgOpenAck(portid1, chanid1, []commitment.Proof{pchan, pstate}, uint64(header.Height), ctx1.GetFromAddress())
// 			fmt.Printf("%v <- %-14v", cid1, msgOpenAck.Type())
// 			res, err = utils.CompleteAndBroadcastTx(txBldr1, ctx1, []sdk.Msg{msgOpenAck}, passphrase1)
// 			if err != nil || !res.IsOK() {
// 				fmt.Println(res)
// 				return err
// 			}
// 			fmt.Printf(" [OK] txid(%v) portid(%v) chanid(%v)\n", res.TxHash, portid1, chanid1)

// 			// Another block has to be passed after msginit is commited
// 			// to retrieve the correct proofs
// 			time.Sleep(8 * time.Second)

// 			header, err = getHeader(ctx1)
// 			if err != nil {
// 				return err
// 			}

// 			viper.Set(flags.FlagChainID, cid2)
// 			msgUpdateClient = client.NewMsgUpdateClient(clientid2, header, ctx2.GetFromAddress())
// 			fmt.Printf("%v <- %-14v", cid2, msgUpdateClient.Type())
// 			res, err = utils.CompleteAndBroadcastTx(txBldr2, ctx2, []sdk.Msg{msgUpdateClient}, passphrase2)
// 			if err != nil || !res.IsOK() {
// 				fmt.Println(res)
// 				return err
// 			}
// 			fmt.Printf(" [OK] txid(%v) client(%v)\n", res.TxHash, clientid2)

// 			q1 = state.NewCLIQuerier(ctx1.WithHeight(header.Height - 1))
// 			_, pstate, err = obj1.StageCLI(q1)
// 			if err != nil {
// 				return err
// 			}

// 			msgOpenConfirm := channel.NewMsgOpenConfirm(portid2, chanid2, []commitment.Proof{pstate}, uint64(header.Height), ctx2.GetFromAddress())
// 			fmt.Printf("%v <- %-14v", cid2, msgOpenConfirm.Type())
// 			res, err = utils.CompleteAndBroadcastTx(txBldr2, ctx2, []sdk.Msg{msgOpenConfirm}, passphrase2)
// 			if err != nil || !res.IsOK() {
// 				fmt.Println(res)
// 				return err
// 			}
// 			fmt.Printf(" [OK] txid(%v) portid(%v) chanid(%v)\n", res.TxHash, portid2, chanid2)

// 			return nil
// 		},
// 	}

// 	cmd.Flags().String(FlagNode1, "tcp://localhost:26657", "RPC port for the first chain")
// 	cmd.Flags().String(FlagNode2, "tcp://localhost:26657", "RPC port for the second chain")
// 	cmd.Flags().String(FlagFrom1, "", "key in local keystore for first chain")
// 	cmd.Flags().String(FlagFrom2, "", "key in local keystore for second chain")
// 	cmd.Flags().String(FlagChainId2, "", "chain-id for the second chain")
// 	cmd.Flags().Bool(FlagOrdered, true, "Pass flag for opening ordered channels")

// 	cmd.MarkFlagRequired(FlagFrom1)
// 	cmd.MarkFlagRequired(FlagFrom2)

// 	return cmd
// }

func channelOrder() types.Order {
	if viper.GetBool(FlagOrdered) {
		return types.ORDERED
	}
	return types.UNORDERED
}