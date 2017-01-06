HAPROXY Enaml Plugin for OMG
## This enaml plugin provides support in the HAProxy release for multiple certs and configuration of trusted domains/cidr ranges, which the HAProxy in PCF Elastic Runtime does not provide by default. These features are useful when trying to employ a L-shaped network topology (private-only system access vs. public access to apps) in front of PCF and you would like to add SSL certificates on a per-app basis, rather than one wildcard cert for the whole domain.

### Plugin setup
```
# download the omg-cli which is used to run plugins
# please be sure to swap osx for linux if that is the system you intend to run
# this on
$ wget -O omg https://github.com/enaml-ops/omg-cli/releases/download/v1.0.3/omg-osx && chmod +x omg

# download the plugin. feel free to swap the version for your desired version
$ wget -O haproxy https://github.com/enaml-ops/haproxy-plugin/releases/download/v0.0.6/haproxy-plugin-osx 

# register the plugin in omg-cli client
$ ./omg register-plugin --type product --pluginpath haproxy

# list registered plugins to show registration was successfull
$ ./omg list-products

# output deploy-product arguments for targetting a bosh
$ ./omg deploy-product --help

# output deploy-product commands for the haproxy product
$ ./omg deploy-product haproxy --help

# to deploy haproxy or print the manifest a command like the following would be used
$ ./omg deploy-product \
   --bosh-url "https://xx.xxx.x.xx" \
   --bosh-port 25555 \
   --bosh-user director \
   --bosh-pass password \
   --ssl-ignore \
   --print-manifest \
   haproxy \ 
   --deployment-name haproxy \
   --vm-type Standard_F1s \
   --network-name ert-network \
   --stemcell-name ubuntu-trusty \
   --stemcell-alias trusty \
   --stemcell-ver 3232.17 \
   --haproxy-release-ver latest \
   --haproxy-release-url https://bosh.io/d/github.com/cloudfoundry-community/haproxy-boshrelease?v=8.0.9 \
   --haproxy-release-sha 13598c70a50f8caf95d06782d67610daede8aeb9 \
   --gorouter-ip xx.xxx.x.xx  \
   --haproxy-ip xx.xxx.x.xx \
   --haproxy-ip xx.xxx.x.xx \
   --cert-filepath certs/apps01.DOMAIN.chain.pem \
   --trusted-domain-cidr "xx.xxx.x.0/23"
```

### Notes
- using the `--print-manifest` flag will simply output the generated manifest to stdout
