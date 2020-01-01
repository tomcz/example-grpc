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

func waitForExit(shutdown func(), runList ...action) {
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
