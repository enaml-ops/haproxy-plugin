package haproxy_plugin_test

import (
	"fmt"
	"os"

	"github.com/enaml-ops/enaml"
	. "github.com/enaml-ops/haproxy-plugin/haproxy/plugin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("haproxy plugin", func() {
	var gPlugin *Plugin

	Context("When flags are read from environment variables", func() {
		var controlAZName = "blah"
		BeforeEach(func() {
			gPlugin = &Plugin{Version: "0.0"}
			os.Setenv("OMG_AZ", controlAZName)
		})
		AfterEach(func() {
			os.Setenv("OMG_AZ", "")
		})

		It("should pass validation of required flags", func() {
			_, err := gPlugin.GetProduct([]string{
				"haproxy-command",
				"--network-name", "net1",
			}, []byte{}, nil)
			Expect(err).ShouldNot(HaveOccurred(), "should pass env var isset required value check")
			Expect(gPlugin.AZs).Should(ConsistOf(controlAZName))
		})

		It("should properly set up the Availability Zones", func() {
			manifestBytes, err := gPlugin.GetProduct([]string{
				"haproxy-command",
				"--network-name", "net1",
			}, []byte{}, nil)
			Expect(err).ShouldNot(HaveOccurred())
			manifest := enaml.NewDeploymentManifest(manifestBytes)
			for _, instanceGroup := range manifest.InstanceGroups {
				Expect(instanceGroup.AZs).Should(Equal([]string{controlAZName}), fmt.Sprintf("Availability ZOnes for instance group %v was not set properly", instanceGroup.Name))
			}
		})
	})

	Context("When a commnd line args are passed", func() {
		BeforeEach(func() {
			gPlugin = &Plugin{Version: "0.0"}
		})

		It("should return error when AZ/s are not provided", func() {
			_, err := gPlugin.GetProduct([]string{
				"haproxy-command",
				"--network-name", "asdf",
			}, []byte{}, nil)
			Expect(err).Should(HaveOccurred())
		})

		It("should return error when network name is not provided", func() {
			_, err := gPlugin.GetProduct([]string{
				"haproxy-command",
				"--az", "asdf",
			}, []byte{}, nil)
			Expect(err).Should(HaveOccurred())
		})

		It("should properly set up the Update segment", func() {
			controlStemcellAlias := "ubuntu-magic"
			manifestBytes, err := gPlugin.GetProduct([]string{
				"haproxy-command",
				"--az", "z1",
				"--network-name", "net1",
				"--stemcell-alias", controlStemcellAlias,
			}, []byte{}, nil)
			Expect(err).ShouldNot(HaveOccurred())
			manifest := enaml.NewDeploymentManifest(manifestBytes)
			Ω(manifest.Update.MaxInFlight).ShouldNot(Equal(0), "max in flight")
			Ω(manifest.Update.UpdateWatchTime).ShouldNot(BeEmpty(), "update watch time")
			Ω(manifest.Update.CanaryWatchTime).ShouldNot(BeEmpty(), "canary watch time")
			Ω(manifest.Update.Canaries).ShouldNot(Equal(0), "number of canaries")
		})

		It("should properly set up the gemfire release", func() {
			controlStemcellAlias := "ubuntu-magic"
			manifestBytes, err := gPlugin.GetProduct([]string{
				"haproxy-command",
				"--az", "z1",
				"--network-name", "net1",
				"--stemcell-alias", controlStemcellAlias,
			}, []byte{}, nil)
			Expect(err).ShouldNot(HaveOccurred())
			manifest := enaml.NewDeploymentManifest(manifestBytes)
			Ω(manifest.Releases).ShouldNot(BeEmpty())
			Ω(manifest.Releases[0]).ShouldNot(BeNil())
			Ω(manifest.Releases[0].Name).Should(Equal("haproxy"))
			Ω(manifest.Releases[0].Version).Should(Equal("latest"))
		})

		It("should properly set up the stemcells", func() {
			controlStemcellAlias := "ubuntu-magic"
			manifestBytes, err := gPlugin.GetProduct([]string{
				"haproxy-command",
				"--az", "z1",
				"--network-name", "net1",
				"--stemcell-alias", controlStemcellAlias,
			}, []byte{}, nil)
			Expect(err).ShouldNot(HaveOccurred())
			manifest := enaml.NewDeploymentManifest(manifestBytes)
			Ω(manifest.Stemcells).ShouldNot(BeNil())
			Ω(manifest.Stemcells[0].Alias).Should(Equal(controlStemcellAlias))
			for _, instanceGroup := range manifest.InstanceGroups {
				Expect(instanceGroup.Stemcell).Should(Equal(controlStemcellAlias), fmt.Sprintf("stemcell for instance group %v was not set properly", instanceGroup.Name))
			}
		})

		It("should properly set up the deployment name", func() {
			var controlName = "haproxy-name"
			manifestBytes, err := gPlugin.GetProduct([]string{
				"haproxy-command",
				"--az", "z1",
				"--deployment-name", controlName,
				"--network-name", "net1",
			}, []byte{}, nil)
			Expect(err).ShouldNot(HaveOccurred())
			manifest := enaml.NewDeploymentManifest(manifestBytes)
			Ω(manifest.Name).Should(Equal(controlName))
		})

		It("should properly set up the Availability Zones", func() {
			var controlAZ = "z1"
			manifestBytes, err := gPlugin.GetProduct([]string{
				"haproxy-command",
				"--az", controlAZ,
				"--network-name", "net1",
			}, []byte{}, nil)
			Expect(err).ShouldNot(HaveOccurred())
			manifest := enaml.NewDeploymentManifest(manifestBytes)
			for _, instanceGroup := range manifest.InstanceGroups {
				Expect(instanceGroup.AZs).Should(Equal([]string{controlAZ}), fmt.Sprintf("Availability ZOnes for instance group %v was not set properly", instanceGroup.Name))
			}
		})
	})
})
