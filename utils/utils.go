package utils

import (
	"fmt"
	"github.com/banzaicloud/google-gke-client/cluster"
	"golang.org/x/net/context"
	gke "google.golang.org/api/container/v1"
	"log"
	"time"
)

const (
	cloudScope        = "https://www.googleapis.com/auth/cloud-platform"
	monitorWriteScope = "https://www.googleapis.com/auth/monitoring.write"
	storageReadScope  = "https://www.googleapis.com/auth/devstorage.read_only"
	statusRunning     = "RUNNING"
)

func WaitForCluster(svc *gke.Service, cc cluster.GKECluster) error {
	message := ""
	for {
		cluster, err := svc.Projects.Zones.Clusters.Get(cc.ProjectID, cc.Zone, cc.Name).Context(context.TODO()).Do()
		if err != nil {
			return err
		}
		if cluster.Status == statusRunning {
			log.Printf("Cluster %v is running", cc.Name)
			return nil
		}
		if cluster.Status != message {
			log.Printf("%v cluster %v", string(cluster.Status), cc.Name)
			message = cluster.Status
		}
		time.Sleep(time.Second * 5)
	}
}

func WaitForNodePool(svc *gke.Service, cc cluster.GKECluster) error {
	message := ""
	for {
		nodepool, err := svc.Projects.Zones.Clusters.NodePools.Get(cc.ProjectID, cc.Zone, cc.Name, cc.NodePoolID).Context(context.TODO()).Do()
		if err != nil {
			return err
		}
		if nodepool.Status == statusRunning {
			log.Printf("Nodepool %v is running", cc.Name)
			return nil
		}
		if nodepool.Status != message {
			log.Printf("%v nodepool %v", string(nodepool.Status), cc.NodePoolID)
			message = nodepool.Status
		}
		time.Sleep(time.Second * 5)
	}
}

func Validate(cc *cluster.GKECluster) error {
	if cc.ProjectID == "" {
		return fmt.Errorf("project ID is required")
	} else if cc.Zone == "" {
		return fmt.Errorf("zone is required")
	} else if cc.Name == "" {
		return fmt.Errorf("cluster name is required")
	}
	return nil
}

func GenerateClusterCreateRequest(cc cluster.GKECluster) *gke.CreateClusterRequest {
	request := gke.CreateClusterRequest{
		Cluster: &gke.Cluster{},
	}
	request.Cluster.Name = cc.Name
	request.Cluster.Zone = cc.Zone
	request.Cluster.InitialClusterVersion = cc.MasterVersion
	request.Cluster.InitialNodeCount = cc.NodeCount
	request.Cluster.ClusterIpv4Cidr = cc.ClusterIpv4Cidr
	request.Cluster.Description = cc.Description
	request.Cluster.EnableKubernetesAlpha = cc.EnableAlphaFeature
	request.Cluster.AddonsConfig = &gke.AddonsConfig{
		HttpLoadBalancing:        &gke.HttpLoadBalancing{Disabled: !cc.HTTPLoadBalancing},
		HorizontalPodAutoscaling: &gke.HorizontalPodAutoscaling{Disabled: !cc.HorizontalPodAutoscaling},
		KubernetesDashboard:      &gke.KubernetesDashboard{Disabled: !cc.KubernetesDashboard},
		NetworkPolicyConfig:      &gke.NetworkPolicyConfig{Disabled: !cc.NetworkPolicyConfig},
	}
	request.Cluster.Network = cc.Network
	request.Cluster.Subnetwork = cc.SubNetwork
	request.Cluster.LegacyAbac = &gke.LegacyAbac{
		Enabled: cc.LegacyAbac,
	}
	request.Cluster.MasterAuth = &gke.MasterAuth{
		Username: "admin",
	}
	request.Cluster.NodeConfig = cc.NodeConfig
	return &request
}
