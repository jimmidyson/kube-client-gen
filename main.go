package main

import "github.com/jimmidyson/kube-client-gen/cmd/generate"

func main() {
	if err := generate.RootCmd.Execute(); err != nil {
		panic(err)
	}
}
