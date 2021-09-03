// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package jobscheduler provides simple functionalities for reliable
// scheduling of Faktory jobs.
//
// It is a common pattern to schedule one or more new jobs when a certain
// system entity (a record) is created or updated on the database. If
// used properly, this package can help achieve reliability between the
// data persisted on database and the jobs that should rely upon it.
//
// Being Faktory and the database two separate external systems, failures and
// latencies can be expected at any level and point of execution.
//
// Let's consider the following illustrative scenario: the program should
// create a new "Thing" object (database table "things") with auto-generated
// ID (primary key) "1" and should also schedule the execution of a Faktory
// job "ProcessThing", specifying the new model's ID. That job will read
// the record from the DB and do something with it.
//
// It could be trivially implemented with two steps, like this:
//   1. create the new "Thing" record on the DB (and get the ID "1")
//   2. create and push the new job "ProcessThing(1)"
//
// If something goes wrong in between those two steps, you would end up with
// the new record on the DB, but the job would never be scheduled.
//
// If losing a job is not critical, or detecting a missed job can be done by
// simply looking at the status of the data from the database, then this might
// be the simplest desirable implementation.
//
// Otherwise, the job loss could be naively solved by inverting the two steps,
// also involving a transaction so that we can know the new ID before
// committing data to the database:
//
//   1. begin transaction
//   2.   create the new "Thing" record (and get the ID "1")
//   3.   create and push the new job "ProcessThing(1)"
//   4. end transaction, actually committing data to the DB
//
// Even if we try to reduce as much as possible the time (and actual
// instructions) between steps 3 and 4, it's still possible that the job will
// run before the data is committed on DB. This might be annoying but
// not critical: the job might simply fail and be retried later, possibly with
// success.
// But things can get worse: the transaction might fail entirely, or be abruptly
// interrupted for any reason. In this case, the job will always fail to fetch
// the new record and its successive retrials are rather a weak point.
//
// The models.PendingJob model can be used as an additional supportive record
// to increase reliability and reduce the downsides of both scenarios depicted
// above.
//
// A new pattern can be established following these steps:
//
//   1. begin transaction
//   2.   create the new "Thing" record (and get the ID "1")
//   3.   create the new job "ProcessThing(1)", but DO NOT push it yet
//   4.   create a new PendingJob object
//        (which contains the serialization of the job created above)
//   5. end transaction, actually committing data to the DB
//   6. if the transaction failed, stop here, otherwise go on
//   7. push the job ("ProcessThing(1)", that was created on step 3)
//   8. if the push failed, stop here, otherwise proceed with the last step
//   9. delete the PendingJob (that was created on step 4 and persisted on
//      step 5) from the database
//
// This approach is somehow closer to the initial two-step scenario, in that
// the new job is pushed to Faktory server after the new "Thing" is persisted
// to the DB (step 7). However, here we also keep track of the intention to
// push a new job within the database.
//
// Creating the two records (the "Thing" and the PendingJob) from the same
// transaction provides the guarantee that either both are created, or both
// are discarded. In case of transaction failure, the guard condition at step
// 6 prevents the actual job to be pushed.
//
// If pushing the job fails, then the PendingJob is still there. It could
// be used later for recovery. For example, a separate process could
// periodically check for existing old PendingJob records; each of them can
// be deserialized and pushed again, finally the record can be removed
// upon success.
//
// The price to pay for this kind of reliability is the chance for
// the same job to be scheduled (pushed) more than once. This can happen
// if something goes wrong between steps 7 and 9: the job is pushed, but the
// PendingJob is not removed, so the recovery process will possibly try to
// push it again. The recovery process implementation itself might be
// susceptible to the same issue as well.
//
// Job idempotency, or any other sort of tolerance to the scheduling of
// identical jobs, is left to the jobs implementations, or to additional
// features provided by Faktory.
//
// The previous examples involve single records for the sake of simplicity,
// but in many real cases it might be possible to consider many records
// in batch at once, still keeping the same order of operations.
//
// This package provides a JobScheduler. It can be used as a supportive
// structure for the implementation of some of the described steps,
// offering simple facilities to reduce repetitive boilerplate code and
// minimize the chance of programming errors.
package jobscheduler
