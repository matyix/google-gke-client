package test

import (
	"fmt"
	"github.com/banzaicloud/google-gke-client/client"
	"github.com/banzaicloud/google-gke-client/cluster"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

const (
	name      = "test_cluster"
	projectID = "ringed-prism-189516"
	zone      = "us-central1-a"
)

func TestCreateCluster(t *testing.T) {

	credentialPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	log.Printf(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	data, err := ioutil.ReadFile(credentialPath)
	if err != nil {
		log.Fatal("GOOGLE_APPLICATION_CREDENTIALS env var is not specified", err)
	}

	svc, err := client.GetServiceClient()
	if err != nil {
		t.Errorf("Could not get service client", err)
	}

	cluster := cluster.GKECluster{
		ProjectID:         projectID,
		Zone:              zone,
		Name:              "kluszterfirst",
		NodeCount:         1,
		CredentialPath:    credentialPath,
		CredentialContent: string(data),
	}

	err = client.CreateCluster(svc, cluster)
	if err != nil {
		t.Errorf("Cluster create failed", err)
	}

}

func TestListClusters(t *testing.T) {

	svc, err := client.GetServiceClient()
	if err != nil {
		t.Errorf("Could not get service client", err)
	}
	list, err := svc.Projects.Zones.Clusters.List(projectID, zone).Do()
	if err != nil {
		t.Errorf("failed to list clusters: %v", err)
	}
	for _, v := range list.Clusters {
		fmt.Printf("Cluster %q (%s) master_version: v%s", v.Name, v.Status, v.CurrentMasterVersion)

		poolList, err := svc.Projects.Zones.Clusters.NodePools.List(projectID, zone, v.Name).Do()
		if err != nil {
			t.Errorf("failed to list node pools for cluster %q: %v", v.Name, err)
		}
		for _, np := range poolList.NodePools {
			fmt.Printf("  -> Pool %q (%s) machineType=%s node_version=v%s autoscaling=%v", np.Name, np.Status,
				np.Config.MachineType, np.Version, np.Autoscaling != nil && np.Autoscaling.Enabled)
		}
	}
}
