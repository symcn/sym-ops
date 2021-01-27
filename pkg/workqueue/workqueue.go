package workqueue

import (
	"context"
	"fmt"
	"time"

	"github.com/symcn/sym-ops/pkg/types"
	"golang.org/x/time/rate"
	ktypes "k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
)

// Queue wrapper workqueue
type Queue struct {
	name        string
	namespace   []string
	threadiness int
	gotIntervel time.Duration
	workqueue   workqueue.RateLimitingInterface
	stats       *stats
	Do          types.Reconciler
}

func NewQueue(reconcile types.Reconciler, name string, threadiness int, gotInterval time.Duration, namespace ...string) (types.WorkQueue, error) {
	stats, err := buildStats(name)
	if err != nil {
		return nil, err
	}

	return &Queue{
		name:        name,
		namespace:   namespace,
		threadiness: threadiness,
		gotIntervel: gotInterval,
		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.NewMaxOfRateLimiter(
			workqueue.NewItemExponentialFailureRateLimiter(1*time.Second, 60*time.Second),
			&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
		), name),
		stats: stats,
		Do:    reconcile,
	}, nil
}

func (q *Queue) Add(item interface{}) {
	q.workqueue.Add(item)
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (q *Queue) Run(ctx context.Context) error {
	defer utilruntime.HandleCrash()
	defer q.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting WrapQueue compoment")

	klog.Info("Starting workers")
	// Launch two workers to process Foo resources
	for i := 0; i < q.threadiness; i++ {
		go wait.UntilWithContext(ctx, q.runWorker, q.gotIntervel)
	}

	klog.Info("Started WrapQueue workers")
	<-ctx.Done()
	klog.Info("Shutting down WrapQueue")
	return nil
}

func (q *Queue) runWorker(ctx context.Context) {
	for q.processNextWorkItem() {
	}
}

func (q *Queue) processNextWorkItem() bool {
	obj, shutdown := q.workqueue.Get()
	if shutdown {
		return false
	}
	q.stats.Dequeue.Inc()

	start := time.Now()
	defer func() {
		q.stats.ReconcileDuration.Observe(float64(time.Since(start)))
	}()

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer q.workqueue.Done(obj)

		// TODO: invoke Reconcile
		var req ktypes.NamespacedName
		var ok bool
		if req, ok = obj.(ktypes.NamespacedName); !ok {
			q.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected types.NamespacedName in workqueue but got %#v", obj))
			q.stats.UnExpectedObj.Inc()
			return nil
		}

		requeue, after, err := q.Do.Reconcile(req)
		if err != nil {
			q.workqueue.AddRateLimited(req)
			q.stats.ReconcileFail.Inc()
			q.stats.RequeueRateLimit.Inc()
			return nil
		}

		q.stats.ReconcileSucc.Inc()

		if after > 0 {
			q.workqueue.Forget(obj)
			q.workqueue.AddAfter(req, after)
			q.stats.RequeueAfter.Inc()
			return nil
		}
		if requeue == types.Requeue {
			q.workqueue.AddRateLimited(req)
			q.stats.RequeueRateLimit.Inc()
			return nil
		}

		q.workqueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}
