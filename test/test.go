package test

import (
	"testing"
	"github.com/matyix/gke-test/cluster"
	"github.com/matyix/gke-test/client"
	"os"
	"log"
	"io/ioutil"
)

const (
	name = "test_cluster"
	projectID = "ringed-prism-189516"
	zone = "us-central1-a"
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

