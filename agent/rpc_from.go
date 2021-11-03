/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package agent

import (
	"encoding/json"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/ns1labs/orb/fleet"
	"go.uber.org/zap"
	"time"
)

func (a *orbAgent) handleGroupMembership(rpc fleet.GroupMembershipRPCPayload) {
	// if this is the full list, reset all group subscriptions and subscribed to this list
	if rpc.FullList {
		a.unsubscribeGroupChannels()
		a.groupChannels = a.subscribeGroupChannels(rpc.Groups)
	} else {
		// otherwise, just add these subscriptions to the existing list
		successList := a.subscribeGroupChannels(rpc.Groups)
		a.groupChannels = append(a.groupChannels, successList...)
	}
}

func (a *orbAgent) handleAgentPolicies(rpc []fleet.AgentPolicyRPCPayload) {

	for _, payload := range rpc {
		a.policyManager.ManagePolicy(payload)
	}

	// heart beat with new policy status after application
	a.sendSingleHeartbeat(time.Now(), fleet.Online)

}

func (a *orbAgent) handleGroupRPCFromCore(client mqtt.Client, message mqtt.Message) {

	a.logger.Debug("Group RPC message from core", zap.String("topic", message.Topic()), zap.ByteString("payload", message.Payload()))

	var rpc fleet.RPC
	if err := json.Unmarshal(message.Payload(), &rpc); err != nil {
		a.logger.Error("error decoding RPC message from core", zap.Error(fleet.ErrSchemaMalformed))
		return
	}
	if rpc.SchemaVersion != fleet.CurrentRPCSchemaVersion {
		a.logger.Error("error decoding RPC message from core", zap.Error(fleet.ErrSchemaVersion))
		return
	}
	if rpc.Func == "" || rpc.Payload == nil {
		a.logger.Error("error decoding RPC message from core", zap.Error(fleet.ErrSchemaMalformed))
		return
	}

	// dispatch
	switch rpc.Func {
	case fleet.AgentPolicyRPCFunc:
		var r fleet.AgentPolicyRPC
		if err := json.Unmarshal(message.Payload(), &r); err != nil {
			a.logger.Error("error decoding agent policy message from core", zap.Error(fleet.ErrSchemaMalformed))
			return
		}
		a.handleAgentPolicies(r.Payload)
	case fleet.GroupRemovedRPCFunc:
		var r fleet.GroupRemovedRPC
		if err := json.Unmarshal(message.Payload(), &r); err != nil {
			a.logger.Error("error decoding agent group removal message from core", zap.Error(fleet.ErrSchemaMalformed))
			return
		}
		a.handleAgentGroupRemoval(r.Payload)
	default:
		a.logger.Warn("unsupported/unhandled core RPC, ignoring",
			zap.String("func", rpc.Func),
			zap.Any("payload", rpc.Payload))
	}

}

func (a *orbAgent) handleAgentGroupRemoval(rpc fleet.GroupRemovedRPCPayload) {
	a.unsubscribeGroupChannel(rpc.ChannelID)
}

func (a *orbAgent) handleDatasetRemoval(rpc fleet.DatasetRemovedRPCPayload) {
	a.unsubscribeGroupChannel(rpc.ChannelID)

	var index int
	for ag :=  range a.groupChannels{
		if a.groupChannels[ag] == rpc.ChannelID{
			index = ag
		}
	}

	fmt.Println(a.groupChannels)

	temp := make([]string, 0)
	temp = append(temp, a.groupChannels[:index]...)
	temp = append(temp, a.groupChannels[index+1:]...)

	a.groupChannels = temp
	fmt.Println(a.groupChannels)
}

func (a *orbAgent) handleRPCFromCore(client mqtt.Client, message mqtt.Message) {

	a.logger.Debug("RPC message from core", zap.String("topic", message.Topic()), zap.ByteString("payload", message.Payload()))

	var rpc fleet.RPC
	if err := json.Unmarshal(message.Payload(), &rpc); err != nil {
		a.logger.Error("error decoding RPC message from core", zap.Error(fleet.ErrSchemaMalformed))
		return
	}
	if rpc.SchemaVersion != fleet.CurrentRPCSchemaVersion {
		a.logger.Error("error decoding RPC message from core", zap.Error(fleet.ErrSchemaVersion))
		return
	}
	if rpc.Func == "" || rpc.Payload == nil {
		a.logger.Error("error decoding RPC message from core", zap.Error(fleet.ErrSchemaMalformed))
		return
	}

	// dispatch
	switch rpc.Func {
	case fleet.GroupMembershipRPCFunc:
		var r fleet.GroupMembershipRPC
		if err := json.Unmarshal(message.Payload(), &r); err != nil {
			a.logger.Error("error decoding group membership message from core", zap.Error(fleet.ErrSchemaMalformed))
			return
		}
		a.handleGroupMembership(r.Payload)
	case fleet.AgentPolicyRPCFunc:
		var r fleet.AgentPolicyRPC
		if err := json.Unmarshal(message.Payload(), &r); err != nil {
			a.logger.Error("error decoding agent policy message from core", zap.Error(fleet.ErrSchemaMalformed))
			return
		}
		a.handleAgentPolicies(r.Payload)
	case fleet.GroupRemovedRPCFunc:
		var r fleet.GroupRemovedRPC
		if err := json.Unmarshal(message.Payload(), &r); err != nil {
			a.logger.Error("error decoding agent group removal message from core", zap.Error(fleet.ErrSchemaMalformed))
			return
		}
		a.handleAgentGroupRemoval(r.Payload)
	case fleet.DatasetRemovedRPCFunc:
		var r fleet.DatasetRemovedRPC
		if err := json.Unmarshal(message.Payload(), &r); err != nil {
			a.logger.Error("error decoding dataset removal message from core", zap.Error(fleet.ErrSchemaMalformed))
			return
		}
		a.handleDatasetRemoval(r.Payload)
	default:
		a.logger.Warn("unsupported/unhandled core RPC, ignoring",
			zap.String("func", rpc.Func),
			zap.Any("payload", rpc.Payload))
	}

}
