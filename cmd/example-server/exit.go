package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"

	"github.com/hashicorp/go-multierror"
)

type action func() error

func runAndWaitForExit(shutdown func(), runList ...action) error {
	ctx, shutdown := withCancel(shutdown)
	runList = append(runList, waitForSignalAction(ctx))
	var res sync.Map
	var once sync.Once
	var wg sync.WaitGroup
	run := func(idx int, item action) {
		err := invoke(item)
		if err != nil {
			res.Store(idx, err)
		}
		once.Do(shutdown)
		wg.Done()
	}
	wg.Add(len(runList))
	for i, item := range runList {
		go run(i, item)
	}
	wg.Wait()
	var err error
	res.Range(func(key, value interface{}) bool {
		err = multierror.Append(err, value.(error))
		return true
	})
	return err
}

func withCancel(shutdown func()) (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	return ctx, func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("shutdown panic: %v, stack: %s\n", r, string(debug.Stack()))
			}
		}()
		cancel()
		shutdown()
	}
}

func invoke(item action) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v, stack: %s", r, string(debug.Stack()))
		}
	}()
	return item()
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
		log.Println("shutdown received")
		return nil
	case <-ctx.Done():
		return nil
	}
}
