// Copyright 2020 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package chaosdaemon

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"

	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

const (
	iptablesCmd = "iptables"

	iptablesBadRuleErr           = "Bad rule (does a matching rule exist in that chain?)."
	iptablesIPSetNotExistErr     = "doesn't exist."
	iptablesChainAlreadyExistErr = "iptables: Chain already exists."
)

func (s *daemonServer) SetIptablesChains(ctx context.Context, req *pb.IptablesChainsRequest) (*empty.Empty, error) {
	log.Info("Set iptables chains", "request", req)

	pid, err := s.crClient.GetPidFromContainerID(ctx, req.ContainerId)
	if err != nil {
		log.Error(err, "error while getting PID")
		return nil, err
	}

	nsPath := GetNsPath(pid, netNS)

	iptables := buildIptablesClient(ctx, nsPath)
	err = iptables.initializeEnv()
	if err != nil {
		return nil, err
	}

	err = iptables.setIptablesChains(req.Chains)
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

type iptablesClient struct {
	ctx    context.Context
	nsPath string
}

type iptablesChain struct {
	Name  string
	Rules []string
}

func buildIptablesClient(ctx context.Context, nsPath string) iptablesClient {
	return iptablesClient{
		ctx,
		nsPath,
	}
}

func (iptables *iptablesClient) setIptablesChains(chains []*pb.Chain) error {
	for _, chain := range chains {
		err := iptables.setIptablesChain(chain)
		if err != nil {
			return err
		}
	}

	return nil
}

func (iptables *iptablesClient) setIptablesChain(chain *pb.Chain) error {
	var matchPart string
	if chain.Direction == pb.Chain_INPUT {
		matchPart = "src"
	} else if chain.Direction == pb.Chain_OUTPUT {
		matchPart = "dst"
	} else {
		return fmt.Errorf("unknown chain direction %d", chain.Direction)
	}

	rules := []string{}
	for _, ipset := range chain.Ipsets {
		rules = append(rules, fmt.Sprintf("-A %s -m set --match-set %s %s -j DROP -w 5", chain.Name, ipset, matchPart))
	}
	err := iptables.createNewChain(&iptablesChain{
		Name:  chain.Name,
		Rules: rules,
	})
	if err != nil {
		return err
	}

	if chain.Direction == pb.Chain_INPUT {
		err := iptables.ensureRule(&iptablesChain{
			Name: "CHAOS-INPUT",
		}, "-A CHAOS-INPUT -j "+chain.Name)
		if err != nil {
			return err
		}
	} else if chain.Direction == pb.Chain_OUTPUT {
		iptables.ensureRule(&iptablesChain{
			Name: "CHAOS-OUTPUT",
		}, "-A CHAOS-OUTPUT -j "+chain.Name)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unknown direction %d", chain.Direction)
	}
	return nil
}

func (iptables *iptablesClient) initializeEnv() error {
	for _, direction := range []string{"INPUT", "OUTPUT"} {
		chainName := "CHAOS-" + direction

		err := iptables.createNewChain(&iptablesChain{
			Name:  chainName,
			Rules: []string{},
		})
		if err != nil {
			return err
		}

		iptables.ensureRule(&iptablesChain{
			Name:  direction,
			Rules: []string{},
		}, "-A "+direction+" -j "+chainName)
	}

	return nil
}

// createNewChain will cover existing chain
func (iptables *iptablesClient) createNewChain(chain *iptablesChain) error {
	cmd := withNetNS(iptables.ctx, iptables.nsPath, iptablesCmd, "-N", chain.Name)
	out, err := cmd.CombinedOutput()

	if (err == nil && len(out) == 0) ||
		(err != nil && strings.Contains(string(out), iptablesChainAlreadyExistErr)) {
		// Successfully create a new chain
		return iptables.deleteAndWriteRules(chain)
	}

	return encodeOutputToError(out, err)
}

// deleteAndWriteRules will remove all existing function in the chain
// and replace with the new settings
func (iptables *iptablesClient) deleteAndWriteRules(chain *iptablesChain) error {

	// This chain should already exist
	err := iptables.flushIptablesChain(chain)
	if err != nil {
		return err
	}

	for _, rule := range chain.Rules {
		err := iptables.ensureRule(chain, rule)
		if err != nil {
			return err
		}
	}

	return nil
}

func (iptables *iptablesClient) ensureRule(chain *iptablesChain, rule string) error {
	cmd := withNetNS(iptables.ctx, iptables.nsPath, iptablesCmd, "-S", chain.Name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return encodeOutputToError(out, err)
	}

	if strings.Contains(string(out), rule) {
		// The required rule already exist in chain
		return nil
	}

	cmd = withNetNS(iptables.ctx, iptables.nsPath, iptablesCmd, strings.Split(rule, " ")...)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return encodeOutputToError(out, err)
	}

	return nil
}

func (iptables *iptablesClient) flushIptablesChain(chain *iptablesChain) error {
	cmd := withNetNS(iptables.ctx, iptables.nsPath, iptablesCmd, "-F", chain.Name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return encodeOutputToError(out, err)
	}

	return nil
}

func encodeOutputToError(output []byte, err error) error {
	return fmt.Errorf("error code: %v, msg: %s", err, string(output))
}
