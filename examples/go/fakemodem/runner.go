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
	"sync"
)

/*************
 * Run State *
 *************/

// Runner keeps track of whether or not something is running
type Runner struct {
	done        chan struct{}
	running     bool
	runningLock *sync.Mutex
}

type StateChangeFunction func()

// NewRunner creates a new Runner instance
func NewRunner() *Runner {
	state := &Runner{
		running:     false,
		runningLock: &sync.Mutex{},
	}
	return state
}

// Start begins execution
func (r *Runner) Start(startFunction StateChangeFunction, stopFunction StateChangeFunction) {
	// First stop any currently running operations
	r.Stop()
	r.Wait()

	r.runningLock.Lock()
	defer func() {
		r.running = true
		r.runningLock.Unlock()
	}()

	r.done = make(chan struct{})

	go func() {
		<-r.done
		stopFunction()
	}()

	startFunction()
}

// Stop ends execution
func (r *Runner) Stop() {
	r.runningLock.Lock()
	defer func() {
		r.running = false
		r.runningLock.Unlock()
	}()

	if r.running {
		close(r.done)
	}
}

// Wait will wait for execution to end
func (r *Runner) Wait() {
	r.runningLock.Lock()
	running := r.running
	done := r.done
	r.runningLock.Unlock()

	if !running {
		return
	}
	<-done
}

// Done returns the current done channel for use in select statements
func (r *Runner) Done() chan struct{} {
	r.runningLock.Lock()
	done := r.done
	r.runningLock.Unlock()

	return done
}
