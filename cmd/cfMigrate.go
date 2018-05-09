package cmd

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/jinzhu/gorm"
	"gopkg.in/mgo.v2/bson"
	// "context"
	"encoding/json"
	"strings"
	// "os"

	"github.com/checkr/codeflow/server/plugins/codeflow"
	codeamp_plugins "github.com/codeamp/circuit/plugins"
	codeamp "github.com/codeamp/circuit/plugins/codeamp"
	codeamp_resolvers "github.com/codeamp/circuit/plugins/codeamp/resolvers"
	"github.com/go-bongo/bongo"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	mgo "gopkg.in/mgo.v2"
	// "github.com/davecgh/go-spew/spew"
	// uuid "github.com/satori/go.uuid"
)

var codeflowDB *bongo.Connection

// migrateCmd represents the migrate command
var cfMigrateCmd = &cobra.Command{
	Use:   "cfmigrate",
	Short: "Migrate Codeflow projects to CodeAmp",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[+] Codeflow to CodeAmp migration started.\n ---------------------------------------- \n")

		// init DB connection for codeflow
		fmt.Println("[*] Initializing Codeflow DB Connection")
		createCodeflowDBConnection()
		fmt.Println("[+] Successfully initialized Codeflow DB Connection")

		// init DB connection for codeamp
		fmt.Println("[*] Initializing CodeAmp Resolver")
		codeampDB, err := createCodeampDB()
		if err != nil {
			panic(err.Error())
		}
		fmt.Println("[+] Successfully initialized Codeamp Resolver")

		fmt.Println("[*] Initializing CodeAmp transistor...")

		fmt.Println("[+] Successfully initialized CodeAmp transistor")

		// // TODO: Remove for production
		// fmt.Println("[*] Cleaning Codeamp DB of all rows. REMOVE FOR PRODUCTION.")
		// codeampDB.Debug().Unscoped().Delete(&codeamp_resolvers.Service{})
		// codeampDB.Debug().Unscoped().Delete(&codeamp_resolvers.Secret{})
		// codeampDB.Debug().Unscoped().Delete(&codeamp_resolvers.SecretValue{})
		// codeampDB.Debug().Unscoped().Delete(&codeamp_resolvers.Project{})
		// codeampDB.Debug().Unscoped().Delete(&codeamp_resolvers.ServiceSpec{})
		// codeampDB.Debug().Unscoped().Delete(&codeamp_resolvers.Feature{})
		// codeampDB.Debug().Unscoped().Delete(&codeamp_resolvers.Release{})
		// codeampDB.Debug().Unscoped().Delete(&codeamp_resolvers.ProjectExtension{})
		// fmt.Println("[+] Successfully cleaned Codeamp DB of all rows")

		projects := []codeflow.Project{}
		projectsResults := codeflowDB.Collection("projects").Find(bson.M{"deleted": false})
		services := viper.GetStringSlice("services")
		tmpCodeflowProject := codeflow.Project{}

		for projectsResults.Next(&tmpCodeflowProject) {
			for _, service := range services {
				if tmpCodeflowProject.Name == service {
					fmt.Println("found service ", tmpCodeflowProject.Name)
					projects = append(projects, tmpCodeflowProject)
				}
			}
		}

		reg, err := regexp.Compile("[^0-9]+")
		if err != nil {
			panic(err.Error())
		}

		type ServiceNameMapping struct {
			NewServiceName string `json:"newServiceName"`
			OldServiceName string `json:"oldServiceName"`
		}

		newServiceNames := []ServiceNameMapping{}

		// create service specs
		fmt.Println("[*] Porting service specs")
		codeflowServiceSpec := codeflow.ServiceSpec{}
		serviceSpecResults := codeflowDB.Collection("serviceSpecs").Find(bson.M{})
		// results.Query.All(&codeflowServiceSpecs)
		for serviceSpecResults.Next(&codeflowServiceSpec) {
			fmt.Println(fmt.Sprintf("[*] Transferring %s", codeflowServiceSpec.Name))
			cpuRequest, err := strconv.Atoi(reg.ReplaceAllString(codeflowServiceSpec.CpuRequest, ""))
			if err != nil {
				panic(err.Error())
			}

			if strings.Contains(codeflowServiceSpec.CpuRequest, "Gi") {
				cpuRequest = cpuRequest * 1000
			}

			cpuLimit, err := strconv.Atoi(reg.ReplaceAllString(codeflowServiceSpec.CpuLimit, ""))
			if err != nil {
				panic(err.Error())
			}

			if strings.Contains(codeflowServiceSpec.CpuLimit, "Gi") {
				cpuLimit = cpuLimit * 1000
			}

			memRequest, err := strconv.Atoi(reg.ReplaceAllString(codeflowServiceSpec.MemoryRequest, ""))
			if err != nil {
				panic(err.Error())
			}

			if strings.Contains(codeflowServiceSpec.MemoryRequest, "Gi") {
				memRequest = memRequest * 1000
			}

			memLimit, err := strconv.Atoi(reg.ReplaceAllString(codeflowServiceSpec.MemoryLimit, ""))
			if err != nil {
				panic(err.Error())
			}

			if strings.Contains(codeflowServiceSpec.MemoryLimit, "Gi") {
				memLimit = memLimit * 1000
			}

			codeampServiceSpec := codeamp_resolvers.ServiceSpec{
				Name:                   codeflowServiceSpec.Name,
				CpuRequest:             strconv.Itoa(cpuRequest),
				CpuLimit:               strconv.Itoa(cpuLimit),
				MemoryRequest:          strconv.Itoa(memRequest),
				MemoryLimit:            strconv.Itoa(memLimit),
				TerminationGracePeriod: strconv.Itoa(int(codeflowServiceSpec.TerminationGracePeriodSeconds)),
			}
			codeampDB.Debug().Where(codeamp_resolvers.ServiceSpec{Name: codeflowServiceSpec.Name}).Assign(codeampServiceSpec).FirstOrCreate(&codeampServiceSpec)
			fmt.Println(fmt.Sprintf("[+] Successfully transferred %s", codeflowServiceSpec.Name))
		}
		fmt.Println("[+] Finished porting service spec \n\n")

		codeampUser := codeamp_resolvers.User{}
		if codeampDB.Debug().Where("email = ?", "kilgore@kilgore.trout").First(&codeampUser).RecordNotFound() {
			panic("Could not find CodeAmp user with email kilgore@kilgore.trout")
		}
		var wg sync.WaitGroup
		fmt.Println("[*] Porting projects")
		for _, project := range projects {
			fmt.Println("\n\n\n\n\n ---------------------- \n\n [+] STARTING PROJECT ", project.Slug)

			// fmt.Println(fmt.Sprintf("[*] Creating corresponding CodeAmp project for %s", project.Slug))
			codeampProject := codeamp_resolvers.Project{
				Name:          project.Name,
				Slug:          project.Slug,
				Repository:    project.Repository,
				Secret:        project.Secret,
				GitUrl:        project.GitUrl,
				GitProtocol:   project.GitProtocol,
				RsaPrivateKey: project.RsaPrivateKey,
				RsaPublicKey:  project.RsaPublicKey,
			}
			codeampDB.Debug().Where(codeamp_resolvers.Project{
				Slug: project.Slug,
			}).Assign(codeampProject).FirstOrCreate(&codeampProject)

			for {
				if codeflowDB.Session.Ping() == nil {
					break
				}
				codeflowDB.Session.Refresh()
				time.Sleep(1)
			}

			fmt.Println(project.Name)

			// fmt.Println("[*] Porting features")
			// // find the features tied to the project
			// codeflowFeatures := []codeflow.Feature{}
			// results = codeflowDB.Collection("features").Find(bson.M{ "projectId": bson.ObjectId(project.Id) })
			// results.Query.All(&codeflowFeatures)

			// for _, feature := range codeflowFeatures {
			// 	fmt.Println("[*] Porting feature ", feature.Hash)
			// 	// create codeamp feature
			// 	codeampFeature := codeamp_resolvers.Feature{
			// 		ProjectID: codeampProject.Model.ID,
			// 		Message: feature.Message,
			// 		User: feature.User,
			// 		Ref: feature.Ref,
			// 		ParentHash: feature.ParentHash,
			// 		Created: feature.Created,
			// 		Hash: feature.Hash,
			// 	}
			// 	codeampDB.Debug().Create(&codeampFeature)
			// }
			// fmt.Println("[+] Successfully ported features! \n")

			fmt.Println("[*] Porting environments...")
			// get envs in codeamp
			envs := []codeamp_resolvers.Environment{}
			codeampDB.Debug().Where("key = ?", "production").Find(&envs)

			for _, env := range envs {
				fmt.Println(fmt.Sprintf("[*] Filling in environment %s", env.Key))

				fmt.Println("[*] Porting secrets...")
				// find and create the secrets tied to the project
				// secret := codeflow.Secret{}
				codeflowSecrets := []codeflow.Secret{}

				// bson.M{ "deleted": false, "projectId": project.Id } not working
				// so doing a manually-looped filter
				secretResults := codeflowDB.Collection("secrets").Find(bson.M{"projectId": bson.ObjectId(project.Id), "deleted": false})
				secretResults.Query.All(&codeflowSecrets)

				wg.Add(len(codeflowSecrets))

				for _, secret := range codeflowSecrets {

					go func(secret codeflow.Secret, env codeamp_resolvers.Environment, codeampProject codeamp_resolvers.Project) {
						defer wg.Done()

						fmt.Println(fmt.Sprintf("[*] Creating secret %s", secret.Key))
						isSecret := false
						if string(secret.Type) == "protected-env" {
							isSecret = true
						}

						codeampSecretValue := codeamp_resolvers.SecretValue{
							Value:  secret.Value,
							UserID: codeampUser.Model.ID,
						}

						codeampSecret := codeamp_resolvers.Secret{
							Key:           secret.Key,
							Scope:         codeamp_resolvers.GetSecretScope("project"),
							EnvironmentID: env.Model.ID,
							IsSecret:      isSecret,
							ProjectID:     codeampProject.Model.ID,
							Type:          codeamp_plugins.GetType(string(secret.Type)),
							Value:         codeampSecretValue,
						}

						codeampDB.Debug().Where(codeamp_resolvers.Secret{Key: secret.Key, EnvironmentID: env.Model.ID, ProjectID: codeampProject.Model.ID}).Assign(codeampSecret).FirstOrCreate(&codeampSecret)
					}(secret, env, codeampProject)

					// codeampDB.Debug().Create(&codeampSecretValue)
					// codeampSecret.Value = codeampSecretValue

					fmt.Println(fmt.Sprintf("[+] Successfully created Secret %s => %s", secret.Key, secret.Value))
				}
				fmt.Println("[+] Successfully ported secrets! \n\n")

				fmt.Println("[*] Porting services...")
				// find the services tied to the project
				codeflowServices := []codeflow.Service{}
				svcResults := codeflowDB.Collection("services").Find(bson.M{"projectId": bson.ObjectId(project.Id)})
				svcResults.Query.All(&codeflowServices)
				codeampServices := []codeamp_resolvers.Service{}

				wg.Add(len(codeflowServices))

				for _, codeflowService := range codeflowServices {
					go func(codeampProject codeamp_resolvers.Project, env codeamp_resolvers.Environment, codeflowService codeflow.Service) {
						defer wg.Done()

						if string(codeflowService.State) != "deleted" {
							fmt.Println("[*] Porting service ", codeflowService.Name, codeflowService.Id, codeflowService.SpecId)
							// get service spec
							codeflowServiceSpec := codeflow.ServiceSpec{}
							svcSpecResults := codeflowDB.Collection("serviceSpecs").Find(bson.M{"_id": bson.ObjectId(codeflowService.SpecId)})
							svcSpecResults.Query.One(&codeflowServiceSpec)

							codeampServiceSpec := codeamp_resolvers.ServiceSpec{}
							if !codeampDB.Debug().Where("name = ?", codeflowServiceSpec.Name).First(&codeampServiceSpec).RecordNotFound() {
								// fmt.Println(fmt.Sprintf("[-] Could not find ServiceSpec %s in CodeAmp", codeflowServiceSpec.Name))
								codeampServiceType := codeamp_plugins.GetType("general")
								if codeflowService.OneShot {
									codeampServiceType = codeamp_plugins.GetType("one-shot")
								}

								count := codeflowService.Count
								if codeflowService.Count > 2 {
									count = 2
								}

								codeampService := codeamp_resolvers.Service{
									ProjectID:     codeampProject.Model.ID,
									ServiceSpecID: codeampServiceSpec.Model.ID,
									Command:       codeflowService.Command,
									EnvironmentID: env.Model.ID,
									Count:         strconv.Itoa(count),
									Type:          codeampServiceType,
									Name:          codeflowService.Name,
								}
								codeampDB.Debug().Where(codeamp_resolvers.Service{
									ProjectID:     codeampProject.Model.ID,
									Name:          codeflowService.Name,
									EnvironmentID: env.Model.ID,
								}).Assign(codeampService).FirstOrCreate(&codeampService)

								// create ports arr
								codeampPorts := []codeamp_resolvers.ServicePort{}
								for _, codeflowPort := range codeflowService.Listeners {
									codeampPort := codeamp_resolvers.ServicePort{
										ServiceID: codeampService.Model.ID,
										Port:      strconv.Itoa(codeflowPort.Port),
										Protocol:  codeflowPort.Protocol,
									}

									codeampDB.Debug().Where(codeamp_resolvers.ServicePort{
										ServiceID: codeampService.Model.ID,
										Port:      strconv.Itoa(codeflowPort.Port),
										Protocol:  codeflowPort.Protocol,
									}).Assign(codeampPort).FirstOrCreate(&codeampPort)
									codeampPorts = append(codeampPorts, codeampPort)
								}
								codeampService.Ports = codeampPorts
								codeampServices = append(codeampServices, codeampService)
							}
						}
					}(codeampProject, env, codeflowService)
				}
				fmt.Println("[+] Succesfully ported services! \n")

				// create additional objects i.e. ProjectSettings, ProjectEnvironments
				fmt.Println("[*] Creating ProjectSettings... ", env, codeampProject.Slug)
				projectSettings := codeamp_resolvers.ProjectSettings{
					EnvironmentID:    env.Model.ID,
					ProjectID:        codeampProject.Model.ID,
					GitBranch:        "master",
					ContinuousDeploy: project.ContinuousDelivery,
				}

				codeampDB.Debug().Where(codeamp_resolvers.ProjectSettings{EnvironmentID: env.Model.ID, ProjectID: codeampProject.Model.ID}).Assign(projectSettings).FirstOrCreate(&projectSettings)

				fmt.Println("[+] Successfully created ProjectSettings")

				fmt.Println("[*] Creating ProjectEnvironment permission... ", env, codeampProject.Slug)
				projectEnvironment := codeamp_resolvers.ProjectEnvironment{
					EnvironmentID: env.Model.ID,
					ProjectID:     codeampProject.Model.ID,
				}
				codeampDB.Debug().Where(codeamp_resolvers.ProjectEnvironment{EnvironmentID: env.Model.ID, ProjectID: codeampProject.Model.ID}).Assign(projectEnvironment).FirstOrCreate(&projectEnvironment)
				fmt.Println("[+] Successfully created ProjectEnvironment")

				// Create project extensions
				fmt.Println("[*] Creating Project Extensions...")
				// Create DockerBuilder extension
				dockerBuilderDBExtension := codeamp_resolvers.Extension{}
				if codeampDB.Debug().Where("environment_id = ? and key = ?", env.Model.ID, "dockerbuilder").Find(&dockerBuilderDBExtension).RecordNotFound() {
					panic(err.Error())
				}

				dockerBuilderProjectExtension := codeamp_resolvers.ProjectExtension{
					ProjectID:     codeampProject.Model.ID,
					ExtensionID:   dockerBuilderDBExtension.Model.ID,
					State:         codeamp_plugins.GetState("complete"),
					StateMessage:  "Migrated, click update to send an event.",
					Artifacts:     postgres.Jsonb{[]byte("[]")},
					Config:        postgres.Jsonb{[]byte("[]")},
					CustomConfig:  postgres.Jsonb{[]byte("{}")},
					EnvironmentID: env.Model.ID,
				}
				codeampDB.Debug().Where(codeamp_resolvers.ProjectExtension{
					ProjectID:     codeampProject.Model.ID,
					ExtensionID:   dockerBuilderDBExtension.Model.ID,
					EnvironmentID: env.Model.ID,
				}).Assign(dockerBuilderProjectExtension).FirstOrCreate(&dockerBuilderProjectExtension)

				// get relevant information for project's corresponding load balancers in codeflow
				extensionResults := codeflowDB.Collection("extensions").Find(bson.M{
					"projectId": bson.ObjectId(project.Id),
					"extension": "LoadBalancer",
					"state":     "complete",
				})
				codeflowLoadBalancer := codeflow.LoadBalancer{}
				// results.Query.All(&codeflowLoadBalancers)
				for extensionResults.Next(&codeflowLoadBalancer) {
					listenerPairs := []map[string]string{}
					codeflowService := codeflow.Service{}
					codeampService := codeamp_resolvers.Service{}

					err = codeflowDB.Collection("services").FindById(bson.ObjectId(codeflowLoadBalancer.ServiceId), &codeflowService)
					if err != nil {
						fmt.Println(err.Error())
						fmt.Println("[-] Could not find codeflow service with id ", codeflowLoadBalancer.ServiceId)
					}

					if codeampDB.Debug().Where("name = ? and project_id = ? and environment_id = ?", codeflowService.Name, codeampProject.Model.ID, env.Model.ID).Find(&codeampService).RecordNotFound() {
						fmt.Println("[-] Could not find codeamp corresponding service for ", codeflowLoadBalancer.Name)
					}

					for _, cfListenerPair := range codeflowLoadBalancer.ListenerPairs {
						listenerPairs = append(listenerPairs, map[string]string{
							"port":            strconv.Itoa(cfListenerPair.Source.Port),
							"containerPort":   strconv.Itoa(cfListenerPair.Destination.Port), // Get container port id from corresponding service
							"serviceProtocol": strings.ToLower(cfListenerPair.Destination.Protocol),
						})
					}

					name := codeflowLoadBalancer.Subdomain
					if string(codeflowLoadBalancer.Type) == "internal" {
						name = codeflowLoadBalancer.Name
					}

					lbCustomConfig := map[string]interface{}{
						"name":           name,
						"type":           codeflowLoadBalancer.Type,
						"service":        codeflowService.Name,
						"listener_pairs": listenerPairs,
					}
					marshaledLbCustomConfig, err := json.Marshal(lbCustomConfig)
					if err != nil {
						panic(err.Error())
					}

					lbConfig := []map[string]interface{}{
						map[string]interface{}{
							"key":    "codeflow_id",
							"value":  codeflowService.Id.Hex(),
							"secret": false,
						},
					}
					marshaledLbConfig, err := json.Marshal(lbConfig)
					if err != nil {
						panic(err.Error())
					}

					// Create Kubernetes Deployments extension
					loadBalancersDBExtension := codeamp_resolvers.Extension{}
					if codeampDB.Debug().Where("environment_id = ? and key = ?", env.Model.ID, "kubernetesloadbalancers").Find(&loadBalancersDBExtension).RecordNotFound() {
						panic(err.Error())
					}

					lbProjectExtension := codeamp_resolvers.ProjectExtension{
						ProjectID:     codeampProject.Model.ID,
						ExtensionID:   loadBalancersDBExtension.Model.ID,
						State:         codeamp_plugins.GetState("failed"),
						StateMessage:  "Migrated, click update to send an event.",
						Config:        postgres.Jsonb{marshaledLbConfig},
						CustomConfig:  postgres.Jsonb{marshaledLbCustomConfig},
						EnvironmentID: env.Model.ID,
					}
					codeampDB.Debug().Where(`project_id = ? and environment_id = ? and custom_config ->> 'type' = ? and config @> ?`,
						codeampProject.Model.ID,
						env.Model.ID,
						codeflowLoadBalancer.Type,
						fmt.Sprintf(`[{"key": "codeflow_id", "value": "%s" }]`, codeflowService.Id.Hex())).Assign(lbProjectExtension).FirstOrCreate(&lbProjectExtension)

					serviceName := fmt.Sprintf("%s-%s", codeflowService.Name, lbProjectExtension.Model.ID.String()[0:5])
					spew.Dump(lbProjectExtension)
					lbArtifacts := []map[string]interface{}{
						map[string]interface{}{
							"key":    "dns",
							"value":  "",
							"secret": false,
						},
						map[string]interface{}{
							"key":    "name",
							"value":  serviceName,
							"secret": false,
						},
						map[string]interface{}{
							"key":    "old_id",
							"value":  codeflowService.Id.Hex(),
							"secret": false,
						},
					}
					marshaledLbArtifacts, err := json.Marshal(lbArtifacts)
					if err != nil {
						panic(err.Error())
					}
					lbProjectExtension.Artifacts = postgres.Jsonb{marshaledLbArtifacts}
					codeampDB.Save(&lbProjectExtension)

					// add to newServiceNames
					newServiceNames = append(newServiceNames, ServiceNameMapping{
						OldServiceName: codeflowLoadBalancer.Name,
						NewServiceName: serviceName,
					})

					if codeflowLoadBalancer.Type != "internal" {
						route53CustomConfig := map[string]interface{}{
							"subdomain":         codeflowLoadBalancer.Subdomain,
							"loadbalancer":      lbProjectExtension.Model.ID.String(),
							"loadbalancer_fqdn": "",
							"loadbalancer_type": codeflowLoadBalancer.Type,
						}
						marshaledRoute53CustomConfig, err := json.Marshal(route53CustomConfig)
						if err != nil {
							panic(err.Error())
						}

						route53Artifacts := []map[string]interface{}{
							map[string]interface{}{
								"key":    "fqdn",
								"value":  fmt.Sprintf("%s.%s", codeflowLoadBalancer.Subdomain, codeflowLoadBalancer.FQDN),
								"secret": "false",
							},
							map[string]interface{}{
								"key":    "lb_fqdn",
								"value":  "",
								"secret": "false",
							},
						}
						marshaledRoute53Artifacts, err := json.Marshal(route53Artifacts)
						if err != nil {
							panic(err.Error())
						}

						// Create Kubernetes Deployments extension
						route53DBExtension := codeamp_resolvers.Extension{}
						if codeampDB.Debug().Where("environment_id = ? and key = ?", env.Model.ID, "route53").Find(&route53DBExtension).RecordNotFound() {
							panic(err.Error())
						}

						r53ProjectExtension := codeamp_resolvers.ProjectExtension{
							ProjectID:     codeampProject.Model.ID,
							ExtensionID:   route53DBExtension.Model.ID,
							State:         codeamp_plugins.GetState("failed"),
							StateMessage:  "Migrated, click update to send an event.",
							Artifacts:     postgres.Jsonb{marshaledRoute53Artifacts},
							Config:        postgres.Jsonb{[]byte("[]")},
							CustomConfig:  postgres.Jsonb{marshaledRoute53CustomConfig},
							EnvironmentID: env.Model.ID,
						}
						codeampDB.Debug().Where("project_id = ? and environment_id = ? and extension_id = ? and custom_config ->> 'subdomain' = ?",
							codeampProject.Model.ID,
							env.Model.ID,
							route53DBExtension.Model.ID,
							codeflowLoadBalancer.Subdomain).Assign(r53ProjectExtension).FirstOrCreate(&r53ProjectExtension)
					}
				}

				// Create GithubStatus extension for Continuous Integration
				if project.ContinuousIntegration {
					githubStatusDBExtension := codeamp_resolvers.Extension{}
					if codeampDB.Debug().Where("environment_id = ? and key = ?", env.Model.ID, "githubstatus").Find(&githubStatusDBExtension).RecordNotFound() {
						panic(err.Error())
					}

					githubStatusProjectExtension := codeamp_resolvers.ProjectExtension{
						ExtensionID:   githubStatusDBExtension.Model.ID,
						State:         codeamp_plugins.GetState("complete"),
						StateMessage:  "Migrated, click update to send an event.",
						Config:        postgres.Jsonb{[]byte(`[]`)},
						CustomConfig:  postgres.Jsonb{[]byte(`{}`)},
						EnvironmentID: env.Model.ID,
					}
					codeampDB.Debug().Where(`project_id = ? and environment_id = ? and extension_id = ?`,
						codeampProject.Model.ID,
						env.Model.ID,
						githubStatusDBExtension.Model.ID).Assign(githubStatusProjectExtension).FirstOrCreate(&githubStatusProjectExtension)
				}

				// Create Kubernetes Deployments extension
				kubernetesDeploymentsDBExtension := codeamp_resolvers.Extension{}
				if codeampDB.Debug().Where("environment_id = ? and key = ?", env.Model.ID, "kubernetesdeployments").Find(&kubernetesDeploymentsDBExtension).RecordNotFound() {
					panic(err.Error())
				}

				kubernetesProjectExtension := codeamp_resolvers.ProjectExtension{
					ProjectID:     codeampProject.Model.ID,
					ExtensionID:   kubernetesDeploymentsDBExtension.Model.ID,
					State:         codeamp_plugins.GetState("complete"),
					StateMessage:  "Migrated, click update to send an event.",
					Artifacts:     postgres.Jsonb{[]byte("[]")},
					Config:        postgres.Jsonb{[]byte("[]")},
					CustomConfig:  postgres.Jsonb{[]byte("{}")},
					EnvironmentID: env.Model.ID,
				}
				codeampDB.Debug().Where(codeamp_resolvers.ProjectExtension{
					ProjectID:     codeampProject.Model.ID,
					ExtensionID:   kubernetesDeploymentsDBExtension.Model.ID,
					EnvironmentID: env.Model.ID,
				}).Assign(kubernetesProjectExtension).FirstOrCreate(&kubernetesProjectExtension)

				fmt.Println("[+] Successfully created project extensions\n\n")

				fmt.Println("[*] Porting Release...")
				// find and transform the most recent release tied to the project

				// marshaledCodeampServices, err := json.Marshal(codeampServices)
				// if err != nil {
				// 	panic(err.Error())
				// }

				// marshaledCodeampSecrets, err := json.Marshal(codeampSecrets)
				// if err != nil {
				// 	panic(err.Error())
				// }

				codeampRelease := codeamp_resolvers.Release{
					ProjectID:     codeampProject.Model.ID,
					EnvironmentID: env.Model.ID,
					UserID:        codeampUser.Model.ID,
					State:         codeamp_plugins.GetState("complete"),
					StateMessage:  "migrated",
					// Services: postgres.Jsonb{marshaledCodeampServices},
					// Secrets: postgres.Jsonb{marshaledCodeampSecrets},
				}

				for {
					if codeflowDB.Session.Ping() == nil {
						break
					}
					codeflowDB.Session.Refresh()
					time.Sleep(1)
				}

				releasesResults := codeflowDB.Collection("releases").Find(bson.M{"projectId": bson.ObjectId(project.Id), "state": "complete"})
				latestCodeflowRelease := codeflow.Release{}

				releasesResults.Query.Sort("-$natural").One(&latestCodeflowRelease)

				// for releasesResults.Next(&codeflowRelease) {
				// 	if string(codeflowRelease.State) == "complete" && latestCodeflowRelease.Created.Unix() < codeflowRelease.Created.Unix() {
				// 		latestCodeflowRelease = codeflowRelease
				// 	}
				// }

				// spew.Dump(latestCodeflowRelease.State, latestCodeflowRelease.Id.Hex())
				if string(latestCodeflowRelease.State) == "" || latestCodeflowRelease.Id.Hex() == "" {
					continue
				}

				if latestCodeflowRelease.Id.String() != "" {
					fmt.Println("[+] Found latest release! ", latestCodeflowRelease.Id, latestCodeflowRelease.HeadFeatureId, latestCodeflowRelease.TailFeatureId)
					// head feature
					codeflowReleaseHeadFeature := codeflow.Feature{}

					featureResults := codeflowDB.Collection("features").Find(bson.M{"_id": bson.ObjectId(latestCodeflowRelease.HeadFeatureId)})
					featureResults.Query.One(&codeflowReleaseHeadFeature)

					fmt.Println(codeflowReleaseHeadFeature.Message)

					codeampHeadFeature := codeamp_resolvers.Feature{
						ProjectID:  codeampProject.Model.ID,
						Message:    codeflowReleaseHeadFeature.Message,
						User:       codeflowReleaseHeadFeature.User,
						Ref:        codeflowReleaseHeadFeature.Ref,
						ParentHash: codeflowReleaseHeadFeature.ParentHash,
						Created:    codeflowReleaseHeadFeature.Created,
						Hash:       codeflowReleaseHeadFeature.Hash,
					}
					codeampDB.Debug().Where(codeamp_resolvers.Feature{
						Hash: codeflowReleaseHeadFeature.Hash,
					}).FirstOrCreate(&codeampHeadFeature)

					codeampRelease.HeadFeatureID = codeampHeadFeature.Model.ID

					if latestCodeflowRelease.TailFeatureId != latestCodeflowRelease.HeadFeatureId {
						// tail feature
						codeflowReleaseTailFeature := codeflow.Feature{}
						tailFeatureResults := codeflowDB.Collection("features").Find(bson.M{"_id": bson.ObjectId(latestCodeflowRelease.TailFeatureId)})
						tailFeatureResults.Query.One(&codeflowReleaseTailFeature)

						if codeflowReleaseTailFeature.Message == "" {
							// spew.Dump(codeflowReleaseTailFeature)
							continue
						}

						fmt.Println(codeflowReleaseTailFeature.Message)
						codeampTailFeature := codeamp_resolvers.Feature{
							ProjectID:  codeampProject.Model.ID,
							Message:    codeflowReleaseTailFeature.Message,
							User:       codeflowReleaseTailFeature.User,
							Ref:        codeflowReleaseTailFeature.Ref,
							ParentHash: codeflowReleaseTailFeature.ParentHash,
							Created:    codeflowReleaseTailFeature.Created,
							Hash:       codeflowReleaseTailFeature.Hash,
						}
						codeampDB.Debug().Where(codeamp_resolvers.Feature{
							Hash: codeflowReleaseTailFeature.Hash,
						}).FirstOrCreate(&codeampTailFeature)
						codeampRelease.TailFeatureID = codeampTailFeature.Model.ID
					} else {
						codeampRelease.TailFeatureID = codeampHeadFeature.Model.ID
					}

					codeampDB.Debug().Where(codeamp_resolvers.Release{
						ProjectID:     codeampRelease.ProjectID,
						HeadFeatureID: codeampRelease.HeadFeatureID,
						TailFeatureID: codeampRelease.TailFeatureID,
						EnvironmentID: codeampRelease.EnvironmentID,
					}).FirstOrCreate(&codeampRelease)

					fmt.Println("[+] Successfully ported release \n")
				} else {
					fmt.Println("[.] No releases found.")
				}

				fmt.Println(fmt.Sprintf("Done filling objects in env %s", env.Key))
			}

			fmt.Println(fmt.Sprintf("[+] Successfully *fully* created %s for envs %s \n\n ----------------------- \n\n\n", project.Slug, envs))
		}
		fmt.Println("[+] Finished porting all projects!")

		wg.Wait()

		fmt.Println("[*] Finding all secrets that use old load balancer names and changing to the new ones")

		for _, svcNameMapping := range newServiceNames {
			var secretValues []codeamp_resolvers.SecretValue
			var services []codeamp_resolvers.Service

			// query codeamp for OldServiceName
			codeampDB.Where("value like ?", fmt.Sprintf("%%%s%%", svcNameMapping.OldServiceName)).Find(&secretValues)

			for _, secret := range secretValues {
				result := strings.Replace(secret.Value, svcNameMapping.OldServiceName, svcNameMapping.NewServiceName, 1)
				spew.Dump(fmt.Sprintf("%s -> %s", svcNameMapping.OldServiceName, svcNameMapping.NewServiceName), result)
				newSecret := codeamp_resolvers.SecretValue{
					SecretID: secret.SecretID,
					Value:    result,
					UserID:   secret.UserID,
				}
				codeampDB.Create(&newSecret)
			}

			codeampDB.Where("command like ?", fmt.Sprintf("%%%s.%%", svcNameMapping.OldServiceName)).Find(&services)
			for _, svc := range services {
				result := strings.Replace(svc.Command, svcNameMapping.OldServiceName, svcNameMapping.NewServiceName, 1)
				svc.Command = result
				codeampDB.Save(&svc)
				spew.Dump(fmt.Sprintf("svc %s command changed: %s -> %s", svc.Name, svcNameMapping.OldServiceName, svcNameMapping.NewServiceName), result)
			}
		}

		fmt.Println("[+] Finished replacing old service names with new in all relevant secrets")
	},
}

func createCodeflowDBConnection() {
	var err error
	spew.Dump(viper.GetStringMap("codeflow"))
	config := &bongo.Config{
		ConnectionString: viper.GetString("codeflow.mongodb.uri"),
		Database:         viper.GetString("codeflow.mongodb.database"),
	}

	if viper.GetBool("codeflow.mongodb.ssl") {
		if config.DialInfo, err = mgo.ParseURL(config.ConnectionString); err != nil {
			panic(fmt.Sprintf("cannot parse given URI %s due to error: %s", config.ConnectionString, err.Error()))
		}

		tlsConfig := &tls.Config{}
		config.DialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
			return conn, err
		}

		config.DialInfo.Timeout = time.Second * viper.GetDuration("codeflow.mongodb.connection_timeout")
	}

	codeflowDB, err = bongo.Connect(config)
	if err != nil {
		log.Fatal(err)
	}

	// Try to reconnect if connection drops
	go func(session *mgo.Session) {
		var err error
		for {
			err = session.Ping()
			if err != nil {
				fmt.Println("Lost connection to MongoDB!!")
				session.Refresh()
				err = session.Ping()
				if err == nil {
					fmt.Println("Reconnect to MongoDB successful.")
				} else {
					panic("Reconnect to MongoDB failed!!")
				}
			}
			time.Sleep(time.Second * viper.GetDuration("codeflow.mongodb.health_check_interval"))
		}
	}(codeflowDB.Session)
}

func createCodeampDB() (resolver *gorm.DB, err error) {
	db, err := codeamp.NewDB(viper.GetString("codeamp.postgres.host"), viper.GetString("codeamp.postgres.port"), viper.GetString("codeamp.postgres.user"), viper.GetString("codeamp.postgres.dbname"), viper.GetString("codeamp.postgres.sslmode"), viper.GetString("codeamp.postgres.password"))
	if err != nil {
		return nil, err
	}
	return db, nil
}

func insertAllowOverrideAttributeIntoExtConfig(extension codeamp_resolvers.Extension) ([]byte, error) {
	var err error
	type ExtConfig struct {
		Key           string `json:"key"`
		Value         string `json:"value"`
		AllowOverride bool   `json:"allowOverride"`
	}

	// unmarshal config and add AllowOverride to false
	extensionConfig := []ExtConfig{}
	newExtensionConfig := []ExtConfig{}

	err = json.Unmarshal(extension.Config.RawMessage, &extensionConfig)
	if err != nil {
		return nil, err
	}

	for _, kv := range extensionConfig {
		kv.AllowOverride = false
		newExtensionConfig = append(newExtensionConfig, kv)
	}

	marshaledNewExtensionConfig, err := json.Marshal(newExtensionConfig)
	if err != nil {
		return nil, err
	}

	return marshaledNewExtensionConfig, nil
}

func init() {
	RootCmd.AddCommand(cfMigrateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// migrateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
