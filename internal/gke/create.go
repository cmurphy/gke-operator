package gke

import (
	"context"
	"fmt"

	"github.com/rancher/gke-operator/internal/utils"
	gkev1 "github.com/rancher/gke-operator/pkg/apis/gke.cattle.io/v1"
	"github.com/sirupsen/logrus"
	gkeapi "google.golang.org/api/container/v1"
)

// Errors
const (
	cannotBeNilError            = "field [%s] cannot be nil for non-import cluster [%s]"
	cannotBeNilForNodePoolError = "field [%s] cannot be nil for nodepool [%s] in non-nil cluster [%s]"
)

func Create(credential string, config *gkev1.GKEClusterConfig) error {
	err := validateCreateRequest(credential, config)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc, err := GetGKEClient(ctx, credential)
	if err != nil {
		return err
	}

	createClusterRequest, err := newClusterCreateRequest(config)
	if err != nil {
		return err
	}

	operation, err := svc.Projects.
		Locations.
		Clusters.
		Create(
			utils.LocationRRN(config.Spec.ProjectID, config.Spec.Region),
			createClusterRequest).
		Context(ctx).
		Do()

	logrus.Debugf("Cluster %s create is called for project %s and region/zone %s. Status Code %v",
		config.Spec.ClusterName, config.Spec.ProjectID, config.Spec.Region, operation.HTTPStatusCode)

	return nil
}

// newClusterCreateRequest creates a CreateClusterRequest that can be submitted to GKE
func newClusterCreateRequest(config *gkev1.GKEClusterConfig) (*gkeapi.CreateClusterRequest, error) {

	enableAlphaFeatures := config.Spec.EnableAlphaFeature != nil && *config.Spec.EnableAlphaFeature
	request := &gkeapi.CreateClusterRequest{
		Cluster: &gkeapi.Cluster{
			Name:                  config.Spec.ClusterName,
			Description:           config.Spec.Description,
			InitialClusterVersion: *config.Spec.KubernetesVersion,
			EnableKubernetesAlpha: enableAlphaFeatures,
			LoggingService:        *config.Spec.LoggingService,
			MonitoringService:     *config.Spec.MonitoringService,
			IpAllocationPolicy: &gkeapi.IPAllocationPolicy{
				ClusterIpv4CidrBlock:       config.Spec.IPAllocationPolicy.ClusterIpv4CidrBlock,
				ClusterSecondaryRangeName:  config.Spec.IPAllocationPolicy.ClusterSecondaryRangeName,
				CreateSubnetwork:           config.Spec.IPAllocationPolicy.CreateSubnetwork,
				NodeIpv4CidrBlock:          config.Spec.IPAllocationPolicy.NodeIpv4CidrBlock,
				ServicesIpv4CidrBlock:      config.Spec.IPAllocationPolicy.ServicesIpv4CidrBlock,
				ServicesSecondaryRangeName: config.Spec.IPAllocationPolicy.ServicesSecondaryRangeName,
				SubnetworkName:             config.Spec.IPAllocationPolicy.SubnetworkName,
				UseIpAliases:               config.Spec.IPAllocationPolicy.UseIPAliases,
			},
			AddonsConfig: &gkeapi.AddonsConfig{},
			NodePools:    []*gkeapi.NodePool{},
		},
	}

	addons := config.Spec.ClusterAddons
	if addons != nil {
		request.Cluster.AddonsConfig.HttpLoadBalancing = &gkeapi.HttpLoadBalancing{Disabled: !addons.HTTPLoadBalancing}
		request.Cluster.AddonsConfig.HorizontalPodAutoscaling = &gkeapi.HorizontalPodAutoscaling{Disabled: !addons.HorizontalPodAutoscaling}
		request.Cluster.AddonsConfig.NetworkPolicyConfig = &gkeapi.NetworkPolicyConfig{Disabled: !addons.NetworkPolicyConfig}
	}

	request.Cluster.NodePools = make([]*gkeapi.NodePool, 0, len(config.Spec.NodePools))

	for _, np := range config.Spec.NodePools {

		var taints []*gkeapi.NodeTaint = make([]*gkeapi.NodeTaint, 0, len(np.Config.Taints))
		for _, t := range np.Config.Taints {
			taints = append(taints, &gkeapi.NodeTaint{
				Effect: t.Effect,
				Key:    t.Key,
				Value:  t.Value,
			})
		}

		nodePool := &gkeapi.NodePool{
			Name: *np.Name,
			Autoscaling: &gkeapi.NodePoolAutoscaling{
				Enabled:      np.Autoscaling.Enabled,
				MaxNodeCount: np.Autoscaling.MaxNodeCount,
				MinNodeCount: np.Autoscaling.MinNodeCount,
			},
			InitialNodeCount: *np.InitialNodeCount,
			Config: &gkeapi.NodeConfig{
				DiskSizeGb:  np.Config.DiskSizeGb,
				DiskType:    np.Config.DiskType,
				ImageType:   np.Config.ImageType,
				Labels:      np.Config.Labels,
				MachineType: np.Config.MachineType,
				OauthScopes: np.Config.OauthScopes,
				Taints:      taints,
				Preemptible: np.Config.Preemptible,
			},
		}
		// If nil, use default
		if np.MaxPodsConstraint != nil {
			nodePool.MaxPodsConstraint = &gkeapi.MaxPodsConstraint{
				MaxPodsPerNode: *np.MaxPodsConstraint,
			}
		}

		request.Cluster.NodePools = append(request.Cluster.NodePools, nodePool)
	}

	if config.Spec.MasterAuthorizedNetworksConfig != nil {
		var blocks = make([]*gkeapi.CidrBlock, len(config.Spec.MasterAuthorizedNetworksConfig.CidrBlocks))
		for _, b := range config.Spec.MasterAuthorizedNetworksConfig.CidrBlocks {
			blocks = append(blocks, &gkeapi.CidrBlock{
				CidrBlock:   b.CidrBlock,
				DisplayName: b.DisplayName,
			})
		}
		request.Cluster.MasterAuthorizedNetworksConfig = &gkeapi.MasterAuthorizedNetworksConfig{
			Enabled:    config.Spec.MasterAuthorizedNetworksConfig.Enabled,
			CidrBlocks: blocks,
		}
	}

	if config.Spec.NetworkConfig != nil {
		request.Cluster.NetworkConfig = &gkeapi.NetworkConfig{
			Subnetwork: *config.Spec.NetworkConfig.Subnetwork,
			Network:    *config.Spec.NetworkConfig.Network,
		}
	}

	if config.Spec.NetworkPolicy != nil {
		request.Cluster.NetworkPolicy = &gkeapi.NetworkPolicy{
			Enabled: *config.Spec.NetworkPolicy,
		}
	}

	if config.Spec.PrivateClusterConfig != nil {
		request.Cluster.PrivateClusterConfig = &gkeapi.PrivateClusterConfig{
			EnablePrivateEndpoint: *config.Spec.PrivateClusterConfig.EnablePrivateEndpoint,
			EnablePrivateNodes:    *config.Spec.PrivateClusterConfig.EnablePrivateNodes,
			MasterIpv4CidrBlock:   config.Spec.PrivateClusterConfig.MasterIpv4CidrBlock,
			PrivateEndpoint:       config.Spec.PrivateClusterConfig.PrivateEndpoint,
			PublicEndpoint:        config.Spec.PrivateClusterConfig.PublicEndpoint,
		}
	}

	return request, nil
}

// validateCreateRequest checks a config for the ability to generate a create request
func validateCreateRequest(cred string, config *gkev1.GKEClusterConfig) error {
	if config.Spec.ProjectID == "" {
		return fmt.Errorf("project ID is required")
	}
	if config.Spec.Zone == "" && config.Spec.Region == "" {
		return fmt.Errorf("zone or region is required")
	}
	if config.Spec.Zone != "" && config.Spec.Region != "" {
		return fmt.Errorf("only one of zone or region must be specified")
	}
	if config.Spec.ClusterName == "" {
		return fmt.Errorf("cluster name is required")
	}

	for _, np := range config.Spec.NodePools {
		if np.Autoscaling != nil && np.Autoscaling.Enabled {
			if np.Autoscaling.MinNodeCount < 1 || np.Autoscaling.MaxNodeCount < np.Autoscaling.MinNodeCount {
				return fmt.Errorf("minNodeCount in the NodePool must be >= 1 and <= maxNodeCount")
			}
		}
	}

	//check if cluster with same name exists
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc, err := GetGKEClient(ctx, cred)
	if err != nil {
		return err
	}
	operation, err := svc.Projects.Locations.Clusters.List(
		utils.LocationRRN(config.Spec.ProjectID, config.Spec.Region)).Context(ctx).Do()

	for _, cluster := range operation.Clusters {
		if cluster.Name == config.Spec.ClusterName {
			return fmt.Errorf("cannot create cluster [%s] because a cluster in GKE exists with the same name", config.Spec.ClusterName)
		}
	}

	if config.Spec.Imported {
		// Validation from here on out is for nilable attributes, not required for imported clusters
		return nil
	}

	if config.Spec.EnableAlphaFeature == nil {
		return fmt.Errorf(cannotBeNilError, "enableAlphaFeature", config.Name)
	}
	if config.Spec.KubernetesVersion == nil {
		return fmt.Errorf(cannotBeNilError, "kubernetesVersion", config.Name)
	}
	if config.Spec.ClusterAddons == nil {
		return fmt.Errorf(cannotBeNilError, "clusterAddons", config.Name)
	}
	if config.Spec.IPAllocationPolicy == nil {
		return fmt.Errorf(cannotBeNilError, "ipAllocationPolicy", config.Name)
	}
	if config.Spec.LoggingService == nil {
		return fmt.Errorf(cannotBeNilError, "loggingService", config.Name)
	}
	if config.Spec.NetworkConfig == nil {
		return fmt.Errorf(cannotBeNilError, "networkConfig", config.Name)
	}
	if config.Spec.NetworkPolicy == nil {
		return fmt.Errorf(cannotBeNilError, "networkPolicy", config.Name)
	}
	if config.Spec.PrivateClusterConfig == nil {
		return fmt.Errorf(cannotBeNilError, "privateClusterConfig", config.Name)
	}
	if config.Spec.MasterAuthorizedNetworksConfig == nil {
		return fmt.Errorf(cannotBeNilError, "masterAuthorizedNetworksConfig", config.Name)
	}
	if config.Spec.MonitoringService == nil {
		return fmt.Errorf(cannotBeNilError, "monitoringService", config.Name)
	}

	for _, np := range config.Spec.NodePools {
		if np.Name == nil {
			return fmt.Errorf(cannotBeNilError, "nodePool.name", config.Name)
		}
		cannotBeNil := cannotBeNilForNodePoolError
		if np.Version == nil {
			return fmt.Errorf(cannotBeNil, "version", *np.Name, config.Name)
		}
		if np.Autoscaling == nil {
			return fmt.Errorf(cannotBeNil, "autoscaling", *np.Name, config.Name)
		}
		if np.InitialNodeCount == nil {
			return fmt.Errorf(cannotBeNil, "initialNodeCount", *np.Name, config.Name)
		}
		if np.MaxPodsConstraint == nil {
			return fmt.Errorf(cannotBeNil, "maxPodsConstraint", *np.Name, config.Name)
		}
		if np.Config == nil {
			return fmt.Errorf(cannotBeNil, "config", *np.Name, config.Name)
		}
	}

	return nil
}
