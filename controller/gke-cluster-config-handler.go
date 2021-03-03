package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/rancher/gke-operator/internal/gke"
	"github.com/rancher/gke-operator/internal/utils"
	gkev1 "github.com/rancher/gke-operator/pkg/apis/gke.cattle.io/v1"
	v12 "github.com/rancher/gke-operator/pkg/generated/controllers/gke.cattle.io/v1"
	wranglerv1 "github.com/rancher/wrangler/pkg/generated/controllers/core/v1"
	"github.com/sirupsen/logrus"
	gkeapi "google.golang.org/api/container/v1"
)

const (
	controllerName           = "gke-controller"
	controllerRemoveName     = "gke-controller-remove"
	gkeConfigCreatingPhase   = "creating"
	gkeConfigNotCreatedPhase = ""
	gkeConfigActivePhase     = "active"
	gkeConfigUpdatingPhase   = "updating"
	gkeConfigImportingPhase  = "importing"
	wait                     = 30
)

type Handler struct {
	gkeCC           v12.GKEClusterConfigClient
	gkeEnqueueAfter func(namespace, name string, duration time.Duration)
	gkeEnqueue      func(namespace, name string)
	secrets         wranglerv1.SecretClient
	secretsCache    wranglerv1.SecretCache
}

func Register(
	ctx context.Context,
	secrets wranglerv1.SecretController,
	gke v12.GKEClusterConfigController) {

	controller := &Handler{
		gkeCC:           gke,
		gkeEnqueue:      gke.Enqueue,
		gkeEnqueueAfter: gke.EnqueueAfter,
		secretsCache:    secrets.Cache(),
		secrets:         secrets,
	}

	// Register handlers
	gke.OnChange(ctx, controllerName, controller.recordError(controller.OnGkeConfigChanged))
	gke.OnRemove(ctx, controllerRemoveName, controller.OnGkeConfigRemoved)
}

func (h *Handler) OnGkeConfigChanged(key string, config *gkev1.GKEClusterConfig) (*gkev1.GKEClusterConfig, error) {
	if config == nil {
		return nil, nil
	}
	if config.DeletionTimestamp != nil {
		return nil, nil
	}

	switch config.Status.Phase {
	case gkeConfigImportingPhase:
		return h.importCluster(config)
	case gkeConfigNotCreatedPhase:
		return h.create(config)
	case gkeConfigCreatingPhase:
		return h.waitForCreationComplete(config)
	case gkeConfigActivePhase:
		return h.checkAndUpdate(config)
	case gkeConfigUpdatingPhase:
		return h.checkAndUpdate(config)
	}

	return config, nil
}

// recordError writes the error return by onChange to the failureMessage field on status. If there is no error, then
// empty string will be written to status
func (h *Handler) recordError(onChange func(key string, config *gkev1.GKEClusterConfig) (*gkev1.GKEClusterConfig, error)) func(key string, config *gkev1.GKEClusterConfig) (*gkev1.GKEClusterConfig, error) {
	return func(key string, config *gkev1.GKEClusterConfig) (*gkev1.GKEClusterConfig, error) {
		var err error
		var message string
		config, err = onChange(key, config)
		if config == nil {
			// GKE config is likely deleting
			return config, err
		}
		if config.Status.FailureMessage == message {
			return config, err
		}

		config = config.DeepCopy()

		if message != "" {
			if config.Status.Phase == gkeConfigActivePhase {
				// can assume an update is failing
				config.Status.Phase = gkeConfigUpdatingPhase
			}
		}
		config.Status.FailureMessage = message

		var recordErr error
		config, recordErr = h.gkeCC.UpdateStatus(config)
		if recordErr != nil {
			logrus.Errorf("Error recording gkecc [%s] failure message: %s", config.Name, recordErr.Error())
		}
		return config, err
	}
}

// importCluster cluster returns a spec containing the given config's displayName and region.
func (h *Handler) importCluster(config *gkev1.GKEClusterConfig) (*gkev1.GKEClusterConfig, error) {
	config.Status.Phase = gkeConfigActivePhase
	return h.gkeCC.UpdateStatus(config)
}

func (h *Handler) OnGkeConfigRemoved(key string, config *gkev1.GKEClusterConfig) (*gkev1.GKEClusterConfig, error) {
	if config.Spec.Imported {
		logrus.Infof("cluster [%s] is imported, will not delete GKE cluster", config.Name)
		return config, nil
	}
	if config.Status.Phase == gkeConfigNotCreatedPhase {
		// The most likely context here is that the cluster already existed in GKE, so we shouldn't delete it
		logrus.Warnf("cluster [%s] never advanced to creating status, will not delete GKE cluster", config.Name)
		return config, nil
	}

	logrus.Infof("deleting cluster [%s]", config.Name)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cred, err := h.getSecret(ctx, config)
	if err != nil {
		return config, err
	}
	svc, err := gke.GetGKEClient(ctx, cred)
	if err != nil {
		return config, err
	}

	logrus.Debugf("Removing cluster %v from project %v, region/zone %v", config.Spec.ClusterName, config.Spec.ProjectID, config.Spec.Region)
	operation, err := utils.WaitClusterRemoveExp(ctx, svc, config)
	if err != nil && !strings.Contains(err.Error(), "notFound") {
		return config, err
	} else if err == nil {
		logrus.Debugf("Cluster %v delete is called. Status Code %v", config.Spec.ClusterName, operation.HTTPStatusCode)
	} else {
		logrus.Debugf("Cluster %s doesn't exist", config.Spec.ClusterName)
	}

	return config, nil
}

func (h *Handler) create(config *gkev1.GKEClusterConfig) (*gkev1.GKEClusterConfig, error) {
	if config.Spec.Imported {
		config = config.DeepCopy()
		config.Status.Phase = gkeConfigImportingPhase
		return h.gkeCC.UpdateStatus(config)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cred, err := h.getSecret(ctx, config)
	if err != nil {
		return config, err
	}
	err = gke.Create(cred, config)
	if err != nil {
		return config, err
	}

	config = config.DeepCopy()
	config.Status.Phase = gkeConfigCreatingPhase
	return h.gkeCC.UpdateStatus(config)
}

func (h *Handler) checkAndUpdate(config *gkev1.GKEClusterConfig) (*gkev1.GKEClusterConfig, error) {
	if err := h.validateUpdate(config); err != nil {
		config = config.DeepCopy()
		config.Status.Phase = gkeConfigUpdatingPhase
		var updateErr error
		config, updateErr = h.gkeCC.UpdateStatus(config)
		if updateErr != nil {
			return config, updateErr
		}
		return config, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cred, err := h.getSecret(ctx, config)
	if err != nil {
		return config, err
	}
	svc, err := gke.GetGKEClient(ctx, cred)
	if err != nil {
		return config, err
	}

	cluster, err := svc.Projects.Locations.Clusters.Get(utils.ClusterRRN(config.Spec.ProjectID, config.Spec.Region, config.Spec.ClusterName)).Context(ctx).Do()
	if err != nil {
		return config, err
	}

	if cluster.Status == utils.ClusterStatusReconciling {
		// upstream cluster is already updating, must wait until sending next update
		logrus.Infof("waiting for cluster [%s] to finish updating", config.Name)
		if config.Status.Phase != gkeConfigUpdatingPhase {
			config = config.DeepCopy()
			config.Status.Phase = gkeConfigUpdatingPhase
			return h.gkeCC.UpdateStatus(config)
		}
		h.gkeEnqueueAfter(config.Namespace, config.Name, 30*time.Second)
		return config, nil
	}

	for _, np := range cluster.NodePools {

		if status := np.Status; status == utils.NodePoolStatusReconciling || status == utils.NodePoolStatusStopping ||
			status == utils.NodePoolStatusProvisioning {
			if config.Status.Phase != gkeConfigUpdatingPhase {
				config = config.DeepCopy()
				config.Status.Phase = gkeConfigUpdatingPhase
				config, err = h.gkeCC.UpdateStatus(config)
				if err != nil {
					return config, err
				}
			}
			logrus.Infof("waiting for cluster [%s] to update nodegroups [%s]", config.Name, np.Name)
			h.gkeEnqueueAfter(config.Namespace, config.Name, 30*time.Second)
			return config, nil
		}
	}

	upstreamSpec, err := buildUpstreamClusterState(cluster)
	if err != nil {
		return config, err
	}

	return h.updateUpstreamClusterState(config, upstreamSpec)
}

// enqueueUpdate enqueues the config if it is already in the updating phase. Otherwise, the
// phase is updated to "updating". This is important because the object needs to reenter the
// onChange handler to start waiting on the update.
func (h *Handler) enqueueUpdate(config *gkev1.GKEClusterConfig) (*gkev1.GKEClusterConfig, error) {
	if config.Status.Phase == gkeConfigUpdatingPhase {
		h.gkeEnqueue(config.Namespace, config.Name)
		return config, nil
	}
	config = config.DeepCopy()
	config.Status.Phase = gkeConfigUpdatingPhase
	return h.gkeCC.UpdateStatus(config)
}

func (h *Handler) updateUpstreamClusterState(config *gkev1.GKEClusterConfig, upstreamSpec *gkev1.GKEClusterConfigSpec) (*gkev1.GKEClusterConfig, error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cred, err := h.getSecret(ctx, config)
	if err != nil {
		return config, err
	}

	changed, err := gke.UpdateMasterKubernetesVersion(cred, config, upstreamSpec)
	if err != nil {
		return config, err
	}
	if changed == gke.Changed {
		return h.enqueueUpdate(config)
	}

	changed, err = gke.UpdateClusterAddons(cred, config, upstreamSpec)
	if err != nil {
		return config, err
	}
	if changed == gke.Retry {
		h.gkeEnqueueAfter(config.Namespace, config.Name, wait*time.Second)
		return config, nil
	}
	if changed == gke.Changed {
		return h.enqueueUpdate(config)
	}

	changed, err = gke.UpdateMasterAuthorizedNetworks(cred, config, upstreamSpec)
	if err != nil {
		return config, err
	}
	if changed == gke.Changed {
		return h.enqueueUpdate(config)
	}

	changed, err = gke.UpdateLoggingMonitoringService(cred, config, upstreamSpec)
	if err != nil {
		return config, err
	}
	if changed == gke.Changed {
		return h.enqueueUpdate(config)
	}

	changed, err = gke.UpdateNetworkPolicy(cred, config, upstreamSpec)
	if err != nil {
		return config, err
	}
	if changed == gke.Changed {
		return h.enqueueUpdate(config)
	}

	if config.Spec.NodePools != nil {
		upstreamNodePools := buildNodePoolMap(upstreamSpec.NodePools)
		for _, np := range config.Spec.NodePools {
			upstreamNodePool := upstreamNodePools[*np.Name]

			changed, err = gke.UpdateNodePoolKubernetesVersionOrImageType(cred, &np, config, upstreamNodePool)
			if err != nil {
				return config, err
			}
			if changed == gke.Changed {
				return h.enqueueUpdate(config)
			}

			changed, err = gke.UpdateNodePoolSize(cred, &np, config, upstreamNodePool)
			if err != nil {
				return config, err
			}
			if changed == gke.Changed {
				return h.enqueueUpdate(config)
			}

			changed, err = gke.UpdateNodePoolAutoscaling(cred, &np, config, upstreamNodePool)
			if err != nil {
				return config, err
			}
			if changed == gke.Changed {
				return h.enqueueUpdate(config)
			}
		}
	}

	// no new updates, set to active
	if config.Status.Phase != gkeConfigActivePhase {
		logrus.Infof("cluster [%s] finished updating", config.Name)
		config = config.DeepCopy()
		config.Status.Phase = gkeConfigActivePhase
		return h.gkeCC.UpdateStatus(config)
	}

	return config, nil
}

func (h *Handler) waitForCreationComplete(config *gkev1.GKEClusterConfig) (*gkev1.GKEClusterConfig, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cred, err := h.getSecret(ctx, config)
	if err != nil {
		return config, err
	}
	svc, err := gke.GetGKEClient(ctx, cred)
	if err != nil {
		return config, err
	}
	cluster, err := svc.Projects.Locations.Clusters.Get(utils.ClusterRRN(config.Spec.ProjectID, config.Spec.Region, config.Spec.ClusterName)).Context(ctx).Do()
	if err != nil {
		return config, err
	}
	if cluster.Status == utils.ClusterStatusError {
		return config, fmt.Errorf("creation failed for cluster %v", config.Spec.ClusterName)
	}
	if cluster.Status == utils.ClusterStatusRunning {
		logrus.Infof("Cluster %v is running", config.Spec.ClusterName)
		config = config.DeepCopy()
		config.Status.Phase = gkeConfigActivePhase
		return h.gkeCC.UpdateStatus(config)
	}
	logrus.Infof("waiting for cluster [%s] to finish creating", config.Name)
	h.gkeEnqueueAfter(config.Namespace, config.Name, wait*time.Second)

	return config, nil
}

func (h *Handler) validateUpdate(config *gkev1.GKEClusterConfig) error {

	var clusterVersion *semver.Version
	if config.Spec.KubernetesVersion != nil {
		var err error
		clusterVersion, err = semver.New(fmt.Sprintf("%s.0", utils.StringValue(config.Spec.KubernetesVersion)))
		if err != nil {
			return fmt.Errorf("improper version format for cluster [%s]: %s", config.Name, utils.StringValue(config.Spec.KubernetesVersion))
		}
	}

	var errors []string
	// validate nodegroup versions
	for _, np := range config.Spec.NodePools {
		if np.Version == nil {
			continue
		}
		version, err := semver.New(fmt.Sprintf("%s.0", utils.StringValue(np.Version)))
		if err != nil {
			errors = append(errors, fmt.Sprintf("improper version format for nodegroup [%s]: %s", utils.StringValue(np.Name), utils.StringValue(np.Version)))
			continue
		}
		if clusterVersion == nil {
			continue
		}
		if clusterVersion.EQ(*version) {
			continue
		}
		if clusterVersion.Minor-version.Minor == 1 {
			continue
		}
		errors = append(errors, fmt.Sprintf("versions for cluster [%s] and nodegroup [%s] not compatible: all nodegroup kubernetes versions"+
			"must be equal to or one minor version lower than the cluster kubernetes version", utils.StringValue(config.Spec.KubernetesVersion), utils.StringValue(np.Version)))
	}
	if len(errors) != 0 {
		return fmt.Errorf(strings.Join(errors, ";"))
	}
	return nil
}

func (h *Handler) getSecret(ctx context.Context, config *gkev1.GKEClusterConfig) (string, error) {
	ns, id := utils.ParseCredential(config.Spec.CredentialContent)
	secret, err := h.secretsCache.Get(ns, id)
	if err != nil {
		return "", err
	}
	dataBytes, ok := secret.Data["googlecredentialConfig-authEncodedJson"]
	if !ok {
		return "", fmt.Errorf("could not read malformed cloud credential secret %s from namespace %s", id, ns)
	}
	return string(dataBytes), nil
}

func buildNodePoolMap(nodePools []gkev1.NodePoolConfig) map[string]*gkev1.NodePoolConfig {
	ret := make(map[string]*gkev1.NodePoolConfig)
	for i := range nodePools {
		if nodePools[i].Name != nil {
			ret[*nodePools[i].Name] = &nodePools[i]
		}
	}
	return ret
}

func buildUpstreamClusterState(upstreamSpec *gkeapi.Cluster) (*gkev1.GKEClusterConfigSpec, error) {
	newSpec := &gkev1.GKEClusterConfigSpec{
		KubernetesVersion:    &upstreamSpec.CurrentMasterVersion,
		EnableAlphaFeature:   &upstreamSpec.EnableKubernetesAlpha,
		ClusterAddons:        &gkev1.ClusterAddons{},
		ClusterIpv4CidrBlock: upstreamSpec.ClusterIpv4Cidr,
		LoggingService:       &upstreamSpec.LoggingService,
		MonitoringService:    &upstreamSpec.MonitoringService,
		NetworkConfig:        &gkev1.NetworkConfig{},
		PrivateClusterConfig: &gkev1.PrivateClusterConfig{},
		IPAllocationPolicy:   &gkev1.IPAllocationPolicy{},
		MasterAuthorizedNetworksConfig: &gkev1.MasterAuthorizedNetworksConfig{
			Enabled: false,
		},
	}

	networkPolicyEnabled := false
	if upstreamSpec.NetworkPolicy != nil && upstreamSpec.NetworkPolicy.Enabled == true {
		networkPolicyEnabled = true
	}
	newSpec.NetworkPolicy = &networkPolicyEnabled

	if upstreamSpec.NetworkConfig != nil {
		newSpec.NetworkConfig.Network = &upstreamSpec.NetworkConfig.Network
		newSpec.NetworkConfig.Subnetwork = &upstreamSpec.NetworkConfig.Subnetwork
	} else {
		network := "default"
		newSpec.NetworkConfig.Network = &network
		newSpec.NetworkConfig.Subnetwork = &network
	}

	if upstreamSpec.PrivateClusterConfig != nil {
		newSpec.PrivateClusterConfig.EnablePrivateEndpoint = &upstreamSpec.PrivateClusterConfig.EnablePrivateNodes
		newSpec.PrivateClusterConfig.EnablePrivateNodes = &upstreamSpec.PrivateClusterConfig.EnablePrivateNodes
		newSpec.PrivateClusterConfig.MasterIpv4CidrBlock = upstreamSpec.PrivateClusterConfig.MasterIpv4CidrBlock
		newSpec.PrivateClusterConfig.PrivateEndpoint = upstreamSpec.PrivateClusterConfig.PrivateEndpoint
		newSpec.PrivateClusterConfig.PublicEndpoint = upstreamSpec.PrivateClusterConfig.PublicEndpoint
	} else {
		enabled := false
		newSpec.PrivateClusterConfig.EnablePrivateEndpoint = &enabled
		newSpec.PrivateClusterConfig.EnablePrivateNodes = &enabled
	}

	// build cluster addons
	if upstreamSpec.AddonsConfig != nil {
		lb := true
		if upstreamSpec.AddonsConfig.HttpLoadBalancing != nil {
			lb = !upstreamSpec.AddonsConfig.HttpLoadBalancing.Disabled
		}
		newSpec.ClusterAddons.HTTPLoadBalancing = lb
		hpa := true
		if upstreamSpec.AddonsConfig.HorizontalPodAutoscaling != nil {
			hpa = !upstreamSpec.AddonsConfig.HorizontalPodAutoscaling.Disabled
		}
		newSpec.ClusterAddons.HorizontalPodAutoscaling = hpa
		npc := true
		if upstreamSpec.AddonsConfig.NetworkPolicyConfig != nil {
			npc = !upstreamSpec.AddonsConfig.NetworkPolicyConfig.Disabled
		}
		newSpec.ClusterAddons.NetworkPolicyConfig = npc
	}

	if upstreamSpec.IpAllocationPolicy != nil {
		newSpec.IPAllocationPolicy.ClusterIpv4CidrBlock = upstreamSpec.IpAllocationPolicy.ClusterIpv4CidrBlock
		newSpec.IPAllocationPolicy.ClusterSecondaryRangeName = upstreamSpec.IpAllocationPolicy.ClusterSecondaryRangeName
		newSpec.IPAllocationPolicy.CreateSubnetwork = upstreamSpec.IpAllocationPolicy.CreateSubnetwork
		newSpec.IPAllocationPolicy.NodeIpv4CidrBlock = upstreamSpec.IpAllocationPolicy.NodeIpv4CidrBlock
		newSpec.IPAllocationPolicy.ServicesIpv4CidrBlock = upstreamSpec.IpAllocationPolicy.ServicesIpv4CidrBlock
		newSpec.IPAllocationPolicy.ServicesSecondaryRangeName = upstreamSpec.IpAllocationPolicy.ServicesSecondaryRangeName
		newSpec.IPAllocationPolicy.SubnetworkName = upstreamSpec.IpAllocationPolicy.SubnetworkName
		newSpec.IPAllocationPolicy.UseIPAliases = upstreamSpec.IpAllocationPolicy.UseIpAliases
	}

	if upstreamSpec.MasterAuthorizedNetworksConfig != nil {
		if upstreamSpec.MasterAuthorizedNetworksConfig.Enabled {
			newSpec.MasterAuthorizedNetworksConfig.Enabled = upstreamSpec.MasterAuthorizedNetworksConfig.Enabled
			for _, b := range upstreamSpec.MasterAuthorizedNetworksConfig.CidrBlocks {
				block := &gkev1.CidrBlock{
					CidrBlock:   b.CidrBlock,
					DisplayName: b.DisplayName,
				}
				newSpec.MasterAuthorizedNetworksConfig.CidrBlocks = append(newSpec.MasterAuthorizedNetworksConfig.CidrBlocks, block)
			}
		}
	}

	// build node groups
	newSpec.NodePools = make([]gkev1.NodePoolConfig, 0, len(upstreamSpec.NodePools))

	for _, np := range upstreamSpec.NodePools {
		if np.Status == utils.NodePoolStatusStopping {
			continue
		}

		newNP := gkev1.NodePoolConfig{
			Name:              &np.Name,
			Version:           &np.Version,
			InitialNodeCount:  &np.InitialNodeCount,
			MaxPodsConstraint: &np.MaxPodsConstraint.MaxPodsPerNode,
		}

		if np.Config != nil {
			newNP.Config = &gkev1.NodeConfig{
				DiskSizeGb:    np.Config.DiskSizeGb,
				DiskType:      np.Config.DiskType,
				ImageType:     np.Config.ImageType,
				Labels:        np.Config.Labels,
				LocalSsdCount: np.Config.LocalSsdCount,
				MachineType:   np.Config.MachineType,
				Preemptible:   np.Config.Preemptible,
			}

			newNP.Config.Taints = make([]gkev1.NodeTaintConfig, 0, len(np.Config.Taints))
			for _, t := range np.Config.Taints {
				newNP.Config.Taints = append(newNP.Config.Taints, gkev1.NodeTaintConfig{
					Effect: t.Effect,
					Key:    t.Key,
					Value:  t.Value,
				})
			}
		}

		if np.Autoscaling != nil {
			newNP.Autoscaling = &gkev1.NodePoolAutoscaling{
				Enabled:      np.Autoscaling.Enabled,
				MaxNodeCount: np.Autoscaling.MaxNodeCount,
				MinNodeCount: np.Autoscaling.MinNodeCount,
			}
		}

		newSpec.NodePools = append(newSpec.NodePools, newNP)
	}

	return newSpec, nil
}
