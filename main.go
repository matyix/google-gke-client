package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	gke "google.golang.org/api/container/v1"
)

var (
	projectID = flag.String("project", "", "Project ID")
	zone      = flag.String("zone", "", "Compute zone")
)

const (
	cloudScope        = "https://www.googleapis.com/auth/cloud-platform"
	monitorWriteScope = "https://www.googleapis.com/auth/monitoring.write"
	storageReadScope  = "https://www.googleapis.com/auth/devstorage.read_only"
	statusRunning     = "RUNNING"
)

// Struct of GKE
type GKECluster struct {
	// ProjectID is the ID of your project to use when creating a cluster
	ProjectID string
	// The zone to launch the cluster
	Zone string
	// The IP address range of the container pods
	ClusterIpv4Cidr string
	// An optional description of this cluster
	Description string
	// The number of nodes to create in this cluster
	NodeCount int64
	// the kubernetes master version
	MasterVersion string
	// The authentication information for accessing the master
	MasterAuth *gke.MasterAuth
	// the kubernetes node version
	NodeVersion string
	// The name of this cluster
	Name string
	// Parameters used in creating the cluster's nodes
	NodeConfig *gke.NodeConfig
	// The path to the credential file(key.json)
	CredentialPath string
	// The content of the credential
	CredentialContent string
	// the temp file of the credential
	TempCredentialPath string
	// Enable alpha feature
	EnableAlphaFeature bool
	// Configuration for the HTTP (L7) load balancing controller addon
	HTTPLoadBalancing bool
	// Configuration for the horizontal pod autoscaling feature, which increases or decreases the number of replica pods a replication controller has based on the resource usage of the existing pods
	HorizontalPodAutoscaling bool
	// Configuration for the Kubernetes Dashboard
	KubernetesDashboard bool
	// Configuration for NetworkPolicy
	NetworkPolicyConfig bool
	// The list of Google Compute Engine locations in which the cluster's nodes should be located
	Locations []string
	// Network
	Network string
	// Sub Network
	SubNetwork string
	// Configuration for LegacyAbac
	LegacyAbac bool
	// NodePool id
	NodePoolID string
}

func main() {
	flag.Parse()

	if *projectID == "" {
		fmt.Fprintln(os.Stderr, "-project flag missing")
		flag.Usage()
		os.Exit(2)
	}
	if *zone == "" {
		fmt.Fprintln(os.Stderr, "-zone flag missing")
		flag.Usage()
		os.Exit(2)
	}

	svc, err := getServiceClient()
	if err != nil {
		log.Fatalf("Could not initialize gke client: %v", err)
	}

	if err := ListClusters(svc, *projectID, *zone); err != nil {
		log.Fatal(err)
	}

	cluster := GKECluster{}
	cluster.Name = "kluszterfirst"
	cluster.NodeCount = 2
	//cluster.CreateCluster(svc, *projectID, *zone)
	cluster.DeleteCluster(svc, *projectID, *zone)

}

func getServiceClient() (*gke.Service, error) {

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

func ListClusters(svc *gke.Service, projectID, zone string) error {
	list, err := svc.Projects.Zones.Clusters.List(projectID, zone).Do()
	if err != nil {
		return fmt.Errorf("failed to list clusters: %v", err)
	}
	for _, v := range list.Clusters {
		fmt.Printf("Cluster %q (%s) master_version: v%s", v.Name, v.Status, v.CurrentMasterVersion)

		poolList, err := svc.Projects.Zones.Clusters.NodePools.List(projectID, zone, v.Name).Do()
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

func (cc *GKECluster) CreateCluster(svc *gke.Service, projectID, zone string) error {

	cc.ProjectID = projectID
	cc.Zone = zone
	cc.validate()

	fmt.Printf("Cluster request: %v", cc.generateClusterCreateRequest())
	createCall, err := svc.Projects.Zones.Clusters.Create(cc.ProjectID, cc.Zone, cc.generateClusterCreateRequest()).Context(context.Background()).Do()

	fmt.Printf("Cluster request submitted: %v", cc.generateClusterCreateRequest())

	if err != nil && !strings.Contains(err.Error(), "alreadyExists") {
		log.Printf("Contains error", err)
		return err
	}
	if err == nil {
		log.Printf("Cluster %s create is called for project %s and zone %s. Status Code %v", cc.Name, cc.ProjectID, cc.Zone, createCall.HTTPStatusCode)
	}

	return cc.waitForCluster(svc)
}

func (cc *GKECluster) DeleteCluster(svc *gke.Service, projectID, zone string) error {

	cc.ProjectID = projectID
	cc.Zone = zone

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

func (cc *GKECluster) waitForCluster(svc *gke.Service) error {
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
			log.Printf("%v cluster %v......", strings.ToLower(cluster.Status), cc.Name)
			message = cluster.Status
		}
		time.Sleep(time.Second * 5)
	}
}

func (cc *GKECluster) validate() error {
	if cc.ProjectID == "" {
		return fmt.Errorf("project ID is required")
	} else if cc.Zone == "" {
		return fmt.Errorf("zone is required")
	} else if cc.Name == "" {
		return fmt.Errorf("cluster name is required")
	}
	return nil
}

func (cc *GKECluster) generateClusterCreateRequest() *gke.CreateClusterRequest {
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
