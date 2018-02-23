package controller

import(
	"time"
	"context"
	"fmt"

	co_v1aplha1 "github.com/aslanbekirov/cassandra-operator/pkg/apis/cassandra.database.com/v1alpha1"
)

const (
	// Copy from deployment_controller.go:
	// maxRetries is the number of times a etcd backup will be retried before it is dropped out of the queue.
	// With the current rate-limiter in use (5ms*2^(maxRetries-1)) the following numbers represent the times
	// an etcd backup is going to be requeued:
	//
	// 5ms, 10ms, 20ms, 40ms, 80ms, 160ms, 320ms, 640ms, 1.3s, 2.6s, 5.1s, 10.2s, 20.4s, 41s, 82s
	maxRetries = 15
	DefaultRequestTimeout = 5 * time.Second
	// DefaultBackupTimeout is the default maximal allowed time of the entire backup process.
	DefaultBackupTimeout    = 1 * time.Minute
	
)

func (c *Cluster) runWorker() {
	for c.processNextItem() {
	}
}

func (c *Cluster) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	// Tell the queue that we are done with processing this key. This unblocks the key for other workers
	// This allows safe parallel processing because two pods with the same key are never processed in
	// parallel.
	defer c.queue.Done(key)
	err := c.processItem(key.(string))
	// Handle the error if something went wrong during the execution of the business logic
	c.handleErr(err, key)
	return true
}

func (c *Cluster) processItem(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	cc := obj.(*co_v1aplha1.CassandraCluster)
	// don't process the CR if it has a status since
	// having a status means that the backup is either made or failed.
	// if cc.Status.Succeeded || len(cc.Status.Reason) != 0 {
	// 	return nil
	// }
	_, err = c.handleCluster(&cc.Spec)
	// Report backup status
	//c.reportBackupStatus(cs, err, cc)
	return err
}

// func (c *Cluster) reportBackupStatus(cs *co_v1aplha1.CassandraClusterStatus, berr error, cc *co_v1aplha1.CassandraCluster) {
// 	if berr != nil {
// 		eb.Status.Succeeded = false
// 		eb.Status.Reason = berr.Error()
// 	} else {
// 		eb.Status.Succeeded = true
// 		eb.Status.EtcdRevision = bs.EtcdRevision
// 		eb.Status.EtcdVersion = bs.EtcdVersion
// 	}
// 	_, err := b.backupCRCli.EtcdV1beta2().EtcdBackups(b.namespace).Update(eb)
// 	if err != nil {
// 		b.logger.Warningf("failed to update status of backup CR %v : (%v)", eb.Name, err)
// 	}
// }


func (c *Cluster) handleErr(err error, key interface{}) {
	if err == nil {
		// Forget about the #AddRateLimited history of the key on every successful synchronization.
		// This ensures that future processing of updates for this key is not delayed because of
		// an outdated error history.
		c.queue.Forget(key)
		return
	}

	// This controller retries maxRetries times if something goes wrong. After that, it stops trying.
	if c.queue.NumRequeues(key) < maxRetries {
		c.logger.Errorf("error syncing etcd backup (%v): %v", key, err)

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// Report that, even after several retries, we could not successfully process this key
	c.logger.Infof("Dropping etcd backup (%v) out of the queue: %v", key, err)
}

func (c *Cluster) handleCluster(spec *co_v1aplha1.CassandraClusterSpec) (*co_v1aplha1.CassandraClusterStatus, error) {
	err := validate(spec)
	if err != nil {
		return nil, err
	}

	// When BackupPolicy.TimeoutInSecond <= 0, use default DefaultBackupTimeout.
	backupTimeout := time.Duration(DefaultBackupTimeout)
	
	ctx, cancel := context.WithTimeout(context.Background(), backupTimeout)
	defer cancel()
	
    //TODO: create cassandra cluster using statefulsets here
    createCluster(ctx, c.namespace)
	return nil, nil
}

// TODO: move this to initializer
func validate(spec *co_v1aplha1.CassandraClusterSpec) error {
	// if len(spec.EtcdEndpoints) == 0 {
	// 	return errors.New("spec.etcdEndpoints should not be empty")
	// }
	return nil
}

func createCluster(ctx context.Context, namespace string){
  fmt.Println("Creating cassandra clsuter")
}