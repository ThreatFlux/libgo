package ova

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/threatflux/libgo/internal/models/vm"
	"github.com/threatflux/libgo/pkg/logger"
)

// OVFTemplateGenerator generates OVF templates.
type OVFTemplateGenerator struct {
	templateLoader *template.Template
	logger         logger.Logger
}

// OVFTemplateData holds data for OVF template rendering.
type OVFTemplateData struct {
	FileID          string
	VMName          string
	VMID            string
	DiskPath        string
	DiskSizeBytes   uint64
	DiskSizeMB      uint64
	CPUCount        int
	MemorySizeMB    uint64
	TimeStamp       string
	HardwareVersion string
	OSType          string
}

// NewOVFTemplateGenerator creates a new OVFTemplateGenerator.
func NewOVFTemplateGenerator(logger logger.Logger) (*OVFTemplateGenerator, error) {
	// Load the OVF template
	tmpl, err := template.New("ovf").Parse(ovfTemplateContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OVF template: %w", err)
	}

	return &OVFTemplateGenerator{
		templateLoader: tmpl,
		logger:         logger,
	}, nil
}

// GenerateOVF generates an OVF descriptor.
func (g *OVFTemplateGenerator) GenerateOVF(vm *vm.VM, diskPath string, diskSize uint64) (string, error) {
	// Prepare template data
	diskSizeMB := diskSize / (1024 * 1024)
	if diskSize%(1024*1024) > 0 {
		diskSizeMB++ // Round up
	}

	// Initialize data with defaults if not provided
	data := OVFTemplateData{
		FileID:          uuid.New().String(),
		VMName:          vm.Name,
		VMID:            vm.UUID,
		DiskPath:        filepath.Base(diskPath),
		DiskSizeBytes:   diskSize,
		DiskSizeMB:      diskSizeMB,
		CPUCount:        vm.CPU.Count,
		MemorySizeMB:    vm.Memory.SizeBytes / (1024 * 1024),
		TimeStamp:       time.Now().Format(time.RFC3339),
		HardwareVersion: "vmx-10", // VMware Workstation 10+, ESXi 5.5+
		OSType:          "otherLinux64Guest",
	}

	// Use defaults if values are zero
	if data.VMID == "" {
		data.VMID = uuid.New().String()
	}
	if data.CPUCount == 0 {
		data.CPUCount = 1
	}
	if data.MemorySizeMB == 0 {
		data.MemorySizeMB = 1024 // 1 GB
	}

	// Render the template
	var buffer bytes.Buffer
	if err := g.templateLoader.Execute(&buffer, data); err != nil {
		return "", fmt.Errorf("failed to render OVF template: %w", err)
	}

	return buffer.String(), nil
}

// WriteOVFToFile writes OVF to a file.
func (g *OVFTemplateGenerator) WriteOVFToFile(ovfContent string, outPath string) error {
	return os.WriteFile(outPath, []byte(ovfContent), 0644)
}

// ovfTemplateContent contains the OVF template.
const ovfTemplateContent = `<?xml version="1.0" encoding="UTF-8"?>
<Envelope xmlns="http://schemas.dmtf.org/ovf/envelope/1"
          xmlns:cim="http://schemas.dmtf.org/wbem/wscim/1/common"
          xmlns:ovf="http://schemas.dmtf.org/ovf/envelope/1"
          xmlns:rasd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ResourceAllocationSettingData"
          xmlns:vmw="http://www.vmware.com/schema/ovf"
          xmlns:vssd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_VirtualSystemSettingData"
          xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <References>
    <File ovf:href="{{.DiskPath}}" ovf:id="{{.FileID}}" ovf:size="{{.DiskSizeBytes}}"/>
  </References>
  <DiskSection>
    <Info>Virtual disk information</Info>
    <Disk ovf:capacity="{{.DiskSizeMB}}" ovf:capacityAllocationUnits="byte * 2^20" ovf:diskId="vmdisk1" ovf:fileRef="{{.FileID}}" ovf:format="http://www.vmware.com/interfaces/specifications/vmdk.html#streamOptimized"/>
  </DiskSection>
  <NetworkSection>
    <Info>The list of logical networks</Info>
    <Network ovf:name="VM Network">
      <Description>The VM Network network</Description>
    </Network>
  </NetworkSection>
  <VirtualSystem ovf:id="{{.VMName}}">
    <Info>A virtual machine</Info>
    <Name>{{.VMName}}</Name>
    <OperatingSystemSection ovf:id="{{.OSType}}">
      <Info>The kind of installed guest operating system</Info>
      <Description>Linux</Description>
    </OperatingSystemSection>
    <VirtualHardwareSection>
      <Info>Virtual hardware requirements</Info>
      <System>
        <vssd:ElementName>Virtual Hardware Family</vssd:ElementName>
        <vssd:InstanceID>0</vssd:InstanceID>
        <vssd:VirtualSystemIdentifier>{{.VMName}}</vssd:VirtualSystemIdentifier>
        <vssd:VirtualSystemType>{{.HardwareVersion}}</vssd:VirtualSystemType>
      </System>
      <Item>
        <rasd:AllocationUnits>hertz * 10^6</rasd:AllocationUnits>
        <rasd:Description>Number of Virtual CPUs</rasd:Description>
        <rasd:ElementName>{{.CPUCount}} virtual CPU(s)</rasd:ElementName>
        <rasd:InstanceID>1</rasd:InstanceID>
        <rasd:ResourceType>3</rasd:ResourceType>
        <rasd:VirtualQuantity>{{.CPUCount}}</rasd:VirtualQuantity>
      </Item>
      <Item>
        <rasd:AllocationUnits>byte * 2^20</rasd:AllocationUnits>
        <rasd:Description>Memory Size</rasd:Description>
        <rasd:ElementName>{{.MemorySizeMB}} MB of memory</rasd:ElementName>
        <rasd:InstanceID>2</rasd:InstanceID>
        <rasd:ResourceType>4</rasd:ResourceType>
        <rasd:VirtualQuantity>{{.MemorySizeMB}}</rasd:VirtualQuantity>
      </Item>
      <Item>
        <rasd:Address>0</rasd:Address>
        <rasd:Description>SCSI Controller</rasd:Description>
        <rasd:ElementName>SCSI controller 0</rasd:ElementName>
        <rasd:InstanceID>3</rasd:InstanceID>
        <rasd:ResourceSubType>lsilogic</rasd:ResourceSubType>
        <rasd:ResourceType>6</rasd:ResourceType>
      </Item>
      <Item>
        <rasd:AddressOnParent>0</rasd:AddressOnParent>
        <rasd:ElementName>Hard disk 1</rasd:ElementName>
        <rasd:HostResource>ovf:/disk/vmdisk1</rasd:HostResource>
        <rasd:InstanceID>4</rasd:InstanceID>
        <rasd:Parent>3</rasd:Parent>
        <rasd:ResourceType>17</rasd:ResourceType>
      </Item>
      <Item>
        <rasd:AddressOnParent>7</rasd:AddressOnParent>
        <rasd:AutomaticAllocation>true</rasd:AutomaticAllocation>
        <rasd:Connection>VM Network</rasd:Connection>
        <rasd:Description>E1000 Network Adapter</rasd:Description>
        <rasd:ElementName>Network adapter 1</rasd:ElementName>
        <rasd:InstanceID>5</rasd:InstanceID>
        <rasd:ResourceSubType>E1000</rasd:ResourceSubType>
        <rasd:ResourceType>10</rasd:ResourceType>
      </Item>
    </VirtualHardwareSection>
    <ProductSection>
      <Info>Information about the installed software</Info>
      <Product>{{.VMName}}</Product>
      <Vendor>LibGo KVM Manager</Vendor>
      <Version>1.0</Version>
      <FullVersion>1.0.0</FullVersion>
      <ProductUrl>https://github.com/threatflux/libgo</ProductUrl>
      <VendorUrl>https://github.com/threatflux</VendorUrl>
    </ProductSection>
  </VirtualSystem>
</Envelope>`
