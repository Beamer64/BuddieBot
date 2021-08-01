package gcp

import (
	"context"
	"fmt"
	"os"

	compute "cloud.google.com/go/compute/apiv1"
	"github.com/pkg/errors"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

type GCPClient struct {
	project string
	zone    string
}

func NewGCPClient(gcpConfigPath, project, zone string) (*GCPClient, error) {
	if _, err := os.Stat(gcpConfigPath); os.IsNotExist(err) {
		return nil, err
	}

	err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", gcpConfigPath)
	if err != nil {
		return nil, err
	}
	return &GCPClient{
		project: project,
		zone:    zone,
	}, nil
}

func (gc *GCPClient) StopMachine(machine string) error {
	ctx := context.Background()
	c, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return err
	}
	defer func(c *compute.InstancesClient) {
		err := c.Close()
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
		}
	}(c)

	req := &computepb.StopInstanceRequest{
		Instance: machine,
		Zone:     gc.zone,
		Project:  gc.project,
	}
	fmt.Printf("stop machine request: %+v\n", req)
	resp, err := c.Stop(ctx, req)
	if err != nil {
		return err
	}

	fmt.Printf("%+v", resp)
	return nil
}

func (gc *GCPClient) StartMachine(machine string) error {
	ctx := context.Background()
	c, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return err
	}
	defer func(c *compute.InstancesClient) {
		err := c.Close()
		if err != nil {
			fmt.Printf("%+v", errors.WithStack(err))
		}
	}(c)

	req := &computepb.StartInstanceRequest{
		Instance: machine,
		Zone:     gc.zone,
		Project:  gc.project,
	}
	fmt.Printf("start machine request: %+v\n", req)
	resp, err := c.Start(ctx, req)
	if err != nil {
		return err
	}

	fmt.Printf("%+v", resp)
	return nil
}
