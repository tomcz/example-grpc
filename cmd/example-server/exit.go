package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
)

type action func() error

func runAndWaitForExit(shutdown func(), runList ...action) {
	ctx, shutdown := withCancel(shutdown)
	runList = append(runList, waitForSignalAction(ctx))
	var wg sync.WaitGroup
	var once sync.Once
	for _, runItem := range runList {
		wg.Add(1)
		go func(runThis action) {
			if err := handlePanic(runThis); err != nil {
				log.Println(err)
			}
			once.Do(shutdown)
			wg.Done()
		}(runItem)
	}
	wg.Wait()
}

func withCancel(shutdown func()) (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	return ctx, func() {
		cancel()
		shutdown()
	}
}

func handlePanic(runThis action) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v, stack: %s", r, string(debug.Stack()))
		}
	}()
	return runThis()
}

func waitForSignalAction(ctx context.Context) action {
	return func() error {
		return waitForSignal(ctx)
	}
}

func waitForSignal(ctx context.Context) error {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-signalChan:
		return errors.New("shutdown received")
	case <-ctx.Done():
		return ctx.Err()
	}
}
