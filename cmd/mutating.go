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
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/kubernetes-sigs/pspmigrator"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var MutatingCmd = &cobra.Command{
	Use:   "mutating",
	Short: "Check if pods or PSP objects are mutating",
}

func initMutating() {
	podCmd := cobra.Command{
		Use:   "pod [name of pod]",
		Short: "Check if a pod is being mutated by a PSP policy",
		Run: func(cmd *cobra.Command, args []string) {
			pod := args[0]
			podObj, err := clientset.CoreV1().Pods(Namespace).Get(context.TODO(), pod, metav1.GetOptions{})
			if errors.IsNotFound(err) {
				fmt.Printf("Pod %s in namespace %s not found\n", pod, Namespace)
				os.Exit(1)
			} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
				fmt.Printf("Error getting pod %s in namespace %s: %v\n",
					pod, Namespace, statusError.ErrStatus.Message)
				os.Exit(1)
			} else if err != nil {
				log.Fatalln(err.Error())
				os.Exit(1)
			} else {
				mutated, diff, err := pspmigrator.IsPodBeingMutatedByPSP(podObj, clientset, ContainersToIgnore)
				if err != nil {
					log.Println(err)
					os.Exit(1)
				}
				if pspName, ok := podObj.ObjectMeta.Annotations["kubernetes.io/psp"]; ok {
					fmt.Printf("Pod %v is mutated by PSP %v: %v, diff: %v\n", podObj.Name, pspName, mutated, diff)
					pspObj, err := clientset.PolicyV1beta1().PodSecurityPolicies().Get(context.TODO(), pspName, metav1.GetOptions{})
					if errors.IsNotFound(err) {
						fmt.Printf("PodSecurityPolicy %s not found\n", pspName)
					} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
						fmt.Printf("Error getting PodSecurityPolicy %s: %v\n",
							pspName, statusError.ErrStatus.Message)
					} else if err != nil {
						panic(err.Error())
					} else {
						_, fields, annotations := pspmigrator.IsPSPMutating(pspObj)
						fmt.Printf("PSP profile %v has the following mutating fields: %v and annotations: %v\n", pspName, fields, annotations)
					}

				}
			}
		},
		Args: cobra.ExactArgs(1),
	}

	podCmd.Flags().StringVarP(&Namespace, "namespace", "n", "", "K8s namespace (required)")
	podCmd.Flags().StringVarP(&ContainersToIgnore, "containersToIgnore", "c", "", "comma-separated list of containers to ignore in the live pod spec at comparison time")
	podCmd.MarkFlagRequired("namespace")

	podsCmd := cobra.Command{
		Use:   "pods",
		Short: "Check all pods across all namespaces in a cluster are being mutated by a PSP policy",
		Run: func(cmd *cobra.Command, args []string) {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "Namespace", "Mutated", "PSP"})
			pods, err := GetPods()
			if err != nil {
				log.Fatalln("Error getting pods", err.Error())
			}
			fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
			for _, pod := range pods.Items {
				if pspName, ok := pod.ObjectMeta.Annotations["kubernetes.io/psp"]; ok {
					mutated, _, err := pspmigrator.IsPodBeingMutatedByPSP(&pod, clientset, "")
					if err != nil {
						log.Println("error occured checking if pod is mutated:", err)
					}
					table.Append([]string{pod.Name, pod.Namespace, strconv.FormatBool(mutated), pspName})
				}
			}
			table.Render() // Send output
		},
		Args: cobra.NoArgs,
	}

	pspCmd := cobra.Command{
		Use:   "psp [name of PSP object]",
		Short: "Check if a PSP object is potentially mutating pods",
		Run: func(cmd *cobra.Command, args []string) {
			// Examples for error handling:
			// - Use helper functions like e.g. errors.IsNotFound()
			// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
			pspName := args[0]
			pspObj, err := clientset.PolicyV1beta1().PodSecurityPolicies().Get(context.TODO(), pspName, metav1.GetOptions{})
			if errors.IsNotFound(err) {
				fmt.Printf("PodSecurityPolicy %s not found\n", pspName)
				os.Exit(1)
			} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
				fmt.Printf("Error getting PodSecurityPolicy %s: %v\n",
					pspName, statusError.ErrStatus.Message)
				os.Exit(1)
			} else if err != nil {
				log.Fatalln(err.Error())
				os.Exit(1)
			} else {
				_, fields, annotations := pspmigrator.IsPSPMutating(pspObj)
				fmt.Printf("PSP profile %v has the following mutating fields: %v and annotations: %v\n", pspName, fields, annotations)
			}

		},
		Args: cobra.ExactArgs(1),
	}

	MutatingCmd.AddCommand(&podCmd)
	MutatingCmd.AddCommand(&podsCmd)
	MutatingCmd.AddCommand(&pspCmd)
}
