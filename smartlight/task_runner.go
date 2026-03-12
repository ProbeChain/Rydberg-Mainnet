// Copyright 2024 The ProbeChain Authors
// This file is part of the ProbeChain.
//
// The ProbeChain is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The ProbeChain is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the ProbeChain. If not, see <http://www.gnu.org/licenses/>.

package smartlight

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/probechain/go-probe/common"
	"github.com/probechain/go-probe/log"
)

// TaskRunner executes lightweight agent tasks with resource limits.
// Maximum 2 concurrent tasks, 25MB/task, 5-second timeout.
type TaskRunner struct {
	config  *Config
	address common.Address
	signFn  SignerFunc
	sender  PeerSender

	activeTasks int32 // atomic counter
	tasksDone   uint64
	tasksOK     uint64

	mu   sync.Mutex
	quit chan struct{}
}

// NewTaskRunner creates a new agent task runner.
func NewTaskRunner(config *Config, address common.Address, signFn SignerFunc, sender PeerSender) *TaskRunner {
	return &TaskRunner{
		config:  config,
		address: address,
		signFn:  signFn,
		sender:  sender,
		quit:    make(chan struct{}),
	}
}

// Submit submits a task for execution. Returns immediately.
// The task is dropped if the runner is at capacity.
func (r *TaskRunner) Submit(task *TaskRequestMsg) {
	if task == nil {
		return
	}

	// Check capacity
	current := atomic.LoadInt32(&r.activeTasks)
	if int(current) >= r.config.MaxAgentTasks {
		log.Debug("SmartLight: task runner at capacity, dropping task", "taskID", task.TaskID.Hex())
		r.sendResult(task.TaskID, nil, false)
		return
	}

	// Check memory limit
	if len(task.Payload) > r.config.MaxTaskMemoryMB*1024*1024 {
		log.Warn("SmartLight: task payload too large", "taskID", task.TaskID.Hex(), "size", len(task.Payload))
		r.sendResult(task.TaskID, nil, false)
		return
	}

	atomic.AddInt32(&r.activeTasks, 1)
	go r.execute(task)
}

// execute runs a single task with timeout and resource limits.
func (r *TaskRunner) execute(task *TaskRequestMsg) {
	defer atomic.AddInt32(&r.activeTasks, -1)

	timeout := time.Duration(task.TimeoutMs) * time.Millisecond
	if timeout == 0 || timeout > r.config.TaskTimeout {
		timeout = r.config.TaskTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resultCh := make(chan []byte, 1)
	errCh := make(chan error, 1)

	go func() {
		result, err := r.runTask(ctx, task)
		if err != nil {
			errCh <- err
		} else {
			resultCh <- result
		}
	}()

	select {
	case result := <-resultCh:
		r.sendResult(task.TaskID, result, true)
		r.mu.Lock()
		r.tasksDone++
		r.tasksOK++
		r.mu.Unlock()

	case err := <-errCh:
		log.Warn("SmartLight: task failed", "taskID", task.TaskID.Hex(), "err", err)
		r.sendResult(task.TaskID, nil, false)
		r.mu.Lock()
		r.tasksDone++
		r.mu.Unlock()

	case <-ctx.Done():
		log.Warn("SmartLight: task timed out", "taskID", task.TaskID.Hex())
		r.sendResult(task.TaskID, nil, false)
		r.mu.Lock()
		r.tasksDone++
		r.mu.Unlock()

	case <-r.quit:
		return
	}
}

// runTask executes the task payload. This is a lightweight sandbox for agent tasks.
func (r *TaskRunner) runTask(ctx context.Context, task *TaskRequestMsg) ([]byte, error) {
	// For now, the task runner processes the payload as-is.
	// In Phase 2, this will integrate with the probe-lang agent framework
	// to execute PROBE language tasks with proper sandboxing.
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// Echo task payload as placeholder — real agent execution comes in Phase 2
		return task.Payload, nil
	}
}

// sendResult sends a task result to the full node.
func (r *TaskRunner) sendResult(taskID common.Hash, result []byte, success bool) {
	// Sign the result
	msg := make([]byte, 32+1)
	copy(msg[0:32], taskID[:])
	if success {
		msg[32] = 1
	}
	sig, err := r.signFn(msg)
	if err != nil {
		log.Warn("SmartLight: task result sign failed", "err", err)
		return
	}

	resultMsg := TaskResultMsg{
		TaskID:    taskID,
		Address:   r.address,
		Result:    result,
		Success:   success,
		Signature: sig,
	}
	if err := r.sender.Send(SLPTaskResultMsg, resultMsg); err != nil {
		log.Warn("SmartLight: task result send failed", "err", err)
	}
}

// Stop shuts down the task runner.
func (r *TaskRunner) Stop() {
	select {
	case <-r.quit:
	default:
		close(r.quit)
	}
}

// Stats returns task execution statistics.
func (r *TaskRunner) Stats() (done, succeeded uint64, active int32) {
	r.mu.Lock()
	done = r.tasksDone
	succeeded = r.tasksOK
	r.mu.Unlock()
	active = atomic.LoadInt32(&r.activeTasks)
	return
}
