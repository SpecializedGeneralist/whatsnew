// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package jobscheduler

import (
	"context"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	faktory "github.com/contribsys/faktory/client"
	"github.com/contribsys/faktory_worker_go"
	"gorm.io/gorm"
)

// A JobScheduler allows the reliable scheduling of Faktory jobs.
//
// The jobscheduler package provides a general explanation of the problem
// this object might help to solve.
//
// In practice, a JobScheduler is best suitable for scenarios where one or
// more entities are created/updated on the database and one or more jobs
// must be scheduled in relation to each new/modified entity.
//
// By using a JobScheduler and calling its functions in the right order and
// from the right context can make it easy to schedule new jobs reliably.
//
// Here is an outline of the intended usage:
//
//   js := jobscheduler.New()
//
//   trErr := gormDB.Transaction(func(tx *gorm.DB) error {
//       // Create or update one or more new entities on the database
//       // e.g. "tx.Create(...)" or "tx.Save(...)"
//
//       // ...
//
//       // Add new jobs to the scheduler, with "js.AddJob(...)" or
//       // "js.AddJobs(...)".
//       //
//       // The arguments for these jobs will probably include the IDs of the
//       // models created/updated above (or other similar references).
//       //
//       // This operation creates and collects new Jobs and PendingJobs in
//       // the JobScheduler (js), but does not perform any action against
//       // the database or the Faktory server.
//       //
//       // It's important to check for errors returned by these functions. In
//       // case of errors, the transaction should be aborted returning an
//       // error.
//
//       // ...
//
//       // Once all jobs are added, we can create the pending jobs.
//       //
//       // Note that we are still inside the transaction; indeed, "tx"
//       // is passed to the function (and not the initial "gormDB").
//       err := js.CreatePendingJobs(tx)
//       // It's important to abort the transaction in case of errors.
//       if err != nil {
//           return err;
//       }
//
//       // ...
//   })
//
//   // Here we are outside the transaction. If it failed, we cannot proceed
//   // with the remaining operations.
//   if trErr != nil {
//       panic(trErr) // return / os.Exit / ...
//   }
//
//   // Otherwise, we can assume data persisted successfully, so we can push
//   // the jobs to Faktory server.
//   //
//   // If we are inside the processing function of a Faktory job, we can call
//   // "PushJobs", passing to it the Context provided by faktory_worker_go:
//   pushErr := js.PushJobs(ctx)
//   //
//   // ALTERNATIVELY, you can use a Faktory Client and call this function:
//   pushErr := js.PushJobsWithClient(faktoryClient)
//
//   // In any case, if the push failed, do not proceed further.
//   //
//   // By doing so, the database will preserve the PendingJob records,
//   // which can be found later, on a separate process, to attempt recovery
//   // and schedule them again.
//   if pushErr != nil {
//       panic(trErr) // return / os.Exit / ...
//   }
//
//   // If the jobs were pushed successfully, we can finally remove the
//   // PendingJobs from the database.
//   delErr := js.DeletePendingJobs(gormDB)
//   // You can do what you want with this error: there's nothing more to do
//   // with the JobScheduler in any case.
//   //
//   // An error in jobs deletion will probably means that the PendingJob
//   // records will still be present, despite the jobs being successfully
//   // pushed above. This implies that a separate recovery job will still
//   // find those records and attempt a reschedule. It's up to the
//   // implementation of the jobs and the recovery process to tolerate
//   // and handle duplicated job scheduling, according to the specific needs
//   // and requirements.
//
//   // It's desirable to reduce the time between jobs pushing and PendingJobs
//   // deletion as much as possible. Since the two operations should
//   // always happen in close succession, and in this exact order,
//   // you can use the helper functions "PushJobsAndDeletePendingJobs"
//   // or "PushJobsWithClientAndDeletePendingJobs".
//   // A valid reason for not using them might be special error handling.
type JobScheduler struct {
	jobs        []*faktory.Job
	pendingJobs []*models.PendingJob
}

// New creates a new empty JobScheduler.
func New() *JobScheduler {
	return &JobScheduler{}
}

// AddJob adds to the JobScheduler a new Faktory Job, paired with a
// related PendingJob.
//
// This function does not push the job to the server and does not create
// a new record in the database.
func (js *JobScheduler) AddJob(jobType string, args ...interface{}) error {
	job := faktory.NewJob(jobType, args...)

	pj, err := models.NewPendingJob(job)
	if err != nil {
		return err
	}

	js.jobs = append(js.jobs, job)
	js.pendingJobs = append(js.pendingJobs, pj)
	return nil
}

// AddJobs simply calls AddJob for each job type, using the same arguments for
// all jobs.
func (js *JobScheduler) AddJobs(jobTypes []string, args ...interface{}) error {
	for _, jobType := range jobTypes {
		err := js.AddJob(jobType, args...)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreatePendingJobs creates all the collected PendingJobs in the database.
//
// This function does not push any job to the Faktory server.
//
// All new records are created with a single query.
//
// If the JobScheduler does not contain any pending job, the function simply
// does nothing.
//
// To guarantee reliability, if the jobs to be scheduled are somehow related
// to other database entities that are being created or updated, this function
// should be invoked from the same transaction that creates or update those
// other entities.
//
// In the same scenario, to prevent the new jobs to fail because data is not
// yet persisted in the database, this function should always be called
// before PushJobs.
func (js *JobScheduler) CreatePendingJobs(tx *gorm.DB) error {
	if len(js.pendingJobs) == 0 {
		return nil
	}
	res := tx.Create(js.pendingJobs)
	if res.Error != nil {
		return fmt.Errorf("error creating pending jobs: %w", res.Error)
	}
	return nil
}

// PushJobs pushes all collected Jobs to the Faktory server.
//
// This method must only be called within the context of an executing Faktory
// job. The context is the same context of the job. If this condition is not
// met, the method will panic.
//
// Each job is pushed individually. If one Push fails, the error is
// returned immediately, and any remaining job will not be pushed.
//
// This function does not perform any operation on the database.
//
// If the JobScheduler does not contain any job, the function simply
// does nothing.
//
// To guarantee reliability, if the jobs to be scheduled are somehow related
// to other database entities that are being created or updated, this function
// should be invoked after the database transaction that created or updated
// those other entities together with the pending job records (see
// CreatePendingJobs).
func (js *JobScheduler) PushJobs(ctx context.Context) error {
	if len(js.jobs) == 0 {
		return nil
	}
	help := faktory_worker.HelperFor(ctx)
	return help.With(js.PushJobsWithClient)
}

// PushJobsWithClient performs the same operations of PushJobs, but
// accepts a Faktory Client (instead of a job context).
func (js *JobScheduler) PushJobsWithClient(c *faktory.Client) error {
	for _, job := range js.jobs {
		err := c.Push(job)
		if err != nil {
			return fmt.Errorf("error pushing job: %w", err)
		}
	}
	return nil
}

// DeletePendingJobs deletes the collected PendingJobs from the database.
//
// This function does not perform any operation against the Faktory server.
//
// All records are deleted with a single query.
//
// If the JobScheduler does not contain any pending job, the function simply
// does nothing.
//
// This function should be invoked only after the jobs were successfully
// pushed to the Faktory server (with PushJobs or PushJobsWithClient).
//
// After a successful deletion, the JobScheduler has accomplished its goals
// and can be simply discarded.
func (js *JobScheduler) DeletePendingJobs(tx *gorm.DB) error {
	if len(js.pendingJobs) == 0 {
		return nil
	}
	res := tx.Delete(js.pendingJobs)
	if res.Error != nil {
		return fmt.Errorf("error deleting pending jobs: %w", res.Error)
	}
	return nil
}

// PushJobsAndDeletePendingJobs sequentially calls PushJobs and
// DeletePendingJobs.
//
// If PushJobs fails, the error is returned immediately and the second
// operation is not performed.
func (js *JobScheduler) PushJobsAndDeletePendingJobs(ctx context.Context, tx *gorm.DB) error {
	err := js.PushJobs(ctx)
	if err != nil {
		return err
	}
	return js.DeletePendingJobs(tx)
}

// PushJobsWithClientAndDeletePendingJobs sequentially calls PushJobsWithClient
// and DeletePendingJobs.
//
// If PushJobs fails, the error is returned immediately and the second
// operation is not performed.
func (js *JobScheduler) PushJobsWithClientAndDeletePendingJobs(c *faktory.Client, tx *gorm.DB) error {
	err := js.PushJobsWithClient(c)
	if err != nil {
		return err
	}
	return js.DeletePendingJobs(tx)
}
