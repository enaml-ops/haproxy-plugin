package main

import (
	"github.com/enaml-ops/haproxy-plugin/haproxy/plugin"
	"github.com/enaml-ops/pluginlib/productv1"
)

// Version is the version of the haproxy plugin.
var Version string = "v0.0.0" // overridden at link time

func main() {
	product.Run(&haproxy_plugin.Plugin{
		Version: Version,
	})
}
