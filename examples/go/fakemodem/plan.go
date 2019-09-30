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
	"log"
	"sync"
	"time"

	api "github.com/infostellarinc/go-stellarstation/api/v1/groundstation"
)

/****************
 * Plan Watcher *
 ****************/

// PlanWatcher periodically checks for plans for a ground station.
type PlanWatcher struct {
	runner    *Runner
	client    *Client
	plans     map[string]*api.Plan
	plansLock *sync.Mutex
}

// NewPlanWatcher creates a new PlanWatcher instance
func NewPlanWatcher(client *Client) *PlanWatcher {
	watcher := &PlanWatcher{
		runner:    NewRunner(),
		plans:     make(map[string]*api.Plan),
		plansLock: &sync.Mutex{},
		client:    client,
	}
	return watcher
}

// Start begins checking for plans.
func (w *PlanWatcher) Start(planCheckInterval time.Duration) {
	startFunction := func() {
		log.Printf("Starting plan watcher.\n")

		go w.UpdatePlans()

		updatePlans := time.NewTicker(planCheckInterval)

		go func() {
			for {
				select {
				case <-w.runner.Done():
					return
				case <-updatePlans.C:
					go w.UpdatePlans()
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
func (w *PlanWatcher) UpdatePlans() {
	w.plansLock.Lock()
	defer w.plansLock.Unlock()
	plans, err := w.client.ListPlans()
	if err != nil {
		log.Printf("Failed to list plans: %v\n", err)
		return
	}
	for id := range w.plans {
		delete(w.plans, id)
	}
	for _, plan := range plans {
		w.plans[plan.PlanId] = plan
	}
}
