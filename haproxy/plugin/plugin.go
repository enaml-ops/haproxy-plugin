package haproxy_plugin

import (
	"fmt"
	"strings"

	"github.com/enaml-ops/enaml"
	"github.com/enaml-ops/haproxy-plugin/haproxy/enaml-gen/haproxy"
	"github.com/enaml-ops/pluginlib/cred"
	"github.com/enaml-ops/pluginlib/pcli"
	"github.com/enaml-ops/pluginlib/pluginutil"
	"github.com/enaml-ops/pluginlib/productv1"
)

type Plugin struct {
	Version string `omg:"-"`

	DeploymentName    string   `omg:"deployment-name"`
	NetworkName       string   `omg:"network-name"`
	HaproxyReleaseVer string   `omg:"haproxy-release-ver"`
	StemcellName      string   `omg:"stemcell-name"`
	StemcellVer       string   `omg:"stemcell-ver"`
	StemcellAlias     string   `omg:"stemcell-alias"`
	AZs               []string `omg:"az"`
	GoRouterIPs       []string `omg:"gorouter-ip"`
}

// GetProduct generates a BOSH deployment manifest for haproxy.
func (p *Plugin) GetProduct(args []string, cloudConfig []byte, cs cred.Store) ([]byte, error) {
	c := pluginutil.NewContext(args, pluginutil.ToCliFlagArray(p.GetFlags()))
	err := pcli.UnmarshalFlags(p, c)
	if err != nil {
		return nil, err
	}
	deploymentManifest := new(enaml.DeploymentManifest)
	deploymentManifest.SetName(p.DeploymentName)
	deploymentManifest.AddRelease(enaml.Release{Name: releaseName, Version: p.HaproxyReleaseVer})
	deploymentManifest.AddStemcell(enaml.Stemcell{
		OS:      p.StemcellName,
		Version: p.StemcellVer,
		Alias:   p.StemcellAlias,
	})
	deploymentManifest.Update = enaml.Update{
		MaxInFlight:     1,
		UpdateWatchTime: "30000-300000",
		CanaryWatchTime: "30000-300000",
		Serial:          false,
		Canaries:        1,
	}
	deploymentManifest.AddInstanceGroup(p.newInstanceGroup())
	return deploymentManifest.Bytes(), nil
}

func (p *Plugin) newInstanceGroup() *enaml.InstanceGroup {
	ig := &enaml.InstanceGroup{
		Name:     DefaultInstanceGroupName,
		AZs:      p.AZs,
		Stemcell: p.StemcellAlias,
		Jobs:     p.newJobs(),
	}
	return ig
}

func (p *Plugin) newJobs() []enaml.InstanceJob {
	jobs := []enaml.InstanceJob{
		enaml.InstanceJob{
			Name: DefaultJobName,
			Properties: &haproxy.HaproxyJob{
				HaProxy: &haproxy.HaProxy{
					BackendServers: p.GoRouterIPs,
				},
			},
		},
	}
	return jobs
}

func makeEnvVarName(flagName string) string {
	return "OMG_" + strings.Replace(strings.ToUpper(flagName), "-", "_", -1)
}

// GetMeta returns metadata about the haproxy product.
func (p *Plugin) GetMeta() product.Meta {
	return product.Meta{
		Name: "haproxy",
		Stemcell: enaml.Stemcell{
			Name:    defaultStemcellName,
			Alias:   defaultStemcellAlias,
			Version: defaultStemcellVersion,
		},
		Releases: []enaml.Release{
			enaml.Release{
				Name:    releaseName,
				Version: releaseVersion,
			},
		},
		Properties: map[string]interface{}{
			"version":              p.Version,
			"stemcell":             defaultStemcellVersion,
			"pivotal-gemfire-tile": "NOT COMPATIBLE WITH TILE RELEASES",
			"haproxy":              fmt.Sprintf("%s / %s", releaseName, releaseVersion),
			"description":          "this plugin is designed to work with a special haproxy release",
		},
	}
}

// GetFlags returns the CLI flags accepted by the plugin.
func (p *Plugin) GetFlags() []pcli.Flag {
	return []pcli.Flag{
		pcli.Flag{
			FlagType: pcli.StringFlag,
			Name:     "deployment-name",
			Value:    defaultDeploymentName,
			Usage:    "the name bosh will use for this deployment",
		},
		pcli.Flag{
			FlagType: pcli.StringSliceFlag,
			Name:     "az",
			Usage:    "the list of Availability Zones where you wish to deploy gemfire",
		},
		pcli.Flag{
			FlagType: pcli.StringFlag,
			Name:     "network-name",
			Usage:    "the network where you wish to deploy locators and servers",
		},
		pcli.Flag{
			FlagType: pcli.StringFlag,
			Name:     "stemcell-name",
			Value:    p.GetMeta().Stemcell.Name,
			Usage:    "the name of the stemcell you with to use",
		},
		pcli.Flag{
			FlagType: pcli.StringFlag,
			Name:     "stemcell-alias",
			Value:    p.GetMeta().Stemcell.Alias,
			Usage:    "the name of the stemcell you with to use",
		},
		pcli.Flag{
			FlagType: pcli.StringFlag,
			Name:     "stemcell-ver",
			Value:    p.GetMeta().Stemcell.Version,
			Usage:    "the name of the stemcell you with to use",
		},
		pcli.Flag{
			FlagType: pcli.StringFlag,
			Name:     "haproxy-release-ver",
			Value:    releaseVersion,
			Usage:    "the version of the release to use for the deployment",
		},
		pcli.Flag{
			FlagType: pcli.StringSliceFlag,
			Name:     "gorouter-ip",
			Usage:    "gorouter ips (give flag multiple times for multiple IPs)",
		},
	}
}
