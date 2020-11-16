package v1

import (
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type EtcdCluster struct {
	// TypeMeta is the metadata for the resource, like kind and apiversion
	metav1.TypeMeta `json:",inline"`
	// ObjectMeta contains the metadata for the particular object, including
	// things like...
	//  - name
	//  - namespace
	//  - self link
	//  - labels
	//  - ... etc ...
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the custom resource spec
	Spec ClusterSpec `json:"spec"`
	// Status is the custom resource status
	Status ClusterStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// EtcdClusterList is a list of EtcdCluster resources
type EtcdClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EtcdCluster `json:"items"`
}

type ClusterSpec struct {
	// Size is the expected size of the etcd cluster.
	// The etcd-operator will eventually make the size of the running
	// cluster equal to the expected size.
	// The valid range of the size is from 1 to 7.
	Size int `json:"size"`

	// Repository is the name of the repository that hosts
	// etcd container images. It should be direct clone of the repository in official
	// release:
	//   https://github.com/coreos/etcd/releases
	// That means, it should have exact same tags and the same meaning for the tags.
	//
	// By default, it is `quay.io/coreos/etcd`.
	Repository string `json:"repository,omitempty"`

	// Version is the expected version of the etcd cluster.
	// The etcd-operator will eventually make the etcd cluster version
	// equal to the expected version.
	//
	// The version must follow the [semver]( http://semver.org) format, for example "3.2.13".
	// Only etcd released versions are supported: https://github.com/coreos/etcd/releases
	//
	// If version is not set, default is "3.2.13"
	Version string `json:"versin,omitempty"`

	// Paused is to pause the control of the operator for the etcd cluster.
	Paused bool `json:"paused,omitempty"`

	// Pod defines the policy to create pod for the etcd pod.
	//
	// Updating Pod does not take effect on any existing etcd pods.
	Pod *PodPolicy `json:"pod,omitempty"`
}

type PodPolicy struct {
	// Lables specifies the labels to attach to pods the operator creates for the
	// etcd cluster.
	// "app" and "etcd_*" labels are reserved for the internal use of the etcd operator.
	// Do not overwrite them.
	Labels map[string]string `json:"labels,omitempty"`

	// NodeSelector specifies a map of key-value pairs. For the pod to be eligible
	// to run on a node, the node must have each of the indicated key-value pairs as
	// labels.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// The scheduling constraints on etcd pods.
	Affinity *v1.Affinity `json:"affinity,omitempty"`

	// Resources is the resource requirements for the etcd container.
	// This field cannot be updated once the cluster is created.
	Resources v1.ResourceRequirements `json:"resources,omitempty"`

	// Tolerations specifies the pod's tolerations.
	Tolerations []v1.Toleration `json:"tolerations,omitempty"`

	// List of environment variables to set in the etcd container.
	// This is used to cofigure etcd process. etcd cluster cannot be created, when
	// bad environment variables are provided. Do not overwrite any flags used to
	// bootstrap the cluster (for example `--initial-cluster` flag).
	// This field cannot be updated.
	EtcdEnv []v1.EnvVar `json:"etcdEnv,omitempty"`

	// PersistentVolumeClaimSpec is the spec to describe PVC for the etcd container
	// This field is optional, If no PVC spec, etcd container will use emptyDir as volume
	// Note. This feature is in alpha stage. It is currently only used as non-stable storage,
	// not the stable storage. Future work need to make it used as stable storage.
	PersistentVolumeClaimSpec *v1.PersistentVolumeClaimSpec `json:"persistentVolumeClaimSpec,omitempty"`

	// Annotations specifies the annotations to attach to pods the operator creates for the
	// etcd cluster.
	// The "etcd.version" annotation is reserved for the internal use of the etcd operator.
	Annotations map[string]string `json:"annotations,omitempty"`

	// busybox init container image. default is busybox:1.28.0-glibc
	// busybox:latest uses uclibc which contains a bug that sometimes prevents name resolution
	// More info: https://github.com/docker-library/busybox/issues/27
	BusyboxImage string `json:"busyboxImage,omitempty"`

	// SecurityContext specifies the security context for the entire pod
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/security-context
	SecurityContext *v1.PodSecurityContext `json:"securityContext,omitempty"`

	// DNSTimeoutInSecond is the maximum allowed time for the init container of the etcd pod to
	// reverse DNS lookup its IP given the hostname.
	// The default is to wait indefinitely and has a vaule of 0.
	DNSTimeoutInSecond int64 `json:"DNSTimeoutInSecond,omitempty"`

	// ClusterDomain is the cluster domain to use for member URLs E.g.
	// '.cluster.local'.
	// The default is to not set a cluster domain explicitly.
	ClusterDomain string `json:"ClusterDomain"`
}

type ClusterPhase string
type ClusterConditionType string

const (
	ClusterPhaseNone     ClusterPhase = ""
	ClusterPhaseCreating              = "Creating"
	ClusterPhaseRunning               = "Running"
	ClusterPhaseFailed                = "Failed"

	ClusterConditionAvailable  ClusterConditionType = "Available"
	ClusterConditionRecovering                      = "Recovering"
	ClusterConditionScaling                         = "Scaling"
	ClusterConditionUpgrading                       = "Upgrading"
)

type ClusterStatus struct {
	// Phase is the cluster running phase
	Phase  ClusterPhase `json:"phase"`
	Reason string       `json:"reason,omitempty"`

	// ControlPuased indicates the operator pauses the control of the cluster.
	ControlPaused bool `json:"controlPaused,omitempty"`

	// Condition keeps track of all cluster conditions, if they exist.
	Conditions []ClusterCondition `json:"conditions,omitempty"`

	// Size is the current size of the cluster
	Size int `json:"size"`

	// ServiceName is the LB service for accessing etcd nodes.
	ServiceName string `json:"serviceName,omitempty"`

	// ClientPort is the port for etcd client to access.
	// It's the same on client LB service and etcd nodes.
	ClientPort int `json:"clientPort,omitemtpy"`

	// Members are the etcd members in the cluster
	Members MembersStatus `json:"members"`

	// CurrentVersion is the current cluster version
	CurrentVersion string `json:"currentVersion"`

	// TargetVersion is the version the cluster upgrading to.
	// If the cluster is not upgrading, TargetVersion is empty.
	TargetVersion string `json:"targetVersion"`
}

// ClusterCondition represents one current condition of an etcd cluster.
// A condition might not show up if it is not happening.
// For example, if a cluster is not upgrading, the Upgrading condition would not show up.
// If a cluster is upgrading and encountered a problem that prevents the upgrade,
// the Upgrading condition's status will would be False and communicate the problem back.
type ClusterCondition struct {
	// Type of cluster condition.
	Type ClusterConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status"`
	// The last time this condition was updated.
	LastUpdateTime string `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

type MembersStatus struct {
	// Ready are the etcd members that are ready to serve requests
	// The member names are the same as the etcd pod names
	Ready []string `json:"ready,omitempty"`
	// Unready are the etcd members not ready to serve requests
	Unready []string `json:"unready,omitempty"`
}

func (cs *ClusterStatus) IsFailed() bool {
	if cs == nil {
		return false
	}
	return cs.Phase == ClusterPhaseFailed
}

func (cs *ClusterStatus) SetPhase(p ClusterPhase) {
	cs.Phase = p
}

func (cs *ClusterStatus) PauseControl() {
	cs.ControlPaused = true
}

func (cs *ClusterStatus) Control() {
	cs.ControlPaused = false
}

func (cs *ClusterStatus) UpgradeVersionTo(v string) {
	cs.TargetVersion = v
}

func (cs *ClusterStatus) SetVersion(v string) {
	cs.TargetVersion = ""
	cs.CurrentVersion = v
}

func (cs *ClusterStatus) SetReason(r string) {
	cs.Reason = r
}

func (cs *ClusterStatus) SetScalingUpCondition(from, to int) {
	c := newClusterCondition(ClusterConditionScaling, v1.ConditionTrue, "Scaling up", scalingMsg(from, to))
	cs.setClusterCondition(*c)
}

func (cs *ClusterStatus) SetScalingDownCondition(from, to int) {
	c := newClusterCondition(ClusterConditionScaling, v1.ConditionTrue, "Scaling down", scalingMsg(from, to))
	cs.setClusterCondition(*c)
}

func (cs *ClusterStatus) SetRecoveringCondition() {
	c := newClusterCondition(ClusterConditionRecovering, v1.ConditionTrue,
		"Disaster recovery", "Majority is down. Recovering from backup")
	cs.setClusterCondition(*c)

	cs.ClearCondition(ClusterConditionAvailable)
}

func (cs *ClusterStatus) SetUpgradingCondition(to string) {
	// TODO: show x/y members has upgraded.
	c := newClusterCondition(ClusterConditionUpgrading, v1.ConditionTrue,
		"Cluster upgrading", "upgrading to "+to)
	cs.setClusterCondition(*c)
}

func (cs *ClusterStatus) SetReadyCondition() {
	c := newClusterCondition(ClusterConditionAvailable, v1.ConditionTrue, "Cluster available", "")
	cs.setClusterCondition(*c)
}

func (cs *ClusterStatus) ClearCondition(t ClusterConditionType) {
	pos, _ := getClusterCondition(cs, t)
	if pos == -1 {
		return
	}
	cs.Conditions = append(cs.Conditions[:pos], cs.Conditions[pos+1:]...)
}

func (cs *ClusterStatus) setClusterCondition(c ClusterCondition) {
	pos, cp := getClusterCondition(cs, c.Type)
	if cp != nil &&
		cp.Status == c.Status && cp.Reason == c.Reason && cp.Message == c.Message {
		return
	}

	if cp != nil {
		cs.Conditions[pos] = c
	} else {
		cs.Conditions = append(cs.Conditions, c)
	}
}

func getClusterCondition(status *ClusterStatus, t ClusterConditionType) (int, *ClusterCondition) {
	for i, c := range status.Conditions {
		if t == c.Type {
			return i, &c
		}
	}
	return -1, nil
}

func newClusterCondition(condType ClusterConditionType, status v1.ConditionStatus, reason, message string) *ClusterCondition {
	now := time.Now().Format(time.RFC3339)
	return &ClusterCondition{
		Type:               condType,
		Status:             status,
		LastUpdateTime:     now,
		LastTransitionTime: now,
		Reason:             reason,
		Message:            message,
	}
}

func scalingMsg(from, to int) string {
	return fmt.Sprintf("Current cluster size: %d, desired cluster size: %d", from, to)
}
