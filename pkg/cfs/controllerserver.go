/*
Copyright 2017 The Kubernetes Authors.

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

package cfs

import (
	"fmt"
	"time"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/golang/glog"
	"github.com/tiglabs/containerfs/proto/vp"
	"github.com/tiglabs/containerfs/utils"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
	"golang.org/x/net/context"
	"k8s.io/kubernetes/pkg/volume/util"
)

var defaultVolMgr1,defaultVolMgr2,defaultVolMgr3 string

type controllerServer struct {
	*csicommon.DefaultControllerServer
}

func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {

	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		glog.V(3).Infof("invalid create volume req: %v", req)
		return nil, err
	}

	// Volume Size - Default is 1 GiB
	volSizeBytes := int64(1 * 1024 * 1024 * 1024)
	if req.GetCapacityRange() != nil {
		volSizeBytes = int64(req.GetCapacityRange().GetRequiredBytes())
	}
	volSizeGB := int32(util.RoundUpSize(volSizeBytes, 1024*1024*1024))

	//Volume Name
	volName := req.GetName()

	// Volume Parameters
	var VolMgrHosts []string
	VolMgrHosts = make([]string, 3)
	VolMgrHosts[0] = req.GetParameters()["cfsvolmgr1"]
	VolMgrHosts[1] = req.GetParameters()["cfsvolmgr2"]
	VolMgrHosts[2] = req.GetParameters()["cfsvolmgr3"]

	defaultVolMgr1 = VolMgrHosts[0]
	defaultVolMgr2 = VolMgrHosts[1]
	defaultVolMgr3 = VolMgrHosts[2]

	_, conn, err := utils.DialVolMgr(VolMgrHosts)
	if err != nil {
		glog.Errorf("CreateVol failed,Dial to VolMgrHosts:%v fail :%v", VolMgrHosts, err)
		return nil, err
	}
	defer conn.Close()
	vc := vp.NewVolMgrClient(conn)
	pCreateVolReq := &vp.CreateVolReq{
		VolName:    volName,
		SpaceQuota: volSizeGB,
		Tier:       "sas",
		Copies:     "3",
	}
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	ack, err := vc.CreateVol(ctx, pCreateVolReq)
	if err != nil  {
		glog.Errorf("Create CFS Vol failed, VolMgr Leader return failed,  err:%v", err)
		return nil, err
	}
	if ack.Ret != 0 {
		glog.Errorf("Create CFS Vol failed, VolMgr Leader return failed, ret:%v", ack.Ret)
		return nil, fmt.Errorf("failed to create CFS Vol failed, return not equal 0")
	}

	glog.V(1).Infof("Succesfully created CFS volume with size: %d and name: %s and uuid:%v", volSizeGB, volName, ack.UUID)

	cfsvolmgr1 := strings.Split(VolMgrHosts[0],":")[0]
	cfsvolmgr2 := strings.Split(VolMgrHosts[1],":")[0]	
	cfsvolmgr3 := strings.Split(VolMgrHosts[2],":")[0]

	resp := &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			Id: ack.UUID,
			Attributes: map[string]string{
				"cfsvolname":	volName,
				"cfsvol":	ack.UUID,
				"cfsvolmgr1":	cfsvolmgr1,
				"cfsvolmgr2":   cfsvolmgr2,
				"cfsvolmgr3":   cfsvolmgr3,
			},
		},
	}
	return resp, nil

}

func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		glog.V(3).Infof("invalid delete volume req: %v", req)
		return nil, err
	}
	volumeId := req.VolumeId
	glog.V(4).Infof("deleting volume %s: req: %v", volumeId, req)

	var VolMgrHosts []string
	VolMgrHosts = make([]string, 3)
	VolMgrHosts[0] = defaultVolMgr1
        VolMgrHosts[1] = defaultVolMgr2
        VolMgrHosts[2] = defaultVolMgr3

	_, conn, err := utils.DialVolMgr(VolMgrHosts)
	if err != nil {
		glog.Errorf("failed to create cfs rest client")
		return nil, fmt.Errorf("failed to create cfs REST client")
	}
	defer conn.Close()
	vc := vp.NewVolMgrClient(conn)

	pDeleteVolReq := &vp.DeleteVolReq{
		UUID: volumeId,
	}
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	pDeleteVolAck, err := vc.DeleteVol(ctx, pDeleteVolReq)
	if err != nil {
		glog.Errorf("Delete cfs volume:%v error:%v", volumeId, err)
		return nil, err
	}

	if pDeleteVolAck.Ret != 0 {
		glog.Errorf("Delete cfs volume:%v failed :%v", volumeId, pDeleteVolAck.Ret)
		return nil, fmt.Errorf("Delete cfs volume:%v failed :%v", volumeId, pDeleteVolAck.Ret)
	}

	glog.V(1).Infof("cfs volume :%s deleted success", volumeId)
	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *controllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	for _, cap := range req.VolumeCapabilities {
		if cap.GetAccessMode().GetMode() != csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER {
			return &csi.ValidateVolumeCapabilitiesResponse{Supported: false, Message: ""}, nil
		}
	}
	return &csi.ValidateVolumeCapabilitiesResponse{Supported: true, Message: ""}, nil
}
