/*
Copyright 2019 The Kubernetes Authors.

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

package model

import (
	"fmt"
	"strings"

	"k8s.io/klog"
	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/nodeup/pkg/model/resources"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// ContainerdBuilder install containerd (just the packages at the moment)
type ContainerdBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &ContainerdBuilder{}

var containerdVersions = []packageVersion{
	// 1.2.4 - Debian Stretch
	{
		PackageVersion: "1.2.4",
		Name:           "containerd.io",
		Distros:        []distros.Distribution{distros.DistributionDebian9},
		Architectures:  []Architecture{ArchitectureAmd64},
		Version:        "1.2.4-1",
		Source:         "https://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/containerd.io_1.2.4-1_amd64.deb",
		Hash:           "48c6ab0c908316af9a183de5aad64703bc516bdf",
	},

	// 1.2.10 - Debian Stretch
	{
		PackageVersion: "1.2.10",
		Name:           "containerd.io",
		Distros:        []distros.Distribution{distros.DistributionDebian9},
		Architectures:  []Architecture{ArchitectureAmd64},
		Version:        "1.2.10-3",
		Source:         "https://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/containerd.io_1.2.10-3_amd64.deb",
		Hash:           "186f2f2c570f37b363102e6b879073db6dec671d",
	},

	// 1.2.10 - Debian Buster
	{
		PackageVersion: "1.2.10",
		Name:           "containerd.io",
		Distros:        []distros.Distribution{distros.DistributionDebian10},
		Architectures:  []Architecture{ArchitectureAmd64},
		Version:        "1.2.10-3",
		Source:         "https://download.docker.com/linux/debian/dists/buster/pool/stable/amd64/containerd.io_1.2.10-3_amd64.deb",
		Hash:           "365e4a7541ce2cf3c3036ea2a9bf6b40a50893a8",
	},

	// 1.2.10 - Ubuntu Xenial
	{
		PackageVersion: "1.2.10",
		Name:           "containerd.io",
		Distros:        []distros.Distribution{distros.DistributionXenial},
		Architectures:  []Architecture{ArchitectureAmd64},
		Version:        "1.2.10-3",
		Source:         "https://download.docker.com/linux/ubuntu/dists/xenial/pool/stable/amd64/containerd.io_1.2.10-3_amd64.deb",
		Hash:           "b64e7170d9176bc38967b2e12147c69b65bdd0fc",
	},

	// 1.2.10 - Ubuntu Bionic
	{
		PackageVersion: "1.2.10",
		Name:           "containerd.io",
		Distros:        []distros.Distribution{distros.DistributionBionic},
		Architectures:  []Architecture{ArchitectureAmd64},
		Version:        "1.2.10-3",
		Source:         "https://download.docker.com/linux/ubuntu/dists/bionic/pool/stable/amd64/containerd.io_1.2.10-3_amd64.deb",
		Hash:           "f4c941807310e3fa470dddfb068d599174a3daec",
	},

	// 1.2.10 - CentOS7 / Rhel 7
	{
		PackageVersion: "1.2.10",
		Name:           "containerd.io",
		Distros:        []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures:  []Architecture{ArchitectureAmd64},
		Version:        "1.2.10",
		Source:         "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/containerd.io-1.2.10-3.2.el7.x86_64.rpm",
		Hash:           "f6447e84479df3a58ce04a3da87ccc384663493b",
	},

	// 1.2.10 - CentOS8 / Rhel 8
	{
		PackageVersion: "1.2.10",
		Name:           "containerd.io",
		Distros:        []distros.Distribution{distros.DistributionRhel8, distros.DistributionCentos8},
		Architectures:  []Architecture{ArchitectureAmd64},
		Version:        "1.2.10",
		Source:         "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/containerd.io-1.2.10-3.2.el7.x86_64.rpm",
		Hash:           "f6447e84479df3a58ce04a3da87ccc384663493b",
	},

	// 1.2.10 - Linux Generic
	//
	// * AmazonLinux2: the Centos7 package depends on container-selinux, but selinux isn't used on amazonlinux2
	// * UbuntuFocal: no focal version available at download.docker.com
	{
		PackageVersion: "1.2.10",
		PlainBinary:    true,
		Distros:        []distros.Distribution{distros.DistributionAmazonLinux2, distros.DistributionFocal},
		Architectures:  []Architecture{ArchitectureAmd64},
		Source:         "https://storage.googleapis.com/cri-containerd-release/cri-containerd-1.2.10.linux-amd64.tar.gz",
		Hash:           "c84c29dcd1867a6ee9899d2106ab4f28854945f6",
	},

	// 1.2.11 - Linux Generic
	{
		PackageVersion: "1.2.11",
		PlainBinary:    true,
		Architectures:  []Architecture{ArchitectureAmd64},
		Source:         "https://storage.googleapis.com/cri-containerd-release/cri-containerd-1.2.11.linux-amd64.tar.gz",
		Hash:           "c98c9fdfd0984557e5b1a1f209213d2d8ad8471c",
	},

	// 1.2.12 - Linux Generic
	{
		PackageVersion: "1.2.12",
		PlainBinary:    true,
		Architectures:  []Architecture{ArchitectureAmd64},
		Source:         "https://storage.googleapis.com/cri-containerd-release/cri-containerd-1.2.12.linux-amd64.tar.gz",
		Hash:           "9455ca2508ad57438cb02a986ba763033bcb05f7",
	},

	// 1.2.13 - Linux Generic
	{
		PackageVersion: "1.2.13",
		PlainBinary:    true,
		Architectures:  []Architecture{ArchitectureAmd64},
		Source:         "https://storage.googleapis.com/cri-containerd-release/cri-containerd-1.2.13.linux-amd64.tar.gz",
		Hash:           "70ee2821e26116b0cddc679d14806fd20d25d65c",
	},

	// 1.3.2 - Linux Generic
	{
		PackageVersion: "1.3.2",
		PlainBinary:    true,
		Architectures:  []Architecture{ArchitectureAmd64},
		Source:         "https://storage.googleapis.com/cri-containerd-release/cri-containerd-1.3.2.linux-amd64.tar.gz",
		Hash:           "f451d46280104588f236bee277bca1da8babc0e8",
	},

	// 1.3.3 - Linux Generic
	{
		PackageVersion: "1.3.3",
		PlainBinary:    true,
		Architectures:  []Architecture{ArchitectureAmd64},
		Source:         "https://storage.googleapis.com/cri-containerd-release/cri-containerd-1.3.3.linux-amd64.tar.gz",
		Hash:           "921b74e84da366ec3eaa72ff97fa8d6ae56834c6",
	},

	// 1.3.4 - Linux Generic
	{
		PackageVersion: "1.3.4",
		PlainBinary:    true,
		Architectures:  []Architecture{ArchitectureAmd64},
		Source:         "https://storage.googleapis.com/cri-containerd-release/cri-containerd-1.3.4.linux-amd64.tar.gz",
		Hash:           "ce518d8091ffdd40caa7f386c742d9b1d03e01b5",
	},

	// TIP: When adding the next version, copy the previous version, string replace the version and run:
	//   VERIFY_HASHES=1 go test -v ./nodeup/pkg/model -run TestContainerdPackageHashes
	// (you might want to temporarily comment out older versions on a slower connection and then validate)
}

func (b *ContainerdBuilder) containerdVersion() (string, error) {
	containerdVersion := ""
	if b.Cluster.Spec.Containerd != nil {
		containerdVersion = fi.StringValue(b.Cluster.Spec.Containerd.Version)
	}
	if containerdVersion == "" {
		return "", fmt.Errorf("error finding containerd version")
	}
	return containerdVersion, nil
}

// Build is responsible for configuring the containerd daemon
func (b *ContainerdBuilder) Build(c *fi.ModelBuilderContext) error {
	if b.skipInstall() {
		klog.Infof("SkipInstall is set to true; won't install containerd")
		return nil
	}

	// @check: neither flatcar nor containeros need provision containerd.service, just the containerd daemon options
	switch b.Distribution {
	case distros.DistributionFlatcar:
		klog.Infof("Detected Flatcar; won't install containerd")
		if err := b.buildContainerOSConfigurationDropIn(c); err != nil {
			return err
		}
		return nil

	case distros.DistributionContainerOS:
		klog.Infof("Detected ContainerOS; won't install containerd")
		if err := b.buildContainerOSConfigurationDropIn(c); err != nil {
			return err
		}
		return nil
	}

	// Add Apache2 license
	{
		t := &nodetasks.File{
			Path:     "/usr/share/doc/containerd/apache.txt",
			Contents: fi.NewStringResource(resources.ContainerdApache2License),
			Type:     nodetasks.FileType_File,
		}
		c.AddTask(t)
	}

	// Add config file
	{
		containerdConfigOverride := ""
		if b.Cluster.Spec.Containerd != nil {
			containerdConfigOverride = fi.StringValue(b.Cluster.Spec.Containerd.ConfigOverride)
		}

		t := &nodetasks.File{
			Path:     "/etc/containerd/config-kops.toml",
			Contents: fi.NewStringResource(containerdConfigOverride),
			Type:     nodetasks.FileType_File,
		}
		c.AddTask(t)
	}

	containerdVersion, err := b.containerdVersion()
	if err != nil {
		return err
	}

	// Add packages
	{
		count := 0
		for i := range containerdVersions {
			dv := &containerdVersions[i]
			if !dv.matches(b.Architecture, containerdVersion, b.Distribution) {
				continue
			}

			count++

			var packageTask fi.Task
			if dv.PlainBinary {
				packageTask = &nodetasks.Archive{
					Name:      "containerd.io",
					Source:    dv.Source,
					Hash:      dv.Hash,
					TargetDir: "/",
					MapFiles: map[string]string{
						"./usr/local/bin":  "/usr",
						"./usr/local/sbin": "/usr",
					},
				}
				c.AddTask(packageTask)
			} else {
				var extraPkgs []*nodetasks.Package
				for name, pkg := range dv.ExtraPackages {
					dep := &nodetasks.Package{
						Name:         name,
						Version:      s(pkg.Version),
						Source:       s(pkg.Source),
						Hash:         s(pkg.Hash),
						PreventStart: fi.Bool(true),
					}
					extraPkgs = append(extraPkgs, dep)
				}
				packageTask = &nodetasks.Package{
					Name:    dv.Name,
					Version: s(dv.Version),
					Source:  s(dv.Source),
					Hash:    s(dv.Hash),
					Deps:    extraPkgs,

					// TODO: PreventStart is now unused?
					PreventStart: fi.Bool(true),
				}
				c.AddTask(packageTask)
			}

			// As a mitigation for CVE-2019-5736 (possibly a fix, definitely defense-in-depth) we chattr docker-runc to be immutable
			for _, f := range dv.MarkImmutable {
				c.AddTask(&nodetasks.Chattr{
					File: f,
					Mode: "+i",
					Deps: []fi.Task{packageTask},
				})
			}

			for _, dep := range dv.Dependencies {
				c.AddTask(&nodetasks.Package{Name: dep})
			}

			// Note we do _not_ stop looping... centos/rhel comprises multiple packages
		}

		if count == 0 {
			klog.Warningf("Did not find containerd package for %s %s %s", b.Distribution, b.Architecture, containerdVersion)
		}
	}

	c.AddTask(b.buildSystemdService())

	if err := b.buildSysconfig(c); err != nil {
		return err
	}

	// Using containerd with Kubenet requires special configuration. This is a temporary backwards-compatible solution
	// and will be deprecated when Kubenet is deprecated:
	// https://github.com/containerd/cri/blob/master/docs/config.md#cni-config-template
	usesKubenet := components.UsesKubenet(b.Cluster.Spec.Networking)
	if b.Cluster.Spec.ContainerRuntime == "containerd" && usesKubenet {
		b.buildKubenetCNIConfigTemplate(c)
	}

	return nil
}

func (b *ContainerdBuilder) buildSystemdService() *nodetasks.Service {
	// Based on https://github.com/containerd/cri/blob/master/contrib/systemd-units/containerd.service

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "containerd container runtime")
	manifest.Set("Unit", "Documentation", "https://containerd.io")
	manifest.Set("Unit", "After", "network.target local-fs.target")

	manifest.Set("Service", "EnvironmentFile", "/etc/sysconfig/containerd")
	manifest.Set("Service", "EnvironmentFile", "/etc/environment")
	manifest.Set("Service", "ExecStartPre", "-/sbin/modprobe overlay")
	manifest.Set("Service", "ExecStart", "/usr/bin/containerd -c /etc/containerd/config-kops.toml \"$CONTAINERD_OPTS\"")

	manifest.Set("Service", "Restart", "always")
	manifest.Set("Service", "RestartSec", "5")

	// set delegate yes so that systemd does not reset the cgroups of containerd containers
	manifest.Set("Service", "Delegate", "yes")
	// kill only the containerd process, not all processes in the cgroup
	manifest.Set("Service", "KillMode", "process")
	// make killing of processes of this unit under memory pressure very unlikely
	manifest.Set("Service", "OOMScoreAdjust", "-999")

	manifest.Set("Service", "LimitNOFILE", "1048576")
	manifest.Set("Service", "LimitNPROC", "infinity")
	manifest.Set("Service", "LimitCORE", "infinity")
	manifest.Set("Service", "TasksMax", "infinity")

	manifest.Set("Install", "WantedBy", "multi-user.target")

	manifestString := manifest.Render()
	klog.V(8).Infof("Built service manifest %q\n%s", "containerd", manifestString)

	service := &nodetasks.Service{
		Name:       "containerd.service",
		Definition: s(manifestString),
	}

	service.InitDefaults()

	return service
}

// buildContainerOSConfigurationDropIn is responsible for configuring the containerd daemon options
func (b *ContainerdBuilder) buildContainerOSConfigurationDropIn(c *fi.ModelBuilderContext) error {
	lines := []string{
		"[Service]",
		"EnvironmentFile=/etc/sysconfig/containerd",
		"EnvironmentFile=/etc/environment",
		"TasksMax=infinity",
	}
	contents := strings.Join(lines, "\n")

	c.AddTask(&nodetasks.File{
		AfterFiles: []string{"/etc/sysconfig/containerd"},
		Path:       "/etc/systemd/system/containerd.service.d/10-kops.conf",
		Contents:   fi.NewStringResource(contents),
		Type:       nodetasks.FileType_File,
		OnChangeExecute: [][]string{
			{"systemctl", "daemon-reload"},
			{"systemctl", "restart", "containerd.service"},
			// We need to restart kops-configuration service since nodeup needs to load images
			// into containerd with the new config. Restart is on the background because
			// kops-configuration is of type 'one-shot' so the restart command will wait for
			// nodeup to finish executing
			{"systemctl", "restart", "kops-configuration.service", "&"},
		},
	})

	if err := b.buildSysconfig(c); err != nil {
		return err
	}

	return nil
}

// buildSysconfig is responsible for extracting the containerd configuration and writing the sysconfig file
func (b *ContainerdBuilder) buildSysconfig(c *fi.ModelBuilderContext) error {
	var containerd kops.ContainerdConfig
	if b.Cluster.Spec.Containerd != nil {
		containerd = *b.Cluster.Spec.Containerd
	}

	flagsString, err := flagbuilder.BuildFlags(&containerd)
	if err != nil {
		return fmt.Errorf("error building containerd flags: %v", err)
	}

	lines := []string{
		"CONTAINERD_OPTS=" + flagsString,
	}
	contents := strings.Join(lines, "\n")

	c.AddTask(&nodetasks.File{
		Path:     "/etc/sysconfig/containerd",
		Contents: fi.NewStringResource(contents),
		Type:     nodetasks.FileType_File,
	})

	return nil
}

// buildKubenetCNIConfigTemplate is responsible for creating a special template for setups using Kubenet
func (b *ContainerdBuilder) buildKubenetCNIConfigTemplate(c *fi.ModelBuilderContext) {
	lines := []string{
		"{",
		"    \"cniVersion\": \"0.3.1\",",
		"    \"name\": \"kubenet\",",
		"    \"plugins\": [",
		"        {",
		"            \"type\": \"bridge\",",
		"            \"bridge\": \"cbr0\",",
		"            \"mtu\": 1460,",
		"            \"addIf\": \"eth0\",",
		"            \"isGateway\": true,",
		"            \"ipMasq\": true,",
		"            \"promiscMode\": true,",
		"            \"ipam\": {",
		"                \"type\": \"host-local\",",
		"                \"subnet\": \"{{.PodCIDR}}\",",
		"                \"routes\": [{ \"dst\": \"0.0.0.0/0\" }]",
		"            }",
		"        }",
		"    ]",
		"}",
	}
	contents := strings.Join(lines, "\n")
	klog.V(8).Infof("Built kubenet CNI config file\n%s", contents)

	c.AddTask(&nodetasks.File{
		Path:     "/etc/containerd/cni-config.template",
		Contents: fi.NewStringResource(contents),
		Type:     nodetasks.FileType_File,
	})
}

// skipInstall determines if kops should skip the installation and configuration of containerd
func (b *ContainerdBuilder) skipInstall() bool {
	d := b.Cluster.Spec.Containerd

	// don't skip install if the user hasn't specified anything
	if d == nil {
		return false
	}

	return d.SkipInstall
}
