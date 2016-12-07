package haproxy_plugin_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/enaml-ops/enaml"
	"github.com/enaml-ops/haproxy-plugin/haproxy/enaml-gen/haproxy"
	. "github.com/enaml-ops/haproxy-plugin/haproxy/plugin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("haproxy plugin", func() {
	var hplugin *Plugin

	Context("When flags are read from environment variables", func() {
		var controlAZName = "blah"
		var manifestBytes []byte
		var err error

		BeforeEach(func() {
			hplugin = &Plugin{Version: "0.0"}
			os.Setenv("OMG_AZ", controlAZName)

			manifestBytes, err = hplugin.GetProduct([]string{
				"haproxy-command",
				"--cert-filepath", "fixtures/pem1.pem",
				"--haproxy-ip", "asdfasdf",
				"--network-name", "net1",
				"--gorouter-ip", "1.2.3.4",
			}, []byte{}, nil)
		})
		AfterEach(func() {
			os.Setenv("OMG_AZ", "")
		})

		It("should pass validation of required flags", func() {
			Expect(err).ShouldNot(HaveOccurred(), "should pass env var isset required value check")
			Expect(hplugin.AZs).Should(ConsistOf(controlAZName))
		})

		It("should properly set up the Availability Zones", func() {
			Expect(err).ShouldNot(HaveOccurred())
			manifest := enaml.NewDeploymentManifest(manifestBytes)
			for _, instanceGroup := range manifest.InstanceGroups {
				Expect(instanceGroup.AZs).Should(Equal([]string{controlAZName}), fmt.Sprintf("Availability ZOnes for instance group %v was not set properly", instanceGroup.Name))
			}
		})
	})

	Context("When a commnd line args are passed", func() {
		var haproxyInstanceGroup *enaml.InstanceGroup
		var haproxyJobProperties = new(haproxy.HaproxyJob)
		var controlHaProxyIP = "1.1.1.1"
		var controlNetworkName = "net1"
		var controlSyslogURL = "1.2.3.4:888"
		var controlBackendIPs = []string{
			"10.0.0.20",
			"10.0.0.21",
		}
		var controlDomains = []string{
			"blah.domain.io",
			"bleh.dooda.io",
		}
		var controlCIDRs = []string{
			"0.0.0.0/24",
			"1.1.1.1/24",
		}
		BeforeEach(func() {
			var haproxyPropBytes []byte
			hplugin = &Plugin{Version: "0.0"}
			manifestBytes, err := hplugin.GetProduct([]string{
				"haproxy-command",
				"--syslog-url", controlSyslogURL,
				"--cert-filepath", "fixtures/pem1.pem",
				"--haproxy-ip", controlHaProxyIP,
				"--az", "z1",
				"--network-name", controlNetworkName,
				"--stemcell-alias", "trusty",
				"--gorouter-ip", controlBackendIPs[0],
				"--gorouter-ip", controlBackendIPs[1],
				"--internal-only-domain", controlDomains[0],
				"--internal-only-domain", controlDomains[1],
				"--trusted-domain-cidr", controlCIDRs[0],
				"--trusted-domain-cidr", controlCIDRs[1],
			}, []byte{}, nil)
			Expect(err).ShouldNot(HaveOccurred())
			manifest := enaml.NewDeploymentManifest(manifestBytes)
			haproxyInstanceGroup = manifest.GetInstanceGroupByName(DefaultInstanceGroupName)
			haproxyPropBytes, err = yaml.Marshal(haproxyInstanceGroup.GetJobByName(DefaultJobName).Properties)
			Ω(err).ShouldNot(HaveOccurred())
			err = yaml.Unmarshal(haproxyPropBytes, haproxyJobProperties)
			Ω(err).ShouldNot(HaveOccurred())
		})

		Describe("internal only domains", func() {
			Context("when called with a list of internal only domains", func() {

				It("should configure the deployment with internal only domains", func() {
					Ω(haproxyJobProperties.HaProxy.InternalOnlyDomains).Should(HaveLen(len(controlDomains)))
					Ω(haproxyJobProperties.HaProxy.InternalOnlyDomains).Should(ConsistOf(controlDomains))
				})
			})
		})

		Describe("trusted domain cidrs", func() {
			Context("when called with a list of trusted domain cidrs", func() {
				It("should configure the deployment with trusted domain cidrs", func() {
					listOfCIDRs := strings.Split(haproxyJobProperties.HaProxy.TrustedDomainCidrs.(string), " ")
					Ω(listOfCIDRs).Should(ConsistOf(controlCIDRs))
				})
			})
		})

		Describe("syslog", func() {
			Context("when passed a syslog url", func() {
				It("should add the value to the job properties", func() {
					Ω(haproxyJobProperties.HaProxy.SyslogServer).Should(Equal(controlSyslogURL))
				})
			})
			Context("when NOT passed a syslog url", func() {
				BeforeEach(func() {
					var haproxyPropBytes []byte
					hplugin = &Plugin{Version: "0.0"}
					manifestBytes, err := hplugin.GetProduct([]string{
						"haproxy-command",
						"--cert-filepath", "fixtures/pem1.pem",
						"--haproxy-ip", controlHaProxyIP,
						"--az", "z1",
						"--network-name", controlNetworkName,
						"--stemcell-alias", "trusty",
						"--gorouter-ip", controlBackendIPs[0],
						"--gorouter-ip", controlBackendIPs[1],
					}, []byte{}, nil)
					Expect(err).ShouldNot(HaveOccurred())
					manifest := enaml.NewDeploymentManifest(manifestBytes)
					haproxyInstanceGroup = manifest.GetInstanceGroupByName(DefaultInstanceGroupName)
					haproxyPropBytes, err = yaml.Marshal(haproxyInstanceGroup.GetJobByName(DefaultJobName).Properties)
					Ω(err).ShouldNot(HaveOccurred())
					err = yaml.Unmarshal(haproxyPropBytes, haproxyJobProperties)
					Ω(err).ShouldNot(HaveOccurred())
				})
				It("should add the value to the job properties", func() {
					Ω(haproxyJobProperties.HaProxy.SyslogServer.(string)).Should(BeEmpty())
				})
			})
		})

		Describe("ssl_pem flag", func() {
			Context("when given a single pem file path", func() {
				It("then it should define a ssl_pem for the given file", func() {
					var controlPEM string
					pemBytes, _ := ioutil.ReadFile("fixtures/pem1.pem")
					controlPEM = string(pemBytes)
					sslPemRecord := haproxyJobProperties.HaProxy.SslPem.([]interface{})
					Ω(len(sslPemRecord)).Should(Equal(1), "for only one file given should create only a single record")
					Ω(sslPemRecord[0]).Should(Equal(controlPEM))
				})
			})

			Context("when given multiple pem file paths", func() {
				BeforeEach(func() {
					var haproxyPropBytes []byte
					hplugin = &Plugin{Version: "0.0"}
					manifestBytes, err := hplugin.GetProduct([]string{
						"haproxy-command",
						"--cert-filepath", "fixtures/pem1.pem",
						"--cert-filepath", "fixtures/pem2.pem",
						"--haproxy-ip", controlHaProxyIP,
						"--az", "z1",
						"--network-name", controlNetworkName,
						"--stemcell-alias", "trusty",
						"--gorouter-ip", controlBackendIPs[0],
						"--gorouter-ip", controlBackendIPs[1],
					}, []byte{}, nil)
					Expect(err).ShouldNot(HaveOccurred())
					manifest := enaml.NewDeploymentManifest(manifestBytes)
					haproxyInstanceGroup = manifest.GetInstanceGroupByName(DefaultInstanceGroupName)
					haproxyPropBytes, err = yaml.Marshal(haproxyInstanceGroup.GetJobByName(DefaultJobName).Properties)
					Ω(err).ShouldNot(HaveOccurred())
					err = yaml.Unmarshal(haproxyPropBytes, haproxyJobProperties)
					Ω(err).ShouldNot(HaveOccurred())
				})
				It("then it should define a ssl_pem for the given file", func() {
					var controlPEM1 string
					var controlPEM2 string
					pemBytes, _ := ioutil.ReadFile("fixtures/pem1.pem")
					controlPEM1 = string(pemBytes)
					pemBytes, _ = ioutil.ReadFile("fixtures/pem2.pem")
					controlPEM2 = string(pemBytes)
					sslPemRecord := haproxyJobProperties.HaProxy.SslPem.([]interface{})
					Ω(len(sslPemRecord)).Should(Equal(2), "we should have a pem for each file given as an argument")
					Ω(sslPemRecord).Should(ConsistOf(controlPEM1, controlPEM2))
				})
			})
		})

		Context("when given a ip for haproxy", func() {
			It("should have a network name matching the given flag", func() {
				Ω(haproxyInstanceGroup.Networks[0].Name).Should(Equal(controlNetworkName))
			})

			It("should properly set the static ip for the haproxy vm instance", func() {
				Ω(haproxyInstanceGroup.Instances).Should(Equal(DefaultHaProxyInstanceCount), "we only ever want one haproxy VM for now")
				Ω(len(haproxyInstanceGroup.Networks)).Should(BeNumerically(">", 0))
				Ω(len(haproxyInstanceGroup.Networks[0].StaticIPs)).Should(Equal(DefaultHaProxyInstanceCount), "we only ever want one haproxy VM for now")
				Ω(haproxyInstanceGroup.Networks[0].StaticIPs).Should(ConsistOf(controlHaProxyIP))
			})
		})

		Context("when given config values for backend server info (go router ips)", func() {
			It("then it should define a list of backend server IPs", func() {
				Ω(haproxyJobProperties.HaProxy.BackendServers).Should(ConsistOf(controlBackendIPs))
			})
		})

		It("should return error when AZ/s are not provided", func() {
			_, err := hplugin.GetProduct([]string{
				"haproxy-command",
				"--network-name", "asdf",
			}, []byte{}, nil)
			Expect(err).Should(HaveOccurred())
		})

		It("should return error when network name is not provided", func() {
			_, err := hplugin.GetProduct([]string{
				"haproxy-command",
				"--az", "asdf",
			}, []byte{}, nil)
			Expect(err).Should(HaveOccurred())
		})

		Context("when passed valid flags", func() {

			var manifestBytes []byte
			var controlStemcellAlias = "ubuntu-magic"
			var controlStemcellURL = "ubuntu-URL"
			var controlStemcellSHA = "ubuntu-SHA"
			var controlName = "haproxy-name"
			var controlAZ = "z1"
			var controlGoRouterIP = "1.2.3.4"

			BeforeEach(func() {
				var err error
				manifestBytes, err = hplugin.GetProduct([]string{
					"haproxy-command",
					"--haproxy-ip", controlHaProxyIP,
					"--az", controlAZ,
					"--network-name", "net1",
					"--cert-filepath", "fixtures/pem1.pem",
					"--deployment-name", controlName,
					"--stemcell-alias", controlStemcellAlias,
					"--stemcell-url", controlStemcellURL,
					"--stemcell-sha", controlStemcellSHA,
					"--gorouter-ip", controlGoRouterIP,
				}, []byte{}, nil)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("should set a valid release on the instance groups' jobs", func() {
				manifest := enaml.NewDeploymentManifest(manifestBytes)
				for _, instanceGroup := range manifest.InstanceGroups {
					for _, job := range instanceGroup.Jobs {
						Expect(job.Release).ShouldNot(BeEmpty())
					}
				}
			})

			It("should properly set up the Update segment", func() {
				manifest := enaml.NewDeploymentManifest(manifestBytes)
				Ω(manifest.Update.MaxInFlight).ShouldNot(Equal(0), "max in flight")
				Ω(manifest.Update.UpdateWatchTime).ShouldNot(BeEmpty(), "update watch time")
				Ω(manifest.Update.CanaryWatchTime).ShouldNot(BeEmpty(), "canary watch time")
				Ω(manifest.Update.Canaries).ShouldNot(Equal(0), "number of canaries")
			})

			It("should properly set up the haproxy release", func() {
				manifest := enaml.NewDeploymentManifest(manifestBytes)
				Ω(manifest.Releases).ShouldNot(BeEmpty())
				Ω(manifest.Releases[0]).ShouldNot(BeNil())
				Ω(manifest.Releases[0].Name).Should(Equal("haproxy"))
				Ω(manifest.Releases[0].Version).Should(Equal("latest"))
			})

			It("should properly set up the stemcells", func() {
				manifest := enaml.NewDeploymentManifest(manifestBytes)
				Ω(manifest.Stemcells).ShouldNot(BeNil())
				Ω(manifest.Stemcells[0].Alias).Should(Equal(controlStemcellAlias))
				Ω(manifest.Stemcells[0].URL).Should(Equal(controlStemcellURL))
				Ω(manifest.Stemcells[0].SHA1).Should(Equal(controlStemcellSHA))
				for _, instanceGroup := range manifest.InstanceGroups {
					Expect(instanceGroup.Stemcell).Should(Equal(controlStemcellAlias), fmt.Sprintf("stemcell for instance group %v was not set properly", instanceGroup.Name))
				}
			})

			It("should properly set up the deployment name", func() {
				manifest := enaml.NewDeploymentManifest(manifestBytes)
				Ω(manifest.Name).Should(Equal(controlName))
			})

			It("should properly set up the Availability Zones", func() {
				manifest := enaml.NewDeploymentManifest(manifestBytes)
				for _, instanceGroup := range manifest.InstanceGroups {
					Expect(instanceGroup.AZs).Should(Equal([]string{controlAZ}), fmt.Sprintf("Availability ZOnes for instance group %v was not set properly", instanceGroup.Name))
				}
			})
		})
	})
})
