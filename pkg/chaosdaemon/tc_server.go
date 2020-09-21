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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"

	"github.com/golang/protobuf/ptypes/empty"
)

const (
	ruleNotExist             = "Cannot delete qdisc with handle of zero."
	ruleNotExistLowerVersion = "RTNETLINK answers: No such file or directory"
)

func generateQdiscArgs(action string, qdisc *pb.Qdisc) ([]string, error) {

	if qdisc == nil {
		return nil, fmt.Errorf("qdisc is required")
	}

	if qdisc.Type == "" {
		return nil, fmt.Errorf("qdisc.Type is required")
	}

	args := []string{"qdisc", action, "dev", "eth0"}

	if qdisc.Parent == nil {
		args = append(args, "root")
	} else if qdisc.Parent.Major == 1 && qdisc.Parent.Minor == 0 {
		args = append(args, "root")
	} else {
		args = append(args, "parent", fmt.Sprintf("%d:%d", qdisc.Parent.Major, qdisc.Parent.Minor))
	}

	if qdisc.Handle == nil {
		args = append(args, "handle", fmt.Sprintf("%d:%d", 1, 0))
	} else {
		args = append(args, "handle", fmt.Sprintf("%d:%d", qdisc.Handle.Major, qdisc.Handle.Minor))
	}

	args = append(args, qdisc.Type)

	if qdisc.Args != nil {
		args = append(args, qdisc.Args...)
	}

	return args, nil
}

func (s *daemonServer) SetTcs(ctx context.Context, in *pb.TcsRequest) (*empty.Empty, error) {
	log.Info("handling tc request", "tcs", in)

	pid, err := s.crClient.GetPidFromContainerID(ctx, in.ContainerId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get pid from containerID error: %v", err)
	}
	nsPath := GetNsPath(pid, bpm.NetNS)
	client := buildTcClient(ctx, nsPath)
	err = client.flush()
	if err != nil {
		log.Error(err, "error while flushing client")
		return &empty.Empty{}, err
	}

	// tc rules are split into two different kinds according to whether it has filter.
	// all tc rules without filter are called `globalTc` and the tc rules with filter will be called `filterTc`.
	// the `globalTc` rules will be piped one by one from root, and the last `globalTc` will be connected with a PRIO
	// qdisc, which has `3 + len(filterTc)` bands. Then the 4.. bands will be connected to `filterTc` and a filter will
	// be setuped to flow packet from PRIO qdisc to it.

	// for example, four tc rules:
	// - NETEM: 50ms latency without filter
	// - NETEM: 100ms latency without filter
	// - NETEM: 50ms latency with filter ipset A
	// - NETEM: 100ms latency with filter ipset B
	// will generate tc rules:
	//	tc qdisc del dev eth0 root
	//  tc qdisc add dev eth0 root handle 1: netem delay 50000
	//  tc qdisc add dev eth0 parent 1: handle 2: netem delay 100000
	//  tc qdisc add dev eth0 parent 2: handle 3: prio bands 5 priomap 1 2 2 2 1 2 0 0 1 1 1 1 1 1 1 1
	//  tc qdisc add dev eth0 parent 3:1 handle 4: sfq
	//  tc qdisc add dev eth0 parent 3:2 handle 5: sfq
	//  tc qdisc add dev eth0 parent 3:3 handle 6: sfq
	//  tc qdisc add dev eth0 parent 3:4 handle 7: netem delay 50000
	//  tc filter add dev eth0 parent 3: basic match ipset(A dst) classid 3:4
	//  tc qdisc add dev eth0 parent 3:5 handle 8: netem delay 100000
	//  tc filter add dev eth0 parent 3: basic match ipset(B dst) classid 3:5

	globalTc := []*pb.Tc{}
	filterTc := map[string][]*pb.Tc{}

	for _, tc := range in.Tcs {
		if tc.Ipset == "" {
			globalTc = append(globalTc, tc)
		} else {
			// TODO: support multiple tc with one ipset
			filterTc[tc.Ipset] = append(filterTc[tc.Ipset], tc)
		}
	}

	for index, tc := range globalTc {
		parentArg := "root"
		if index > 0 {
			parentArg = fmt.Sprintf("parent %d:", index)
		}

		handleArg := fmt.Sprintf("handle %d:", index+1)

		err := client.addTc(parentArg, handleArg, tc)
		if err != nil {
			log.Error(err, "error while adding tc")
			return &empty.Empty{}, err
		}
	}

	parent := len(globalTc)
	band := 3 + len(filterTc) // 3 handlers for normal sfq on prio qdisc
	err = client.addPrio(parent, band)
	if err != nil {
		log.Error(err, "error while adding prio")
		return &empty.Empty{}, err
	}

	parent++

	index := 0
	currentHandler := parent + 3 // 3 handlers for sfq on prio qdisc
	for ipset, tcs := range filterTc {
		for i, tc := range tcs {
			parentArg := fmt.Sprintf("parent %d:%d", parent, index+4)
			if i > 0 {
				parentArg = fmt.Sprintf("parent %d:", currentHandler)
			}

			currentHandler++
			handleArg := fmt.Sprintf("handle %d:", currentHandler)

			err := client.addTc(parentArg, handleArg, tc)
			if err != nil {
				log.Error(err, "error while adding tc")
				return &empty.Empty{}, err
			}
		}

		parentArg := fmt.Sprintf("parent %d:", parent)
		classid := fmt.Sprintf("classid %d:%d", parent, index+4)
		err = client.addFilter(parentArg, classid, ipset)
		if err != nil {
			log.Error(err, "error while adding filter")
			return &empty.Empty{}, err
		}

		index++
	}
	// TODO: following qdisc

	return &empty.Empty{}, nil
}

type tcClient struct {
	ctx    context.Context
	nsPath string
}

func buildTcClient(ctx context.Context, nsPath string) tcClient {
	return tcClient{
		ctx,
		nsPath,
	}
}

func (c *tcClient) flush() error {
	cmd := bpm.DefaultProcessBuilder("tc", "qdisc", "del", "dev", "eth0", "root").SetNetNS(c.nsPath).SetContext(c.ctx).Build()
	output, err := cmd.CombinedOutput()
	if err != nil {
		if (!strings.Contains(string(output), ruleNotExistLowerVersion)) && (!strings.Contains(string(output), ruleNotExist)) {
			return encodeOutputToError(output, err)
		}
	}
	return nil
}

func (c *tcClient) addTc(parentArg string, handleArg string, tc *pb.Tc) error {
	log.Info("add tc", "tc", tc)

	if tc.Type == pb.Tc_BANDWIDTH {

		if tc.Tbf == nil {
			return fmt.Errorf("tbf is nil while type is BANDWIDTH")
		}
		err := c.addTbf(parentArg, handleArg, tc.Tbf)
		if err != nil {
			return err
		}

	} else if tc.Type == pb.Tc_NETEM {

		if tc.Netem == nil {
			return fmt.Errorf("netem is nil while type is NETEM")
		}
		err := c.addNetem(parentArg, handleArg, tc.Netem)
		if err != nil {
			return err
		}

	} else {
		return fmt.Errorf("unknown tc qdisc type")
	}

	return nil
}

func (c *tcClient) addPrio(parent int, band int) error {
	log.Info("adding prio", "parent", parent)

	parentArg := "root"
	if parent > 0 {
		parentArg = fmt.Sprintf("parent %d:", parent)
	}
	args := fmt.Sprintf("qdisc add dev eth0 %s handle %d: prio bands %d priomap 1 2 2 2 1 2 0 0 1 1 1 1 1 1 1 1", parentArg, parent+1, band)
	cmd := bpm.DefaultProcessBuilder("tc", strings.Split(args, " ")...).SetNetNS(c.nsPath).SetContext(c.ctx).Build()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return encodeOutputToError(output, err)
	}

	for index := 1; index <= 3; index++ {
		args := fmt.Sprintf("qdisc add dev eth0 parent %d:%d handle %d: sfq", parent+1, index, parent+1+index)
		cmd := bpm.DefaultProcessBuilder("tc", strings.Split(args, " ")...).SetNetNS(c.nsPath).SetContext(c.ctx).Build()
		output, err := cmd.CombinedOutput()
		if err != nil {
			return encodeOutputToError(output, err)
		}
	}

	return nil
}

func (c *tcClient) addNetem(parent string, handle string, netem *pb.Netem) error {
	log.Info("adding netem", "parent", parent, "handle", handle)

	args := fmt.Sprintf("qdisc add dev eth0 %s %s netem %s", parent, handle, convertNetemToArgs(netem))
	cmd := bpm.DefaultProcessBuilder("tc", strings.Split(args, " ")...).SetNetNS(c.nsPath).SetContext(c.ctx).Build()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return encodeOutputToError(output, err)
	}
	return nil
}

func (c *tcClient) addTbf(parent string, handle string, tbf *pb.Tbf) error {
	log.Info("adding tbf", "parent", parent, "handle", handle)

	args := fmt.Sprintf("qdisc add dev eth0 %s %s tbf %s", parent, handle, convertTbfToArgs(tbf))
	cmd := bpm.DefaultProcessBuilder("tc", strings.Split(args, " ")...).SetNetNS(c.nsPath).SetContext(c.ctx).Build()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return encodeOutputToError(output, err)
	}
	return nil
}

func (c *tcClient) addFilter(parent string, classid string, ipset string) error {
	log.Info("adding filter", "parent", parent, "classid", classid, "ipset", ipset)

	args := strings.Split(fmt.Sprintf("filter add dev eth0 %s basic match", parent), " ")
	args = append(args, fmt.Sprintf("ipset(%s dst)", ipset))
	args = append(args, strings.Split(classid, " ")...)
	cmd := bpm.DefaultProcessBuilder("tc", args...).SetNetNS(c.nsPath).SetContext(c.ctx).Build()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return encodeOutputToError(output, err)
	}
	return nil
}

func convertNetemToArgs(netem *pb.Netem) string {
	args := ""
	if netem.Time > 0 {
		args = fmt.Sprintf("delay %d", netem.Time)
		if netem.Jitter > 0 {
			args = fmt.Sprintf("%s %d", args, netem.Jitter)

			if netem.DelayCorr > 0 {
				args = fmt.Sprintf("%s %f", args, netem.DelayCorr)
			}
		}

		// reordering not possible without specifying some delay
		if netem.Reorder > 0 {
			args = fmt.Sprintf("%s reorder %f", args, netem.Reorder)
			if netem.ReorderCorr > 0 {
				args = fmt.Sprintf("%s %f", args, netem.ReorderCorr)
			}

			if netem.Gap > 0 {
				args = fmt.Sprintf("%s gap %d", args, netem.Gap)
			}
		}
	}

	if netem.Limit > 0 {
		args = fmt.Sprintf("%s limit %d", args, netem.Limit)
	}

	if netem.Loss > 0 {
		args = fmt.Sprintf("%s loss %f", args, netem.Loss)
		if netem.LossCorr > 0 {
			args = fmt.Sprintf("%s %f", args, netem.LossCorr)
		}
	}

	if netem.Duplicate > 0 {
		args = fmt.Sprintf("%s duplicate %f", args, netem.Duplicate)
		if netem.DuplicateCorr > 0 {
			args = fmt.Sprintf("%s %f", args, netem.DuplicateCorr)
		}
	}

	if netem.Corrupt > 0 {
		args = fmt.Sprintf("%s corrupt %f", args, netem.Corrupt)
		if netem.CorruptCorr > 0 {
			args = fmt.Sprintf("%s %f", args, netem.CorruptCorr)
		}
	}

	trimedArgs := []string{}

	for _, part := range strings.Split(args, " ") {
		if len(part) > 0 {
			trimedArgs = append(trimedArgs, part)
		}
	}

	return strings.Join(trimedArgs, " ")
}

func convertTbfToArgs(tbf *pb.Tbf) string {
	args := fmt.Sprintf("rate %d burst %d", tbf.Rate, tbf.Buffer)
	if tbf.Limit > 0 {
		args = fmt.Sprintf("%s limit %d", args, tbf.Limit)
	}
	if tbf.PeakRate > 0 {
		args = fmt.Sprintf("%s peakrate %d mtu %d", args, tbf.PeakRate, tbf.MinBurst)
	}

	return args
}
