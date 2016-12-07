package haproxy_plugin

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/enaml-ops/enaml"
	"github.com/enaml-ops/haproxy-plugin/haproxy/enaml-gen/haproxy"
	"github.com/enaml-ops/pluginlib/cred"
	"github.com/enaml-ops/pluginlib/pcli"
	"github.com/enaml-ops/pluginlib/pluginutil"
	"github.com/enaml-ops/pluginlib/productv1"
	"github.com/xchapter7x/lo"
)

type Plugin struct {
	Version string `omg:"-"`

	DeploymentName      string   `omg:"deployment-name"`
	NetworkName         string   `omg:"network-name"`
	HaproxyReleaseVer   string   `omg:"haproxy-release-ver"`
	HaproxyReleaseURL   string   `omg:"haproxy-release-url,optional"`
	HaproxyReleaseSHA   string   `omg:"haproxy-release-sha,optional"`
	StemcellName        string   `omg:"stemcell-name"`
	StemcellVer         string   `omg:"stemcell-ver"`
	StemcellAlias       string   `omg:"stemcell-alias"`
	StemcellURL         string   `omg:"stemcell-url,optional"`
	StemcellSHA         string   `omg:"stemcell-sha,optional"`
	AZs                 []string `omg:"az"`
	GoRouterIPs         []string `omg:"gorouter-ip"`
	HaProxyIP           string   `omg:"haproxy-ip"`
	PEMFiles            []string `omg:"cert-filepath"`
	SyslogURL           string   `omg:"syslog-url,optional"`
	InternalOnlyDomains []string `omg:"internal-only-domain,optional"`
	TrustedDomainCidrs  []string `omg:"trusted-domain-cidr,optional"`
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
	deploymentManifest.AddRelease(enaml.Release{
		Name:    releaseName,
		Version: p.HaproxyReleaseVer,
		URL:     p.HaproxyReleaseURL,
		SHA1:    p.HaproxyReleaseSHA,
	})
	deploymentManifest.AddStemcell(enaml.Stemcell{
		OS:      p.StemcellName,
		Version: p.StemcellVer,
		Alias:   p.StemcellAlias,
		URL:     p.StemcellURL,
		SHA1:    p.StemcellSHA,
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
		Instances: DefaultHaProxyInstanceCount,
		Name:      DefaultInstanceGroupName,
		AZs:       p.AZs,
		Stemcell:  p.StemcellAlias,
		Jobs:      p.newJobs(),
		Networks:  p.newNetworks(),
	}
	return ig
}

func (p *Plugin) newNetworks() []enaml.Network {
	var nets []enaml.Network
	nets = append(nets, enaml.Network{
		Name: p.NetworkName,
		StaticIPs: []string{
			p.HaProxyIP,
		},
	})
	return nets
}

func (p *Plugin) newPEMs() []string {
	var pems []string

	for _, pempath := range p.PEMFiles {
		pem, err := ioutil.ReadFile(pempath)

		if err != nil {
			lo.G.Errorf("cant read pem file!!!, @ '%v' ", pempath)
			break
		}
		pems = append(pems, string(pem))
	}
	return pems
}

func (p *Plugin) newJobs() []enaml.InstanceJob {
	jobs := []enaml.InstanceJob{
		enaml.InstanceJob{
			Release: releaseName,
			Name:    DefaultJobName,
			Properties: &haproxy.HaproxyJob{
				HaProxy: &haproxy.HaProxy{
					BackendServers:      p.GoRouterIPs,
					SslPem:              p.newPEMs(),
					SyslogServer:        p.SyslogURL,
					TrustedDomainCidrs:  strings.Join(p.TrustedDomainCidrs, " "),
					InternalOnlyDomains: p.InternalOnlyDomains,
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
				URL:     DefaultReleaseURL,
				SHA1:    DefaultReleaseSHA,
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
			Usage:    "the version of the stemcell you with to use",
		},
		pcli.Flag{
			FlagType: pcli.StringFlag,
			Name:     "stemcell-url",
			Usage:    "the url of the stemcell you with to use (this is optional: it will use a stemcell that already exists in bosh by default)",
		},
		pcli.Flag{
			FlagType: pcli.StringFlag,
			Name:     "stemcell-sha",
			Usage:    "the sha of the stemcell you with to use (if you're giving a optional stemcell URL)",
		},
		pcli.Flag{
			FlagType: pcli.StringFlag,
			Name:     "haproxy-release-ver",
			Value:    releaseVersion,
			Usage:    "the version of the release to use for the deployment",
		},
		pcli.Flag{
			FlagType: pcli.StringFlag,
			Name:     "haproxy-release-url",
			Value:    DefaultReleaseURL,
			Usage:    "the URL of the release to use for the deployment",
		},
		pcli.Flag{
			FlagType: pcli.StringFlag,
			Name:     "haproxy-release-sha",
			Value:    DefaultReleaseSHA,
			Usage:    "the SHA of the release to use for the deployment",
		},
		pcli.Flag{
			FlagType: pcli.StringSliceFlag,
			Name:     "gorouter-ip",
			Usage:    "gorouter ips (give flag multiple times for multiple IPs)",
		},
		pcli.Flag{
			FlagType: pcli.StringFlag,
			Name:     "haproxy-ip",
			Usage:    "ip for haproxy vm to listen on",
		},
		pcli.Flag{
			FlagType: pcli.StringSliceFlag,
			Name:     "cert-filepath",
			Usage:    "the path to your pem file containing entire chain (give multiple flags to use multiple pems)",
		},
		pcli.Flag{
			FlagType: pcli.StringFlag,
			Name:     "syslog-url",
			Usage:    "url for the optionally targetted syslog drain",
		},
		pcli.Flag{
			FlagType: pcli.StringSliceFlag,
			Name:     "internal-only-domain",
			Usage:    "domains for internal-only apps/services - not hostnames for the apps/services (give multiple flags to use multiple domains)",
		},
		pcli.Flag{
			FlagType: pcli.StringSliceFlag,
			Name:     "trusted-domain-cidr",
			Usage:    "trusted domain cidrs to be used with internal only domains (give multiple flags to use multiple cidrs)",
		},
	}
}
