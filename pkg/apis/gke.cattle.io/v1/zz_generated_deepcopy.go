// +build !ignore_autogenerated

/*
Copyright 2019 Wrangler Sample Controller Authors

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

// Code generated by main. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CidrBlock) DeepCopyInto(out *CidrBlock) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CidrBlock.
func (in *CidrBlock) DeepCopy() *CidrBlock {
	if in == nil {
		return nil
	}
	out := new(CidrBlock)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterAddons) DeepCopyInto(out *ClusterAddons) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterAddons.
func (in *ClusterAddons) DeepCopy() *ClusterAddons {
	if in == nil {
		return nil
	}
	out := new(ClusterAddons)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GKEClusterConfig) DeepCopyInto(out *GKEClusterConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GKEClusterConfig.
func (in *GKEClusterConfig) DeepCopy() *GKEClusterConfig {
	if in == nil {
		return nil
	}
	out := new(GKEClusterConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GKEClusterConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GKEClusterConfigList) DeepCopyInto(out *GKEClusterConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GKEClusterConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GKEClusterConfigList.
func (in *GKEClusterConfigList) DeepCopy() *GKEClusterConfigList {
	if in == nil {
		return nil
	}
	out := new(GKEClusterConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GKEClusterConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GKEClusterConfigSpec) DeepCopyInto(out *GKEClusterConfigSpec) {
	*out = *in
	if in.EnableKubernetesAlpha != nil {
		in, out := &in.EnableKubernetesAlpha, &out.EnableKubernetesAlpha
		*out = new(bool)
		**out = **in
	}
	if in.ClusterAddons != nil {
		in, out := &in.ClusterAddons, &out.ClusterAddons
		*out = new(ClusterAddons)
		**out = **in
	}
	if in.ClusterIpv4CidrBlock != nil {
		in, out := &in.ClusterIpv4CidrBlock, &out.ClusterIpv4CidrBlock
		*out = new(string)
		**out = **in
	}
	if in.KubernetesVersion != nil {
		in, out := &in.KubernetesVersion, &out.KubernetesVersion
		*out = new(string)
		**out = **in
	}
	if in.LoggingService != nil {
		in, out := &in.LoggingService, &out.LoggingService
		*out = new(string)
		**out = **in
	}
	if in.MonitoringService != nil {
		in, out := &in.MonitoringService, &out.MonitoringService
		*out = new(string)
		**out = **in
	}
	if in.NodePools != nil {
		in, out := &in.NodePools, &out.NodePools
		*out = make([]NodePoolConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Network != nil {
		in, out := &in.Network, &out.Network
		*out = new(string)
		**out = **in
	}
	if in.Subnetwork != nil {
		in, out := &in.Subnetwork, &out.Subnetwork
		*out = new(string)
		**out = **in
	}
	if in.NetworkPolicyEnabled != nil {
		in, out := &in.NetworkPolicyEnabled, &out.NetworkPolicyEnabled
		*out = new(bool)
		**out = **in
	}
	if in.PrivateClusterConfig != nil {
		in, out := &in.PrivateClusterConfig, &out.PrivateClusterConfig
		*out = new(PrivateClusterConfig)
		**out = **in
	}
	if in.IPAllocationPolicy != nil {
		in, out := &in.IPAllocationPolicy, &out.IPAllocationPolicy
		*out = new(IPAllocationPolicy)
		**out = **in
	}
	if in.MasterAuthorizedNetworksConfig != nil {
		in, out := &in.MasterAuthorizedNetworksConfig, &out.MasterAuthorizedNetworksConfig
		*out = new(MasterAuthorizedNetworksConfig)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GKEClusterConfigSpec.
func (in *GKEClusterConfigSpec) DeepCopy() *GKEClusterConfigSpec {
	if in == nil {
		return nil
	}
	out := new(GKEClusterConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GKEClusterConfigStatus) DeepCopyInto(out *GKEClusterConfigStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GKEClusterConfigStatus.
func (in *GKEClusterConfigStatus) DeepCopy() *GKEClusterConfigStatus {
	if in == nil {
		return nil
	}
	out := new(GKEClusterConfigStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IPAllocationPolicy) DeepCopyInto(out *IPAllocationPolicy) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IPAllocationPolicy.
func (in *IPAllocationPolicy) DeepCopy() *IPAllocationPolicy {
	if in == nil {
		return nil
	}
	out := new(IPAllocationPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MasterAuthorizedNetworksConfig) DeepCopyInto(out *MasterAuthorizedNetworksConfig) {
	*out = *in
	if in.CidrBlocks != nil {
		in, out := &in.CidrBlocks, &out.CidrBlocks
		*out = make([]*CidrBlock, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(CidrBlock)
				**out = **in
			}
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MasterAuthorizedNetworksConfig.
func (in *MasterAuthorizedNetworksConfig) DeepCopy() *MasterAuthorizedNetworksConfig {
	if in == nil {
		return nil
	}
	out := new(MasterAuthorizedNetworksConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeConfig) DeepCopyInto(out *NodeConfig) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.OauthScopes != nil {
		in, out := &in.OauthScopes, &out.OauthScopes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Taints != nil {
		in, out := &in.Taints, &out.Taints
		*out = make([]NodeTaintConfig, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeConfig.
func (in *NodeConfig) DeepCopy() *NodeConfig {
	if in == nil {
		return nil
	}
	out := new(NodeConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodePoolAutoscaling) DeepCopyInto(out *NodePoolAutoscaling) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodePoolAutoscaling.
func (in *NodePoolAutoscaling) DeepCopy() *NodePoolAutoscaling {
	if in == nil {
		return nil
	}
	out := new(NodePoolAutoscaling)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodePoolConfig) DeepCopyInto(out *NodePoolConfig) {
	*out = *in
	if in.Autoscaling != nil {
		in, out := &in.Autoscaling, &out.Autoscaling
		*out = new(NodePoolAutoscaling)
		**out = **in
	}
	if in.Config != nil {
		in, out := &in.Config, &out.Config
		*out = new(NodeConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.InitialNodeCount != nil {
		in, out := &in.InitialNodeCount, &out.InitialNodeCount
		*out = new(int64)
		**out = **in
	}
	if in.MaxPodsConstraint != nil {
		in, out := &in.MaxPodsConstraint, &out.MaxPodsConstraint
		*out = new(int64)
		**out = **in
	}
	if in.Name != nil {
		in, out := &in.Name, &out.Name
		*out = new(string)
		**out = **in
	}
	if in.Version != nil {
		in, out := &in.Version, &out.Version
		*out = new(string)
		**out = **in
	}
	if in.Management != nil {
		in, out := &in.Management, &out.Management
		*out = new(NodePoolManagement)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodePoolConfig.
func (in *NodePoolConfig) DeepCopy() *NodePoolConfig {
	if in == nil {
		return nil
	}
	out := new(NodePoolConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodePoolManagement) DeepCopyInto(out *NodePoolManagement) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodePoolManagement.
func (in *NodePoolManagement) DeepCopy() *NodePoolManagement {
	if in == nil {
		return nil
	}
	out := new(NodePoolManagement)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeTaintConfig) DeepCopyInto(out *NodeTaintConfig) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeTaintConfig.
func (in *NodeTaintConfig) DeepCopy() *NodeTaintConfig {
	if in == nil {
		return nil
	}
	out := new(NodeTaintConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrivateClusterConfig) DeepCopyInto(out *PrivateClusterConfig) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrivateClusterConfig.
func (in *PrivateClusterConfig) DeepCopy() *PrivateClusterConfig {
	if in == nil {
		return nil
	}
	out := new(PrivateClusterConfig)
	in.DeepCopyInto(out)
	return out
}
