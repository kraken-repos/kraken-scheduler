package main

import (
	scheduler "kraken.dev/kraken-scheduler/pkg/reconciler"
	"knative.dev/pkg/injection/sharedmain"
)

const (
	component = "integration-framework-scheduler"
)

func main() {
	sharedmain.Main(component, scheduler.NewController)
}
