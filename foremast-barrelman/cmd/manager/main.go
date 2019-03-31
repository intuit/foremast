/*
Copyright 2018 The Kubernetes Authors.
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

package main

import (
	"foremast.ai/foremast/foremast-barrelman/pkg/client/informers/externalversions"
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"log"
	"os"
	"strings"
	"time"

	"foremast.ai/foremast/foremast-barrelman/pkg/apis"
	"foremast.ai/foremast/foremast-barrelman/pkg/controller"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	clientset "foremast.ai/foremast/foremast-barrelman/pkg/client/clientset/versioned"
	kubeinformers "k8s.io/client-go/informers"
)

func main() {
	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		log.Fatal(err)
	}

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	log.Printf("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Fatal(err)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	foremastClient, err := clientset.NewForConfig(cfg)

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)

	sharedInformerFactory := externalversions.NewSharedInformerFactory(foremastClient, time.Second*10)

	var mode = os.Getenv("MODE")
	if len(mode) == 0 {
		mode = controller.MODE_HPA_AND_HEALTHY_MONITORING
	}
	var hpaStrategy = os.Getenv("HPA_STRATEGY")
	if len(hpaStrategy) == 0 {
		hpaStrategy = controller.HPA_STRATEGY_ENABLED_ONLY
	}

	barrelman := controller.NewBarrelman(kubeClient, foremastClient, mode, hpaStrategy)

	if strings.Contains(mode, "healthy_monitoring") { //If it runs as "hpa_and_healthy_monitoring", start the deployment controller
		deploymentController := controller.NewDeploymentController(kubeClient, foremastClient,
			kubeInformerFactory.Apps().V1().Deployments(), barrelman)

		log.Printf("Starting the deploymentController.")

		// Start the Cmd
		//log.Fatal(mgr.Start(signals.SetupSignalHandler()))

		if err = deploymentController.Run(2, stopCh); err != nil {
			glog.Fatalf("Error running controller: %s", err.Error())
		}
	}

	monitorController := controller.NewController(kubeClient, foremastClient, sharedInformerFactory.Deployment().V1alpha1().DeploymentMonitors(), barrelman)

	if monitorController != nil {
		log.Printf("Monitor controller started.")
	}

	go sharedInformerFactory.Start(stopCh)

	go kubeInformerFactory.Start(stopCh)

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Fatal(err)
	}

}
