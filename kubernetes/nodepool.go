package kubernetes

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/MagaluCloud/mgc-sdk-go/client"
	mgc_http "github.com/MagaluCloud/mgc-sdk-go/internal/http"
	"github.com/MagaluCloud/mgc-sdk-go/internal/utils"
)

const (
	nodePoolIdField    = "nodePoolID"
	clusterIdField     = "clusterID"
	clusterNodepoolURL = "/v0/clusters/%s/node_pools/%s"
)

type (
	// ListOptions provides options for listing resources
	ListOptions struct {
		Limit  *int
		Offset *int
		Sort   *string
		Expand []string
	}

	// NodePoolService provides methods for managing Kubernetes node pools
	NodePoolService interface {
		Nodes(ctx context.Context, clusterID, nodePoolID string) ([]Node, error)
		List(ctx context.Context, clusterID string, opts ListOptions) ([]NodePool, error)
		Create(ctx context.Context, clusterID string, req CreateNodePoolRequest) (*NodePool, error)
		Get(ctx context.Context, clusterID, nodePoolID string) (*NodePool, error)
		Update(ctx context.Context, clusterID, nodePoolID string, req PatchNodePoolRequest) (*NodePool, error)
		Delete(ctx context.Context, clusterID, nodePoolID string) error
	}

	// NodePoolList represents the response when listing node pools
	NodePoolList struct {
		Results []NodePool `json:"results"`
	}

	// InstanceTemplate represents the template for node instances
	InstanceTemplate struct {
		Flavor    Flavor `json:"flavor"`
		NodeImage string `json:"node_image"`
		DiskSize  int    `json:"disk_size"`
		DiskType  string `json:"disk_type"`
	}

	// NodePool represents a Kubernetes node pool
	NodePool struct {
		ID                string            `json:"id"`
		Name              string            `json:"name"`
		InstanceTemplate  InstanceTemplate  `json:"instance_template"`
		Replicas          int               `json:"replicas"`
		Zone              *[]string         `json:"zone,omitempty"`
		Tags              *[]string         `json:"tags"`
		Labels            map[string]string `json:"labels,omitempty"`
		Taints            *[]Taint          `json:"taints,omitempty"`
		SecurityGroups    *[]string         `json:"security_groups,omitempty"`
		CreatedAt         *time.Time        `json:"created_at"`
		UpdatedAt         *time.Time        `json:"updated_at,omitempty"`
		AutoScale         *AutoScale        `json:"auto_scale,omitempty"`
		Status            Status            `json:"status"`
		Flavor            string            `json:"flavor"`
		MaxPodsPerNode    *int              `json:"max_pods_per_node,omitempty"`
		AvailabilityZones *[]string         `json:"availability_zones,omitempty"`
	}

	// Addresses represents network addresses
	Addresses struct {
		Address string `json:"address"`
		Type    string `json:"type"`
	}

	// Allocatable represents allocatable resources
	Allocatable struct {
		CPU              string `json:"cpu"`
		EphemeralStorage string `json:"ephemeral_storage"`
		Hugepages1Gi     string `json:"hugepages_1Gi"`
		Hugepages2Mi     string `json:"hugepages_2Mi"`
		Memory           string `json:"memory"`
		Pods             string `json:"pods"`
	}

	// Capacity represents total capacity
	Capacity struct {
		CPU              string `json:"cpu"`
		EphemeralStorage string `json:"ephemeral_storage"`
		Hugepages1Gi     string `json:"hugepages_1Gi"`
		Hugepages2Mi     string `json:"hugepages_2Mi"`
		Memory           string `json:"memory"`
		Pods             string `json:"pods"`
	}

	// Infrastructure represents node infrastructure information
	Infrastructure struct {
		Allocatable             Allocatable `json:"allocatable"`
		Architecture            string      `json:"architecture"`
		Capacity                Capacity    `json:"capacity"`
		ContainerRuntimeVersion string      `json:"containerRuntimeVersion"`
		KernelVersion           string      `json:"kernelVersion"`
		KubeProxyVersion        string      `json:"kubeProxyVersion"`
		KubeletVersion          string      `json:"kubeletVersion"`
		OperatingSystem         string      `json:"operatingSystem"`
		OsImage                 string      `json:"osImage"`
	}

	// Node represents a Kubernetes node
	Node struct {
		Addresses      []Addresses       `json:"addresses"`
		Annotations    map[string]string `json:"annotations"`
		ClusterName    string            `json:"cluster_name"`
		CreatedAt      time.Time         `json:"created_at"`
		Flavor         string            `json:"flavor"`
		ID             string            `json:"id"`
		Infrastructure Infrastructure    `json:"infrastructure"`
		Labels         map[string]string `json:"labels"`
		Name           string            `json:"name"`
		Namespace      string            `json:"namespace"`
		NodeImage      string            `json:"node_image"`
		NodepoolName   string            `json:"nodepool_name"`
		Status         MessageState      `json:"status"`
		Taints         *[]Taint          `json:"taints,omitempty"`
		Zone           *string           `json:"zone,omitempty"`
	}

	// CreateNodePoolRequest represents the request payload for creating a node pool
	CreateNodePoolRequest struct {
		Name              string     `json:"name"`
		Flavor            string     `json:"flavor"`
		Replicas          int        `json:"replicas"`
		Tags              *[]string  `json:"tags,omitempty"`
		Taints            *[]Taint   `json:"taints,omitempty"`
		AutoScale         *AutoScale `json:"auto_scale,omitempty"`
		MaxPodsPerNode    *int       `json:"max_pods_per_node,omitempty"`
		AvailabilityZones *[]string  `json:"availability_zones,omitempty"`
	}

	// PatchNodePoolRequest represents the request payload for updating a node pool
	PatchNodePoolRequest struct {
		Replicas  *int       `json:"replicas,omitempty"`
		AutoScale *AutoScale `json:"auto_scale,omitempty"`
	}

	// Taint represents a node taint
	Taint struct {
		Key    string `json:"key"`
		Value  string `json:"value"`
		Effect string `json:"effect"`
	}

	// AutoScale represents autoscaling configuration
	AutoScale struct {
		MinReplicas *int `json:"min_replicas"`
		MaxReplicas *int `json:"max_replicas"`
	}

	// nodePoolService implements the NodePoolService interface
	nodePoolService struct {
		client *KubernetesClient
	}
)

// Nodes returns a list of nodes in a specific node pool
func (s *nodePoolService) Nodes(ctx context.Context, clusterID, nodePoolID string) ([]Node, error) {
	if clusterID == "" {
		return nil, &client.ValidationError{Field: clusterIdField, Message: utils.CannotBeEmpty}
	}

	if nodePoolID == "" {
		return nil, &client.ValidationError{Field: nodePoolIdField, Message: utils.CannotBeEmpty}
	}

	type NodeList struct {
		Results []Node `json:"results"`
	}

	resp, err := mgc_http.ExecuteSimpleRequestWithRespBody[NodeList](ctx, s.client.newRequest,
		s.client.GetConfig(), http.MethodGet,
		fmt.Sprintf(clusterNodepoolURL+"/nodes", clusterID, nodePoolID), nil, nil)

	if err != nil {
		return nil, err
	}

	return resp.Results, nil
}

// List returns a list of node pools in a cluster with optional filtering and pagination
func (s *nodePoolService) List(ctx context.Context, clusterID string, opts ListOptions) ([]NodePool, error) {
	if clusterID == "" {
		return nil, &client.ValidationError{Field: clusterIdField, Message: utils.CannotBeEmpty}
	}

	query := url.Values{}
	if opts.Limit != nil {
		query.Add("_limit", strconv.Itoa(*opts.Limit))
	}
	if opts.Offset != nil {
		query.Add("_offset", strconv.Itoa(*opts.Offset))
	}
	if opts.Sort != nil {
		query.Add("_sort", *opts.Sort)
	}

	resp, err := mgc_http.ExecuteSimpleRequestWithRespBody[NodePoolList](ctx, s.client.newRequest,
		s.client.GetConfig(), http.MethodGet, fmt.Sprintf("/v1alpha0/clusters/%s/node-pools", clusterID), nil, query)
	if err != nil {
		return nil, err
	}

	return resp.Results, nil
}

// Create creates a new node pool in a cluster
func (s *nodePoolService) Create(ctx context.Context, clusterID string, req CreateNodePoolRequest) (*NodePool, error) {
	if clusterID == "" {
		return nil, &client.ValidationError{Field: clusterIdField, Message: utils.CannotBeEmpty}
	}

	return mgc_http.ExecuteSimpleRequestWithRespBody[NodePool](ctx, s.client.newRequest,
		s.client.GetConfig(), http.MethodPost,
		fmt.Sprintf("/v0/clusters/%s/node_pools", clusterID), req, nil)
}

// Get retrieves detailed information about a specific node pool
func (s *nodePoolService) Get(ctx context.Context, clusterID, nodePoolID string) (*NodePool, error) {
	if clusterID == "" {
		return nil, &client.ValidationError{Field: clusterIdField, Message: utils.CannotBeEmpty}
	}

	if nodePoolID == "" {
		return nil, &client.ValidationError{Field: nodePoolIdField, Message: utils.CannotBeEmpty}
	}

	return mgc_http.ExecuteSimpleRequestWithRespBody[NodePool](ctx, s.client.newRequest,
		s.client.GetConfig(), http.MethodGet,
		fmt.Sprintf(clusterNodepoolURL, clusterID, nodePoolID), nil, nil)
}

// Update updates a node pool's properties
func (s *nodePoolService) Update(ctx context.Context, clusterID, nodePoolID string, req PatchNodePoolRequest) (*NodePool, error) {
	if clusterID == "" {
		return nil, &client.ValidationError{Field: clusterIdField, Message: utils.CannotBeEmpty}
	}

	if nodePoolID == "" {
		return nil, &client.ValidationError{Field: nodePoolIdField, Message: utils.CannotBeEmpty}
	}

	return mgc_http.ExecuteSimpleRequestWithRespBody[NodePool](ctx, s.client.newRequest,
		s.client.GetConfig(), http.MethodPatch,
		fmt.Sprintf(clusterNodepoolURL, clusterID, nodePoolID), req, nil)
}

// Delete removes a node pool from a cluster
func (s *nodePoolService) Delete(ctx context.Context, clusterID, nodePoolID string) error {
	if clusterID == "" {
		return &client.ValidationError{Field: clusterIdField, Message: utils.CannotBeEmpty}
	}

	if nodePoolID == "" {
		return &client.ValidationError{Field: nodePoolIdField, Message: utils.CannotBeEmpty}
	}

	return mgc_http.ExecuteSimpleRequest(ctx, s.client.newRequest,
		s.client.GetConfig(), http.MethodDelete,
		fmt.Sprintf(clusterNodepoolURL, clusterID, nodePoolID), nil, nil)
}
