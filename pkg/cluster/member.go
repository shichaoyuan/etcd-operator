package cluster

import (
	"fmt"

	"etcd-operator/pkg/util/etcdutil"
	"etcd-operator/pkg/util/k8sutil"
	"etcd/etcdserver/etcdserverpb"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

func (c *Cluster) updateMembers(known etcdutil.MemberSet) error {
	resp, err := etcdutil.ListMembers(known.ClientURLs(), c.tlsConfig)
	if err != nil {
		return err
	}
	members := etcdutil.MemberSet{}
	for _, m := range resp.Members {
		name, err := getMemberName(m, c.cluster.GetName())
		if err != nil {
			return errors.Wrap(err, "get member name failed")
		}

		members[name] = &etcdutil.Member{
			Name:         name,
			Namespace:    c.cluster.Namespace,
			ID:           m.ID,
			SecurePeer:   c.isSecurePeer(),
			SecureClient: c.isSecureClient(),
		}
	}
	c.members = members
	return nil
}

func (c *Cluster) newMember() *etcdutil.Member {
	name := k8sutil.UniqueMemberName(c.cluster.Name)
	m := &etcdutil.Member{
		Name:         name,
		Namespace:    c.cluster.Namespace,
		SecurePeer:   c.isSecurePeer(),
		SecureClient: c.isSecureClient(),
	}

	if c.cluster.Spec.Pod != nil {
		m.ClusterDomain = c.cluster.Spec.Pod.ClusterDomain
	}
	return m
}

func podsToMemberSet(pods []*v1.Pod, sc bool) etcdutil.MemberSet {
	members := etcdutil.MemberSet{}
	for _, pod := range pods {
		m := &etcdutil.Member{Name: pod.Name, Namespace: pod.Namespace, SecureClient: sc}
		members.Add(m)
	}
	return members
}

func getMemberName(m *etcdserverpb.Member, clusterName string) (string, error) {
	name, err := etcdutil.MemberNameFromPeerURL(m.PeerURLs[0])
	if err != nil {
		return "", newFatalError(fmt.Sprintf("invalid member peerURL (%s): %v", m.PeerURLs[0], err))
	}
	return name, nil
}
