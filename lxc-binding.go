// Copyright © 2013, 2014, The Go-LXC Authors. All rights reserved.
// Use of this source code is governed by a LGPLv2.1
// license that can be found in the LICENSE file.

// +build linux,cgo

package lxc

// #include <lxc/lxccontainer.h>
// #include <lxc/version.h>
// #include "lxc-binding.h"
// #ifndef LXC_DEVEL
// #define LXC_DEVEL 0
// #endif
import "C"

import (
	"fmt"
	"strconv"
	"strings"
	"unsafe"
)

// NewContainer returns a new container struct.
// Caller needs to call Release() on the returned container to release its resources.
func NewContainer(name string, lxcpath ...string) (*Container, error) {
	var container *C.struct_lxc_container

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	if lxcpath != nil && len(lxcpath) == 1 {
		clxcpath := C.CString(lxcpath[0])
		defer C.free(unsafe.Pointer(clxcpath))

		container = C.lxc_container_new(cname, clxcpath)
	} else {
		container = C.lxc_container_new(cname, nil)
	}

	if container == nil {
		return nil, ErrNewFailed
	}
	c := &Container{container: container, verbosity: Quiet}

	return c, nil
}

// Acquire increments the reference counter of the container object.
func Acquire(c *Container) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return C.lxc_container_get(c.container) == 1
}

// Release decrements the reference counter of the container object.
func Release(c *Container) bool {
	return c.Release() == nil
}

// Version returns the LXC version.
func Version() string {
	version := C.GoString(C.lxc_get_version())

	// New liblxc versions append "-devel" when LXC_DEVEL is set.
	if strings.HasSuffix(version, "-devel") {
		return fmt.Sprintf("%s (devel)", version[:(len(version)-len("-devel"))])
	}

	return version
}

// GlobalConfigItem returns the value of the given global config key.
func GlobalConfigItem(name string) string {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	return C.GoString(C.lxc_get_global_config_item(cname))
}

// DefaultConfigPath returns default config path.
func DefaultConfigPath() string {
	return GlobalConfigItem("lxc.lxcpath")
}

// DefaultLvmVg returns the name of the default LVM volume group.
func DefaultLvmVg() string {
	return GlobalConfigItem("lxc.bdev.lvm.vg")
}

// DefaultZfsRoot returns the name of the default ZFS root.
func DefaultZfsRoot() string {
	return GlobalConfigItem("lxc.bdev.zfs.root")
}

// ContainerNames returns the names of defined and active containers on the system.
func ContainerNames(lxcpath ...string) []string {
	var size int
	var cnames **C.char

	if lxcpath != nil && len(lxcpath) == 1 {
		clxcpath := C.CString(lxcpath[0])
		defer C.free(unsafe.Pointer(clxcpath))

		size = int(C.list_all_containers(clxcpath, &cnames, nil))
	} else {

		size = int(C.list_all_containers(nil, &cnames, nil))
	}

	if size < 1 {
		return nil
	}
	return convertNArgs(cnames, size)
}

// Containers returns the defined and active containers on the system. Only
// containers that could retrieved successfully are returned.
// Caller needs to call Release() on the returned containers to release resources.
func Containers(lxcpath ...string) []*Container {
	var containers []*Container

	for _, v := range ContainerNames(lxcpath...) {
		if container, err := NewContainer(v, lxcpath...); err == nil {
			containers = append(containers, container)
		}
	}

	return containers
}

// DefinedContainerNames returns the names of the defined containers on the system.
func DefinedContainerNames(lxcpath ...string) []string {
	var size int
	var cnames **C.char

	if lxcpath != nil && len(lxcpath) == 1 {
		clxcpath := C.CString(lxcpath[0])
		defer C.free(unsafe.Pointer(clxcpath))

		size = int(C.list_defined_containers(clxcpath, &cnames, nil))
	} else {

		size = int(C.list_defined_containers(nil, &cnames, nil))
	}

	if size < 1 {
		return nil
	}
	return convertNArgs(cnames, size)
}

// DefinedContainers returns the defined containers on the system.  Only
// containers that could retrieved successfully are returned.
// Caller needs to call Release() on the returned containers to release resources.
func DefinedContainers(lxcpath ...string) []*Container {
	var containers []*Container

	for _, v := range DefinedContainerNames(lxcpath...) {
		if container, err := NewContainer(v, lxcpath...); err == nil {
			containers = append(containers, container)
		}
	}

	return containers
}

// ActiveContainerNames returns the names of the active containers on the system.
func ActiveContainerNames(lxcpath ...string) []string {
	var size int
	var cnames **C.char

	if lxcpath != nil && len(lxcpath) == 1 {
		clxcpath := C.CString(lxcpath[0])
		defer C.free(unsafe.Pointer(clxcpath))

		size = int(C.list_active_containers(clxcpath, &cnames, nil))
	} else {

		size = int(C.list_active_containers(nil, &cnames, nil))
	}

	if size < 1 {
		return nil
	}
	return convertNArgs(cnames, size)
}

// ActiveContainers returns the active containers on the system. Only
// containers that could retrieved successfully are returned.
// Caller needs to call Release() on the returned containers to release resources.
func ActiveContainers(lxcpath ...string) []*Container {
	var containers []*Container

	for _, v := range ActiveContainerNames(lxcpath...) {
		if container, err := NewContainer(v, lxcpath...); err == nil {
			containers = append(containers, container)
		}
	}

	return containers
}

// VersionNumber returns the LXC version.
func VersionNumber() (major int, minor int) {
	major = C.LXC_VERSION_MAJOR
	minor = C.LXC_VERSION_MINOR

	return
}

// VersionAtLeast returns true when the tested version >= current version.
func VersionAtLeast(major int, minor int, micro int) bool {
	if C.LXC_DEVEL == 1 {
		return true
	}

	if major > C.LXC_VERSION_MAJOR {
		return false
	}

	if major == C.LXC_VERSION_MAJOR &&
		minor > C.LXC_VERSION_MINOR {
		return false
	}

	if major == C.LXC_VERSION_MAJOR &&
		minor == C.LXC_VERSION_MINOR &&
		micro > C.LXC_VERSION_MICRO {
		return false
	}

	return true
}

// IsSupportedConfigItem returns true if the key belongs to a supported config item.
func IsSupportedConfigItem(key string) bool {
	configItem := C.CString(key)
	defer C.free(unsafe.Pointer(configItem))
	return bool(C.go_lxc_config_item_is_supported(configItem))
}

// RuntimeLiblxcVersionAtLeast checks if the system's liblxc matches the
// provided version requirement
func RuntimeLiblxcVersionAtLeast(version string, major int, minor int, micro int) bool {
	// Strip git versioning from pre-release snapshots.
	version = strings.Split(version, "~")[0]

	// Convert devel indicator into a valid version.
	version = strings.Replace(version, " (devel)", "-devel", 1)

	// Split the version into its major, minor and micro parts.
	parts := strings.Split(version, ".")
	partsLen := len(parts)
	if partsLen == 0 {
		return false
	}

	// If the last part includes -devel, assume everything is supported.
	develParts := strings.Split(parts[partsLen-1], "-")
	if len(develParts) == 2 && develParts[1] == "devel" {
		return true
	}

	// Actually parse and compare the version string now.
	maj := -1
	min := -1
	mic := -1

	for i, v := range parts {
		if i > 2 {
			break
		}

		num, err := strconv.Atoi(v)
		if err != nil {
			return false
		}

		switch i {
		case 0:
			maj = num
		case 1:
			min = num
		case 2:
			mic = num
		}
	}

	/* Major version is greater. */
	if maj > major {
		return true
	}

	if maj < major {
		return false
	}

	/* Minor number is greater.*/
	if min > minor {
		return true
	}

	if min < minor {
		return false
	}

	/* Patch number is greater. */
	if mic > micro {
		return true
	}

	if mic < micro {
		return false
	}

	return true
}

// HasApiExtension returns true if the extension is supported.
// Deprecated: Please use HasAPIExtension instead.
func HasApiExtension(extension string) bool {
	return HasAPIExtension(extension)
}

// HasAPIExtension returns true if the extension is supported.
func HasAPIExtension(extension string) bool {
	if RuntimeLiblxcVersionAtLeast(Version(), 3, 1, 0) {
		apiExtension := C.CString(extension)
		defer C.free(unsafe.Pointer(apiExtension))
		return bool(C.go_lxc_has_api_extension(apiExtension))
	}
	return false
}
