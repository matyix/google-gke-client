package client

import (
	"flag"
	"fmt"
	"github.com/banzaicloud/google-gke-client/cluster"
	"github.com/banzaicloud/google-gke-client/utils"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	gke "google.golang.org/api/container/v1"
	"log"
	"os"
	"strings"
)

var (
	projectID = flag.String("project", "", "Project ID")
	zone      = flag.String("zone", "", "Compute zone")
	ops       = flag.String("ops", "", "Operation not specified - should be c -create, d -delete, u -update. Default is list")
)

func GetServiceClient() (*gke.Service, error) {

	// See https://cloud.google.com/docs/authentication/.
	// Use GOOGLE_APPLICATION_CREDENTIALS environment variable to specify
	// a service account key file to authenticate to the API.

	client, err := google.DefaultClient(context.Background(), gke.CloudPlatformScope)
	if err != nil {
		log.Fatalf("Could not get authenticated client: %v", err)
		return nil, err
	}
	service, err := gke.New(client)
	if err != nil {
		log.Fatalf("Could not initialize gke client: %v", err)
		return nil, err
	}
	log.Printf("Using service acc: %v", service)
	return service, nil
}

func ListClusters(svc *gke.Service, cc cluster.GKECluster) error {
	list, err := svc.Projects.Zones.Clusters.List(cc.ProjectID, cc.Zone).Do()
	if err != nil {
		return fmt.Errorf("failed to list clusters: %v", err)
	}
	for _, v := range list.Clusters {
		fmt.Printf("Cluster %q (%s) master_version: v%s", v.Name, v.Status, v.CurrentMasterVersion)

		poolList, err := svc.Projects.Zones.Clusters.NodePools.List(cc.ProjectID, cc.Zone, v.Name).Do()
		if err != nil {
			return fmt.Errorf("failed to list node pools for cluster %q: %v", v.Name, err)
		}
		for _, np := range poolList.NodePools {
			fmt.Printf("  -> Pool %q (%s) machineType=%s node_version=v%s autoscaling=%v", np.Name, np.Status,
				np.Config.MachineType, np.Version, np.Autoscaling != nil && np.Autoscaling.Enabled)
		}
	}
	return nil
}

func CreateCluster(svc *gke.Service, cc cluster.GKECluster) error {

	//cc.validate()

	log.Printf("Cluster request: %v", utils.GenerateClusterCreateRequest(cc))
	createCall, err := svc.Projects.Zones.Clusters.Create(cc.ProjectID, cc.Zone, utils.GenerateClusterCreateRequest(cc)).Context(context.Background()).Do()

	fmt.Printf("Cluster request submitted: %v", utils.GenerateClusterCreateRequest(cc))

	if err != nil && !strings.Contains(err.Error(), "alreadyExists") {
		log.Printf("Contains error", err)
		return err
	}
	if err == nil {
		log.Printf("Cluster %s create is called for project %s and zone %s. Status Code %v", cc.Name, cc.ProjectID, cc.Zone, createCall.HTTPStatusCode)
	}

	return utils.WaitForCluster(svc, cc)
}

func UpdateCluster(svc *gke.Service, cc cluster.GKECluster) error {

	log.Printf("Updating cluster. MasterVersion: %s, NodeVersion: %s, NodeCount: %v", cc.MasterVersion, cc.NodeVersion, cc.NodeCount)
	if cc.NodePoolID == "" {
		cluster, err := svc.Projects.Zones.Clusters.Get(cc.ProjectID, cc.Zone, cc.Name).Context(context.Background()).Do()
		if err != nil {
			log.Printf("Contains error", err)
			return err
		}
		cc.NodePoolID = cluster.NodePools[0].Name
	}

	if cc.MasterVersion != "" {
		log.Printf("Updating master to %v version", cc.MasterVersion)
		updateCall, err := svc.Projects.Zones.Clusters.Update(cc.ProjectID, cc.Zone, cc.Name, &gke.UpdateClusterRequest{
			Update: &gke.ClusterUpdate{
				DesiredMasterVersion: cc.MasterVersion,
			},
		}).Context(context.Background()).Do()
		if err != nil {
			return err
		}
		log.Printf("Cluster %s update is called for project %s and zone %s. Status Code %v", cc.Name, cc.ProjectID, cc.Zone, updateCall.HTTPStatusCode)
		if err := utils.WaitForCluster(svc, cc); err != nil {
			log.Printf("Contains error", err)
			return err
		}
	}

	if cc.NodeVersion != "" {
		log.Printf("Updating node to %v verison", cc.NodeVersion)
		updateCall, err := svc.Projects.Zones.Clusters.NodePools.Update(cc.ProjectID, cc.Zone, cc.Name, cc.NodePoolID, &gke.UpdateNodePoolRequest{
			NodeVersion: cc.NodeVersion,
		}).Context(context.Background()).Do()
		if err != nil {
			log.Printf("Contains error", err)
			return err
		}
		log.Printf("Nodepool %s update is called for project %s, zone %s and cluster %s. Status Code %v", cc.NodePoolID, cc.ProjectID, cc.Zone, cc.Name, updateCall.HTTPStatusCode)
		if err := utils.WaitForNodePool(svc, cc); err != nil {
			log.Printf("Contains error", err)
			return err
		}
	}

	if cc.NodeCount != 0 {
		log.Printf("Updating node size to %v", cc.NodeCount)
		updateCall, err := svc.Projects.Zones.Clusters.NodePools.SetSize(cc.ProjectID, cc.Zone, cc.Name, cc.NodePoolID, &gke.SetNodePoolSizeRequest{
			NodeCount: cc.NodeCount,
		}).Context(context.Background()).Do()
		if err != nil {
			return err
		}
		log.Printf("Nodepool %s size change is called for project %s, zone %s and cluster %s. Status Code %v", cc.NodePoolID, cc.ProjectID, cc.Zone, cc.Name, updateCall.HTTPStatusCode)
		if err := utils.WaitForCluster(svc, cc); err != nil {
			log.Printf("Contains error", err)
			return err
		}
	}
	return nil
}

func DeleteCluster(svc *gke.Service, cc cluster.GKECluster) error {

	log.Printf("Removing cluster %v from project %v, zone %v", cc.Name, cc.ProjectID, cc.Zone)
	deleteCall, err := svc.Projects.Zones.Clusters.Delete(cc.ProjectID, cc.Zone, cc.Name).Context(context.Background()).Do()
	if err != nil && !strings.Contains(err.Error(), "notFound") {
		return err
	} else if err == nil {
		log.Printf("Cluster %v delete is called. Status Code %v", cc.Name, deleteCall.HTTPStatusCode)
	} else {
		log.Printf("Cluster %s doesn't exist", cc.Name)
	}
	os.RemoveAll(cc.TempCredentialPath)
	return nil
}
