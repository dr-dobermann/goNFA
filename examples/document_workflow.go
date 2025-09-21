// Document workflow example demonstrates basic usage of the goNFA library.
// This example implements a simple document approval workflow with states:
// Draft -> InReview -> Approved/Rejected -> InReview (rework)
//
// goNFA is a universal, lightweight and idiomatic Go library for creating
// and managing non-deterministic finite automata (NFA). It provides reliable
// state management mechanisms for complex systems such as business process
// engines (BPM).
//
// Project: https://github.com/dr-dobermann/gonfa
// Author: dr-dobermann (rgabtiov@gmail.com)
// License: LGPL-2.1 (see LICENSE file in the project root)
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/dr-dobermann/gonfa/pkg/builder"
	"github.com/dr-dobermann/gonfa/pkg/gonfa"
	"github.com/dr-dobermann/gonfa/pkg/machine"
)

// Document represents a document in the workflow
type Document struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	Reviewer string `json:"reviewer,omitempty"`
}

// LogAction logs workflow events
type LogAction struct {
	message string
}

func (l *LogAction) Execute(ctx context.Context, payload gonfa.Payload) error {
	doc, ok := payload.(*Document)
	if !ok {
		return fmt.Errorf("expected *Document, got %T", payload)
	}
	fmt.Printf("[LOG] %s - Document: %s (ID: %s)\n",
		l.message, doc.Title, doc.ID)
	return nil
}

// AssignReviewerAction assigns a reviewer to the document
type AssignReviewerAction struct{}

func (a *AssignReviewerAction) Execute(
	ctx context.Context,
	payload gonfa.Payload,
) error {
	doc, ok := payload.(*Document)
	if !ok {
		return fmt.Errorf("expected *Document, got %T", payload)
	}
	doc.Reviewer = "John Doe" // In real app, would use proper logic
	fmt.Printf("[ACTION] Assigned reviewer '%s' to document '%s'\n",
		doc.Reviewer, doc.Title)
	return nil
}

// NotifyAuthorAction sends notification to document author
type NotifyAuthorAction struct{}

func (n *NotifyAuthorAction) Execute(
	ctx context.Context,
	payload gonfa.Payload,
) error {
	doc, ok := payload.(*Document)
	if !ok {
		return fmt.Errorf("expected *Document, got %T", payload)
	}
	fmt.Printf("[NOTIFY] Notified author '%s' about document '%s'\n",
		doc.Author, doc.Title)
	return nil
}

// IsManagerGuard checks if the user is a manager
type IsManagerGuard struct {
	userRole string
}

func (g *IsManagerGuard) Check(
	ctx context.Context,
	payload gonfa.Payload,
) bool {
	// In real app, would check user permissions from context
	isManager := g.userRole == "manager"
	fmt.Printf("[GUARD] Manager check: %v\n", isManager)
	return isManager
}

func main() {
	fmt.Println("=== Document Workflow Example ===")

	// Create the state machine definition using Builder
	definition, err := builder.New().
		InitialState(gonfa.State("Draft")).
		// Define state actions
		OnEntry(gonfa.State("InReview"), &AssignReviewerAction{}).
		OnExit(gonfa.State("InReview"), &LogAction{
			message: "Leaving review state",
		}).
		OnEntry(gonfa.State("Approved"), &LogAction{
			message: "Document approved!",
		}).
		// Define transitions
		AddTransition(
			gonfa.State("Draft"),
			gonfa.State("InReview"),
			gonfa.Event("Submit"),
		).
		WithActions(&NotifyAuthorAction{}).
		AddTransition(
			gonfa.State("InReview"),
			gonfa.State("Approved"),
			gonfa.Event("Approve"),
		).
		WithGuards(&IsManagerGuard{userRole: "manager"}).
		AddTransition(
			gonfa.State("InReview"),
			gonfa.State("Rejected"),
			gonfa.Event("Reject"),
		).
		WithGuards(&IsManagerGuard{userRole: "manager"}).
		AddTransition(
			gonfa.State("Rejected"),
			gonfa.State("InReview"),
			gonfa.Event("Rework"),
		).
		// Add global hooks
		WithSuccessHooks(&LogAction{message: "Transition successful"}).
		WithFailureHooks(&LogAction{message: "Transition failed"}).
		Build()

	if err != nil {
		log.Fatalf("Failed to build definition: %v", err)
	}

	// Create a machine instance
	sm := machine.NewMachine(definition)

	// Create a document
	doc := &Document{
		ID:     "DOC-001",
		Title:  "Project Proposal",
		Author: "Alice Smith",
	}

	ctx := context.Background()

	fmt.Printf("Initial state: %s\n\n", sm.CurrentState())

	// Submit document for review
	fmt.Println("1. Submitting document for review...")
	success, err := sm.Fire(ctx, gonfa.Event("Submit"), doc)
	if err != nil {
		log.Printf("Error during Submit: %v", err)
	}
	fmt.Printf("Submit success: %v, Current state: %s\n\n",
		success, sm.CurrentState())

	// Try to approve (should succeed as manager)
	fmt.Println("2. Approving document as manager...")
	success, err = sm.Fire(ctx, gonfa.Event("Approve"), doc)
	if err != nil {
		log.Printf("Error during Approve: %v", err)
	}
	fmt.Printf("Approve success: %v, Current state: %s\n\n",
		success, sm.CurrentState())

	// Demonstrate serialization
	fmt.Println("3. Serializing machine state...")
	storable, err := sm.Marshal()
	if err != nil {
		log.Fatalf("Failed to marshal machine state: %v", err)
	}

	jsonData, err := json.MarshalIndent(storable, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal to JSON: %v", err)
	}

	fmt.Printf("Serialized state:\n%s\n\n", jsonData)

	// Demonstrate restoration
	fmt.Println("4. Restoring machine from serialized state...")
	var restoredStorable gonfa.Storable
	err = json.Unmarshal(jsonData, &restoredStorable)
	if err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	restoredMachine, err := machine.RestoreMachine(definition, &restoredStorable)
	if err != nil {
		log.Fatalf("Failed to restore machine: %v", err)
	}

	fmt.Printf("Restored machine state: %s\n", restoredMachine.CurrentState())

	// Show history
	history := restoredMachine.History()
	fmt.Printf("Transition history (%d entries):\n", len(history))
	for i, entry := range history {
		fmt.Printf("  %d. %s -> %s (event: %s) at %s\n",
			i+1, entry.From, entry.To, entry.On,
			entry.Timestamp.Format("15:04:05"))
	}

	fmt.Println("\n=== Example completed successfully ===")
}
