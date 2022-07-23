package leaderelection

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/coordination/v1"
	k8sleaderelection "k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

type Client struct {
	client          v1.CoordinationV1Interface
	lockName        string
	namespace       string
	identity        string
	leaseDuration   time.Duration
	renewDeadline   time.Duration
	retryPeriod     time.Duration
	releaseOnCancel bool

	onStartedLeading func(c context.Context)
	onStoppedLeading func()
	onNewLeader      func(currentID, lockID string)
}

func (c *Client) RegisterOnStart(callback func(c context.Context)) {
	c.onStartedLeading = callback
}

func (c *Client) RegisterOnStop(callback func()) {
	c.onStoppedLeading = callback
}

func (c *Client) RegisterOnNew(callback func(currentID string, lockID string)) {
	c.onNewLeader = callback
}

func (c *Client) Run(ctx context.Context) {
	k8sleaderelection.RunOrDie(ctx, k8sleaderelection.LeaderElectionConfig{
		Lock:            c.createLock(),
		ReleaseOnCancel: c.releaseOnCancel,
		LeaseDuration:   c.leaseDuration,
		RenewDeadline:   c.renewDeadline,
		RetryPeriod:     c.retryPeriod,
		Callbacks: k8sleaderelection.LeaderCallbacks{
			OnStartedLeading: c.onStartedLeading,
			OnStoppedLeading: c.onStoppedLeading,
			OnNewLeader: func(identity string) {
				c.onNewLeader(identity, c.identity)
			},
		},
	})
}

func (c *Client) createLock() *resourcelock.LeaseLock {
	return &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      c.lockName,
			Namespace: c.namespace,
		},
		Client: c.client,
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: c.identity,
		},
	}
}

func New(
	client v1.CoordinationV1Interface,
	lockName string,
	namespace string,
	identity string,
	leaseDuration time.Duration,
	renewDeadline time.Duration,
	retryPeriod time.Duration,
	releaseOnCancel bool,
) *Client {
	return &Client{
		client,
		lockName,
		namespace,
		identity,
		leaseDuration,
		renewDeadline,
		retryPeriod,
		releaseOnCancel,
		func(c context.Context) {},
		func() {},
		func(currentID, lockID string) {},
	}
}
