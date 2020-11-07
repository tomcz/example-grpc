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
)

type action func() error

func runAndWaitForExit(shutdown func(), runList ...action) error {
	ctx, shutdown := withCancel(shutdown)
	runList = append(runList, waitForSignalAction(ctx))
	var res sync.Map
	var once sync.Once
	var wg sync.WaitGroup
	run := func(item action) {
		err := invoke(item)
		if err != nil {
			log.Println(err)
			res.LoadOrStore("error", err)
		}
		once.Do(shutdown)
		wg.Done()
	}
	wg.Add(len(runList))
	for _, item := range runList {
		go run(item)
	}
	wg.Wait()
	if err, ok := res.Load("error"); ok {
		return err.(error)
	}
	return nil
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
