package types

import (
	"fmt"

	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// IBC channel events
const (
	AttributeKeyConnectionID       = "connection_id"
	AttributeKeyPortID             = "port_id"
	AttributeKeyChannelID          = "channel_id"
	AttributeCounterpartyPortID    = "counterparty_port_id"
	AttributeCounterpartyChannelID = "counterparty_channel_id"

	EventTypeSendPacket        = "send_packet"
	EventTypeRecvPacket        = "recv_packet"
	EventTypeAcknowledgePacket = "acknowledge_packet"
	EventTypeCleanupPacket     = "cleanup_packet"
	EventTypeTimeoutPacket     = "timeout_packet"

	AttributeKeyData       = "packet_data"
	AttributeKeyTimeout    = "packet_timeout"
	AttributeKeySequence   = "packet_sequence"
	AttributeKeySrcPort    = "packet_src_port"
	AttributeKeySrcChannel = "packet_src_channel"
	AttributeKeyDstPort    = "packet_dst_port"
	AttributeKeyDstChannel = "packet_dst_channel"
)

// IBC channel events vars
var (
	EventTypeChannelOpenInit     = MsgChannelOpenInit{}.Type()
	EventTypeChannelOpenTry      = MsgChannelOpenTry{}.Type()
	EventTypeChannelOpenAck      = MsgChannelOpenAck{}.Type()
	EventTypeChannelOpenConfirm  = MsgChannelOpenConfirm{}.Type()
	EventTypeChannelCloseInit    = MsgChannelCloseInit{}.Type()
	EventTypeChannelCloseConfirm = MsgChannelCloseConfirm{}.Type()

	AttributeValueCategory = fmt.Sprintf("%s_%s", ibctypes.ModuleName, SubModuleName)
)
