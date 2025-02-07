/*
   Copyright 2022 Erigon-Lightclient contributors
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at
       http://www.apache.org/licenses/LICENSE-2.0
   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package handlers

import (
	"context"

	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon/cl/clparams"
	"github.com/ledgerwatch/erigon/cl/cltypes"
	"github.com/ledgerwatch/erigon/cmd/sentinel/sentinel/peers"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
)

type ConsensusHandlers struct {
	handlers      map[protocol.ID]network.StreamHandler
	host          host.Host
	peers         *peers.Peers
	metadata      *cltypes.MetadataV2
	beaconConfig  *clparams.BeaconChainConfig
	genesisConfig *clparams.GenesisConfig
	ctx           context.Context

	db kv.RoDB // Read stuff from database to answer
}

const (
	SuccessfulResponsePrefix = 0x00
	ResourceUnavaiablePrefix = 0x03
)

var NoRequestHandlers = map[string]bool{
	MetadataProtocolV1:          true,
	MetadataProtocolV2:          true,
	LightClientFinalityUpdateV1: true,
}

func NewConsensusHandlers(ctx context.Context, db kv.RoDB, host host.Host,
	peers *peers.Peers, beaconConfig *clparams.BeaconChainConfig, genesisConfig *clparams.GenesisConfig, metadata *cltypes.MetadataV2) *ConsensusHandlers {
	c := &ConsensusHandlers{
		peers:         peers,
		host:          host,
		metadata:      metadata,
		db:            db,
		genesisConfig: genesisConfig,
		beaconConfig:  beaconConfig,
		ctx:           ctx,
	}
	c.handlers = map[protocol.ID]network.StreamHandler{
		protocol.ID(PingProtocolV1):                c.pingHandler,
		protocol.ID(GoodbyeProtocolV1):             c.goodbyeHandler,
		protocol.ID(StatusProtocolV1):              c.statusHandler,
		protocol.ID(MetadataProtocolV1):            c.metadataV1Handler,
		protocol.ID(MetadataProtocolV2):            c.metadataV2Handler,
		protocol.ID(BeaconBlocksByRangeProtocolV1): c.blocksByRangeHandler,
		protocol.ID(BeaconBlocksByRootProtocolV1):  c.beaconBlocksByRootHandler,
		protocol.ID(LightClientFinalityUpdateV1):   c.lightClientFinalityUpdateHandler,
		protocol.ID(LightClientOptimisticUpdateV1): c.lightClientOptimisticUpdateHandler,
	}
	return c
}

func (c *ConsensusHandlers) Start() {
	for id, handler := range c.handlers {
		c.host.SetStreamHandler(id, handler)
	}
}
