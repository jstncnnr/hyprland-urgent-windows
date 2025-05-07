package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/jstncnnr/go-hyprland/hypr"
	"github.com/jstncnnr/go-hyprland/hypr/commands"
	"github.com/jstncnnr/go-hyprland/hypr/event"
	"os"
	"os/signal"
	"slices"
	"syscall"
)

func main() {
	client, err := events.NewClient()
	if err != nil {
		fmt.Printf("Error creating event client: %v\n", err)
		os.Exit(1)
	}

	client.RegisterListener(eventListener)

	// Setup interrupt handler so we can cleanly close the event client
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		interrupt, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		<-interrupt.Done()
		cancel()
	}()

	if err := client.Listen(ctx); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Printf("Error running event client: %v\n", err)
		os.Exit(1)
	}
}

var urgentWindows = make([]string, 1)

func eventListener(event events.Event) {
	switch event.(type) {
	case events.UrgentEvent:
		event := event.(events.UrgentEvent)

		urgentWindows = append(urgentWindows, event.WindowAddress)
		addTag(event.WindowAddress)
		break

	// V2 event provides the window address
	case events.ActiveWindowV2Event:
		event := event.(events.ActiveWindowV2Event)

		if slices.Contains(urgentWindows, event.WindowAddress) {
			removeTag(event.WindowAddress)
			urgentWindows = slices.DeleteFunc(urgentWindows, func(s string) bool {
				return s == event.WindowAddress
			})
		}
		break

	// Cleanup any windows that are closed before we get a chance to
	// remove the urgent tag.
	case events.CloseWindowEvent:
		event := event.(events.CloseWindowEvent)

		if slices.Contains(urgentWindows, event.WindowAddress) {
			urgentWindows = slices.DeleteFunc(urgentWindows, func(s string) bool {
				return s == event.WindowAddress
			})
		}
	}
}

func addTag(address string) {
	err := hypr.NewRequest().
		Dispatch("tagwindow", fmt.Sprintf("+urgent address:%s", address)).
		Send()

	if err != nil {
		fmt.Printf("Error adding urgent tag: %v\n", err)
		_ = hypr.NewRequest().
			Notify(commands.IconError, 1500, "0", fmt.Sprintf("Error adding urgent tag: %v", err)).
			Send()
	}
}

func removeTag(address string) {
	err := hypr.NewRequest().
		Dispatch("tagwindow", fmt.Sprintf("-urgent address:%s", address)).
		Send()

	if err != nil {
		fmt.Printf("Error removing urgent tag: %v\n", err)
		_ = hypr.NewRequest().
			Notify(commands.IconError, 1500, "0", fmt.Sprintf("Error removing urgent tag: %v", err)).
			Send()
	}
}
