/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var RootCmd = &cobra.Command{
	Use:   "pspmigrator",
	Short: "pspmigrator is a tool to help migrate from PSP to PSA",
}

var (
	clientset          *kubernetes.Clientset
	err                error
	Namespace          string
	ContainersToIgnore string
)

func init() {
	initMutating()
	RootCmd.AddCommand(MutatingCmd)
	RootCmd.AddCommand(MigrateCmd)
	var kubeconfig string

	if home := homedir.HomeDir(); home != "" {
		RootCmd.PersistentFlags().StringVarP(&kubeconfig, "kubeconfig", "k",
			filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		RootCmd.PersistentFlags().StringVarP(&kubeconfig, "kubeconfig", "k", "", "absolute path to the kubeconfig file")
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	config.UserAgent = "pspmigrator"
	config.WarningHandler = rest.NoWarnings{}

	// create the clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
}
