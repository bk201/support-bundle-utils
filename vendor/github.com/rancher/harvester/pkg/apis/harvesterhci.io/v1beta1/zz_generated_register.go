/*
Copyright 2021 Rancher Labs, Inc.

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

// Code generated by main. DO NOT EDIT.

// +k8s:deepcopy-gen=package
// +groupName=harvesterhci.io
package v1beta1

import (
	harvesterhci "github.com/rancher/harvester/pkg/apis/harvesterhci.io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	KeyPairResourceName                       = "keypairs"
	PreferenceResourceName                    = "preferences"
	SettingResourceName                       = "settings"
	SupportBundleResourceName                 = "supportbundles"
	UpgradeResourceName                       = "upgrades"
	UserResourceName                          = "users"
	VirtualMachineBackupResourceName          = "virtualmachinebackups"
	VirtualMachineBackupContentResourceName   = "virtualmachinebackupcontents"
	VirtualMachineImageResourceName           = "virtualmachineimages"
	VirtualMachineRestoreResourceName         = "virtualmachinerestores"
	VirtualMachineTemplateResourceName        = "virtualmachinetemplates"
	VirtualMachineTemplateVersionResourceName = "virtualmachinetemplateversions"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: harvesterhci.GroupName, Version: "v1beta1"}

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

// Adds the list of known types to Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&KeyPair{},
		&KeyPairList{},
		&Preference{},
		&PreferenceList{},
		&Setting{},
		&SettingList{},
		&SupportBundle{},
		&SupportBundleList{},
		&Upgrade{},
		&UpgradeList{},
		&User{},
		&UserList{},
		&VirtualMachineBackup{},
		&VirtualMachineBackupList{},
		&VirtualMachineBackupContent{},
		&VirtualMachineBackupContentList{},
		&VirtualMachineImage{},
		&VirtualMachineImageList{},
		&VirtualMachineRestore{},
		&VirtualMachineRestoreList{},
		&VirtualMachineTemplate{},
		&VirtualMachineTemplateList{},
		&VirtualMachineTemplateVersion{},
		&VirtualMachineTemplateVersionList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}