package queue

import (
	"context"
	"time"
)

const (
	// inboxReadFrequency is the minimum time between reads from the inbox.
	inboxReadFrequency = 1 * time.Millisecond
)

// Queue is a queue of DecodedSSVMessage with dynamic (per-pop) prioritization.
type Queue interface {
	// Push blocks until the message is pushed to the queue.
	Push(*DecodedSSVMessage)

	// TryPush returns immediately with true if the message was pushed to the queue,
	// or false if the queue is full.
	TryPush(*DecodedSSVMessage) bool

	// Pop returns and removes the next message in the queue, or blocks until a message is available.
	// When the context is canceled, Pop returns immediately with any leftover message or nil.
	Pop(context.Context, MessagePrioritizer) *DecodedSSVMessage

	// TryPop returns immediately with the next message in the queue, or nil if there is none.
	TryPop(MessagePrioritizer) *DecodedSSVMessage

	// Empty returns true if the queue is empty.
	Empty() bool
}

type priorityQueue struct {
	head     *item
	inbox    chan *DecodedSSVMessage
	lastRead time.Time
}

// New returns an implementation of Queue optimized for concurrent push and sequential pop.
// Pops aren't thread-safe, so don't call Pop from multiple goroutines.
func New(capacity int) Queue {
	return &priorityQueue{
		inbox: make(chan *DecodedSSVMessage, capacity),
	}
}

// NewDefault returns an implementation of Queue optimized for concurrent push and sequential pop,
// with a capacity of 32 and a PusherDropping.
func NewDefault() Queue {
	return New(32)
}

func (q *priorityQueue) Push(msg *DecodedSSVMessage) {
	q.inbox <- msg
}

func (q *priorityQueue) TryPush(msg *DecodedSSVMessage) bool {
	select {
	case q.inbox <- msg:
		return true
	default:
		return false
	}
}

func (q *priorityQueue) TryPop(prioritizer MessagePrioritizer) *DecodedSSVMessage {
	// Read any pending messages from the inbox.
	q.readInbox()

	// Pop the highest priority message.
	if q.head != nil {
		return q.pop(prioritizer)
	}

	return nil
}

func (q *priorityQueue) Pop(ctx context.Context, prioritizer MessagePrioritizer) *DecodedSSVMessage {
	// Read any pending messages from the inbox, if enough time has passed.
	// inboxReadFrequency is a tradeoff between responsiveness and computational cost,
	// since reading the inbox is more expensive than just reading the head.
	if time.Since(q.lastRead) > inboxReadFrequency {
		q.readInbox()
	}

	// Try to pop immediately.
	if q.head != nil {
		return q.pop(prioritizer)
	}

	// Wait for a message to be pushed.
	select {
	case msg := <-q.inbox:
		q.head = &item{message: msg}
	case <-ctx.Done():
	}

	// Read any messages that were pushed while waiting.
	q.readInbox()

	// Pop the highest priority message.
	if q.head != nil {
		return q.pop(prioritizer)
	}

	return nil
}

func (q *priorityQueue) readInbox() {
	q.lastRead = time.Now()

	for {
		select {
		case msg := <-q.inbox:
			if q.head == nil {
				q.head = &item{message: msg}
			} else {
				q.head = &item{message: msg, next: q.head}
			}
		default:
			return
		}
	}
}

func (q *priorityQueue) pop(prioritizer MessagePrioritizer) *DecodedSSVMessage {
	if q.head.next == nil {
		m := q.head.message
		q.head = nil
		return m
	}

	// Remove the highest priority message and return it.
	var (
		prior   *item
		highest = q.head
		current = q.head
	)
	for {
		if prioritizer.Prior(current.next.message, highest.message) {
			highest = current.next
			prior = current
		}
		current = current.next
		if current.next == nil {
			break
		}
	}
	if prior == nil {
		q.head = highest.next
	} else {
		prior.next = highest.next
	}
	return highest.message
}

func (q *priorityQueue) Empty() bool {
	return q.head == nil && len(q.inbox) == 0
}

// item is a node in a linked list of DecodedSSVMessage.
type item struct {
	message *DecodedSSVMessage
	next    *item
}
