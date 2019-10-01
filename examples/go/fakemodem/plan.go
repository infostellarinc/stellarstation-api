/*
 * Copyright 2019 Infostellar, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"

	api "github.com/infostellarinc/go-stellarstation/api/v1/groundstation"
)

/****************
 * Plan Watcher *
 ****************/

type PlanStartFunction func(plan *api.Plan)
type PlanEndFunction func(plan *api.Plan)

// PlanWatcher periodically checks for plans for a ground station.
type PlanWatcher struct {
	runner    *Runner
	client    *Client
	plans     []*api.Plan
	plansLock *sync.Mutex
}

// NewPlanWatcher creates a new PlanWatcher instance
func NewPlanWatcher(client *Client) *PlanWatcher {
	watcher := &PlanWatcher{
		runner:    NewRunner(),
		plans:     make([]*api.Plan, 0),
		plansLock: &sync.Mutex{},
		client:    client,
	}
	return watcher
}

// Start begins checking for plans.
func (w *PlanWatcher) Start(planCheckInterval time.Duration, planStart PlanStartFunction, planEnd PlanEndFunction) {
	startFunction := func(ctx context.Context) {
		log.Printf("Starting plan watcher.\n")

		go w.UpdatePlans(planStart, planEnd)

		updatePlans := time.NewTicker(planCheckInterval)
		checkPlans := time.NewTicker(time.Second)

		go func() {
			for {
				select {
				case <-ctx.Done():
					updatePlans.Stop()
					checkPlans.Stop()
					return
				case <-updatePlans.C:
					go w.UpdatePlans(planStart, planEnd)
				case <-checkPlans.C:
					w.CheckPlanState(planStart, planEnd)
				}
			}
		}()
	}

	stopFunction := func() {
		log.Printf("Shutting down plan watcher\n")
	}

	w.runner.Start(startFunction, stopFunction)
}

// Stop stops checking for plans
func (w *PlanWatcher) Stop() {
	w.runner.Stop()
}

// Wait will wait for the watcher to stop watching for plans
func (w *PlanWatcher) Wait() {
	w.runner.Wait()
}

// UpdatePlans updates the plan list for the current ground station
func (w *PlanWatcher) UpdatePlans(planStart PlanStartFunction, planEnd PlanEndFunction) {
	w.plansLock.Lock()
	defer w.plansLock.Unlock()

	now := time.Now()
	start := now.Add(-time.Hour)
	end := now.Add(time.Hour)

	plans, err := w.client.ListPlans(start, end)
	if err != nil {
		log.Printf("Failed to list plans: %v\n", err)
		return
	}

	existingPlans := make(map[string]bool, len(w.plans))
	for _, plan := range w.plans {
		existingPlans[plan.PlanId] = true
	}

	for _, plan := range plans {
		_, found := existingPlans[plan.PlanId]
		if !found {
			log.Printf("===== New Plan. %v\n", shortPlanData(plan))
			planStartTime, _, _, planEndTime, err := parseTimestamps(plan)
			if err != nil {
				log.Printf("Could not parse plan timestamps. Err: %+v\n", err)
				continue
			}

			startDelta := now.Sub(planStartTime)
			endDelta := now.Sub(planEndTime)

			if startDelta > time.Second && endDelta <= -time.Second {
				log.Printf("----- Resuming Plan. %v\n", shortPlanData(plan))
				go planStart(plan)
			}
		}
	}

	w.plans = plans
}

// CheckPlanState checks for any plan state changes that occurred in the past second
func (w *PlanWatcher) CheckPlanState(planStart PlanStartFunction, planEnd PlanEndFunction) {
	w.plansLock.Lock()
	defer w.plansLock.Unlock()
	now := time.Now()
	for i, plan := range w.plans {
		if plan == nil {
			continue
		}

		start, _, _, end, err := parseTimestamps(plan)
		if err != nil {
			log.Printf("Could not parse plan timestamps. Err: %+v\n", err)
			w.plans[i] = nil
			continue
		}

		startDelta := now.Sub(start)
		endDelta := now.Sub(end)

		if startDelta >= 0 && startDelta <= time.Second {
			go planStart(plan)
		}

		if endDelta >= 0 && endDelta <= time.Second {
			go planEnd(plan)
		}
	}
}

func parseTimestamps(plan *api.Plan) (start, aos, los, end time.Time, err error) {
	if plan == nil {
		err = fmt.Errorf("could not parse plan timestamps: no plan provided")
		return
	}

	start, err = ptypes.Timestamp(plan.StartTime)
	if err != nil {
		err = fmt.Errorf("could not parse plan start time: %+v", err)
		return
	}

	end, err = ptypes.Timestamp(plan.EndTime)
	if err != nil {
		err = fmt.Errorf("could not parse plan end time: %+v", err)
		return
	}
	aos, err = ptypes.Timestamp(plan.AosTime)
	if err != nil {
		err = fmt.Errorf("could not parse plan aos time: %+v", err)
		return
	}

	los, err = ptypes.Timestamp(plan.LosTime)
	if err != nil {
		err = fmt.Errorf("could not parse plan los time: %+v", err)
		return
	}
	return
}

func localTimeString(t time.Time) string {
	return t.Local().Format("15:04:05")
}

func shortPlanData(plan *api.Plan) string {
	start, aos, los, end, err := parseTimestamps(plan)
	if err != nil {
		return fmt.Sprintf("Plan ID: %v, Error: %v\n", plan.PlanId, err)
	}
	return fmt.Sprintf("Plan ID: %v, Start: %v, AOS: %v, LOS: %v, End: %v (%v)",
		plan.PlanId,
		localTimeString(start),
		localTimeString(aos),
		localTimeString(los),
		localTimeString(end),
		end.Local().Format("-0700 MST"))
}
