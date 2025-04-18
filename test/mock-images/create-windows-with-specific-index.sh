#!/bin/bash
# Script to create a Windows Server VM with fully automated installation
# Addressing "No images are available" error with specific image index value

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Log function
log() {
  local level=$1
  local message=$2
  local color=$NC

  case $level in
    "INFO") color=$GREEN ;;
    "WARN") color=$YELLOW ;;
    "ERROR") color=$RED ;;
    "STEP") color=$BLUE ;;
  esac

  echo -e "${color}[$level] $message${NC}"
}

# Variables
VM_NAME="windows-auto-test"
PROJECT_ROOT="/home/vtriple/libgo/libgo"
ISO_PATH="$PROJECT_ROOT/iso/windows-server-2022.iso"
VIRTIO_ISO="$PROJECT_ROOT/virtio-win.iso"
STORAGE_DIR="/home/vtriple/libgo-storage"
DISK_PATH="$STORAGE_DIR/windows-auto-test.qcow2"
AUTOUNATTEND_DIR="$PROJECT_ROOT/tmp/autounattend-iso"
AUTOUNATTEND_ISO="$PROJECT_ROOT/tmp/autounattend.iso"

# Check ISO exists
if [ ! -f "$ISO_PATH" ]; then
  log "ERROR" "Windows ISO not found at $ISO_PATH"
  exit 1
fi

# Create storage directories
log "INFO" "Creating storage directories..."
mkdir -p "$STORAGE_DIR"
mkdir -p "$PROJECT_ROOT/tmp"
mkdir -p "$AUTOUNATTEND_DIR"

# Create autounattend.xml file - modified to use IMAGE/NAME instead of INDEX
log "STEP" "Creating autounattend.xml file with IMAGE/NAME..."
cat > "$AUTOUNATTEND_DIR/autounattend.xml" << 'EOF'
<?xml version="1.0" encoding="utf-8"?>
<unattend xmlns="urn:schemas-microsoft-com:unattend">
    <settings pass="windowsPE">
        <component name="Microsoft-Windows-International-Core-WinPE" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <SetupUILanguage>
                <UILanguage>en-US</UILanguage>
            </SetupUILanguage>
            <InputLocale>en-US</InputLocale>
            <SystemLocale>en-US</SystemLocale>
            <UILanguage>en-US</UILanguage>
            <UILanguageFallback>en-US</UILanguageFallback>
            <UserLocale>en-US</UserLocale>
        </component>
        <component name="Microsoft-Windows-Setup" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <DiskConfiguration>
                <Disk wcm:action="add">
                    <CreatePartitions>
                        <CreatePartition wcm:action="add">
                            <Order>1</Order>
                            <Type>Primary</Type>
                            <Extend>true</Extend>
                        </CreatePartition>
                    </CreatePartitions>
                    <ModifyPartitions>
                        <ModifyPartition wcm:action="add">
                            <Format>NTFS</Format>
                            <Label>Windows</Label>
                            <Letter>C</Letter>
                            <Order>1</Order>
                            <PartitionID>1</PartitionID>
                            <Active>true</Active>
                        </ModifyPartition>
                    </ModifyPartitions>
                    <DiskID>0</DiskID>
                    <WillWipeDisk>true</WillWipeDisk>
                </Disk>
            </DiskConfiguration>
            <ImageInstall>
                <OSImage>
                    <InstallFrom>
                        <MetaData wcm:action="add">
                            <Key>/IMAGE/NAME</Key>
                            <Value>Windows Server 2022 SERVERSTANDARDCORE</Value>
                        </MetaData>
                    </InstallFrom>
                    <InstallTo>
                        <DiskID>0</DiskID>
                        <PartitionID>1</PartitionID>
                    </InstallTo>
                </OSImage>
            </ImageInstall>
            <UserData>
                <AcceptEula>true</AcceptEula>
                <FullName>Administrator</FullName>
                <Organization>LibGo Test</Organization>
            </UserData>
            <EnableFirewall>false</EnableFirewall>
        </component>
    </settings>
    <settings pass="specialize">
        <component name="Microsoft-Windows-Shell-Setup" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <ComputerName>WIN-SERVER</ComputerName>
            <TimeZone>Eastern Standard Time</TimeZone>
        </component>
        <component name="Microsoft-Windows-ServerManager-SvrMgrNc" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <DoNotOpenServerManagerAtLogon>true</DoNotOpenServerManagerAtLogon>
        </component>
    </settings>
    <settings pass="oobeSystem">
        <component name="Microsoft-Windows-Shell-Setup" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <AutoLogon>
                <Password>
                    <Value>Password123</Value>
                    <PlainText>true</PlainText>
                </Password>
                <Enabled>true</Enabled>
                <LogonCount>999</LogonCount>
                <Username>Administrator</Username>
            </AutoLogon>
            <OOBE>
                <HideEULAPage>true</HideEULAPage>
                <HideLocalAccountScreen>true</HideLocalAccountScreen>
                <HideOEMRegistrationScreen>true</HideOEMRegistrationScreen>
                <HideOnlineAccountScreens>true</HideOnlineAccountScreens>
                <HideWirelessSetupInOOBE>true</HideWirelessSetupInOOBE>
                <NetworkLocation>Work</NetworkLocation>
                <ProtectYourPC>1</ProtectYourPC>
            </OOBE>
            <UserAccounts>
                <AdministratorPassword>
                    <Value>Password123</Value>
                    <PlainText>true</PlainText>
                </AdministratorPassword>
            </UserAccounts>
            <RegisteredOrganization>LibGo Test</RegisteredOrganization>
            <RegisteredOwner>Administrator</RegisteredOwner>
            <TimeZone>Eastern Standard Time</TimeZone>
            <FirstLogonCommands>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>cmd.exe /c powershell -Command "Install-WindowsFeature -name Web-Server -IncludeManagementTools"</CommandLine>
                    <Description>Install IIS</Description>
                    <Order>1</Order>
                    <RequiresUserInput>false</RequiresUserInput>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>cmd.exe /c shutdown /s /t 30 /c "Installation complete"</CommandLine>
                    <Description>Shutdown when done</Description>
                    <Order>2</Order>
                    <RequiresUserInput>false</RequiresUserInput>
                </SynchronousCommand>
            </FirstLogonCommands>
        </component>
    </settings>
</unattend>
EOF

# Also try with "DATA" image type which might be right for this specific Windows Server ISO
log "STEP" "Creating alternate autounattend-alternate.xml file..."
cat > "$AUTOUNATTEND_DIR/autounattend-alternate.xml" << 'EOF'
<?xml version="1.0" encoding="utf-8"?>
<unattend xmlns="urn:schemas-microsoft-com:unattend">
    <settings pass="windowsPE">
        <component name="Microsoft-Windows-International-Core-WinPE" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <SetupUILanguage>
                <UILanguage>en-US</UILanguage>
            </SetupUILanguage>
            <InputLocale>en-US</InputLocale>
            <SystemLocale>en-US</SystemLocale>
            <UILanguage>en-US</UILanguage>
            <UILanguageFallback>en-US</UILanguageFallback>
            <UserLocale>en-US</UserLocale>
        </component>
        <component name="Microsoft-Windows-Setup" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <DiskConfiguration>
                <Disk wcm:action="add">
                    <CreatePartitions>
                        <CreatePartition wcm:action="add">
                            <Order>1</Order>
                            <Type>Primary</Type>
                            <Extend>true</Extend>
                        </CreatePartition>
                    </CreatePartitions>
                    <ModifyPartitions>
                        <ModifyPartition wcm:action="add">
                            <Format>NTFS</Format>
                            <Label>Windows</Label>
                            <Letter>C</Letter>
                            <Order>1</Order>
                            <PartitionID>1</PartitionID>
                            <Active>true</Active>
                        </ModifyPartition>
                    </ModifyPartitions>
                    <DiskID>0</DiskID>
                    <WillWipeDisk>true</WillWipeDisk>
                </Disk>
            </DiskConfiguration>
            <ImageInstall>
                <OSImage>
                    <InstallFrom>
                        <MetaData wcm:action="add">
                            <Key>/IMAGE/NAME</Key>
                            <Value>Windows Server 2022 SERVERSTANDARD</Value>
                        </MetaData>
                    </InstallFrom>
                    <InstallTo>
                        <DiskID>0</DiskID>
                        <PartitionID>1</PartitionID>
                    </InstallTo>
                </OSImage>
            </ImageInstall>
            <UserData>
                <AcceptEula>true</AcceptEula>
                <FullName>Administrator</FullName>
                <Organization>LibGo Test</Organization>
            </UserData>
            <EnableFirewall>false</EnableFirewall>
        </component>
    </settings>
    <settings pass="specialize">
        <component name="Microsoft-Windows-Shell-Setup" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <ComputerName>WIN-SERVER</ComputerName>
            <TimeZone>Eastern Standard Time</TimeZone>
        </component>
    </settings>
    <settings pass="oobeSystem">
        <component name="Microsoft-Windows-Shell-Setup" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <AutoLogon>
                <Password>
                    <Value>Password123</Value>
                    <PlainText>true</PlainText>
                </Password>
                <Enabled>true</Enabled>
                <LogonCount>999</LogonCount>
                <Username>Administrator</Username>
            </AutoLogon>
            <OOBE>
                <HideEULAPage>true</HideEULAPage>
                <HideLocalAccountScreen>true</HideLocalAccountScreen>
                <HideOEMRegistrationScreen>true</HideOEMRegistrationScreen>
                <HideOnlineAccountScreens>true</HideOnlineAccountScreens>
                <HideWirelessSetupInOOBE>true</HideWirelessSetupInOOBE>
                <NetworkLocation>Work</NetworkLocation>
                <ProtectYourPC>1</ProtectYourPC>
            </OOBE>
            <UserAccounts>
                <AdministratorPassword>
                    <Value>Password123</Value>
                    <PlainText>true</PlainText>
                </AdministratorPassword>
            </UserAccounts>
        </component>
    </settings>
</unattend>
EOF

# Create all key files Windows looks for with exact case
log "STEP" "Creating additional autounattend files with exact case sensitivity..."
touch "$AUTOUNATTEND_DIR/autounattend.xml"
cp "$AUTOUNATTEND_DIR/autounattend.xml" "$AUTOUNATTEND_DIR/Autounattend.xml"
cp "$AUTOUNATTEND_DIR/autounattend.xml" "$AUTOUNATTEND_DIR/AutoUnattend.xml"
mkdir -p "$AUTOUNATTEND_DIR/sources"
touch "$AUTOUNATTEND_DIR/autorun.inf"

# Create autounattend.iso with both versions - being comprehensive here
log "STEP" "Creating ISO image with autounattend.xml..."
rm -f "$AUTOUNATTEND_ISO"
genisoimage -J -r -o "$AUTOUNATTEND_ISO" "$AUTOUNATTEND_DIR"

# Remove any existing VM
log "STEP" "Cleaning up any existing VM..."
virsh destroy "$VM_NAME" 2>/dev/null || true
virsh undefine "$VM_NAME" --remove-all-storage 2>/dev/null || true
rm -f "$DISK_PATH"

# Create disk image
log "STEP" "Creating disk image..."
qemu-img create -f qcow2 "$DISK_PATH" 40G

# Create the VM with primary boot from CD-ROM, simplified VM creation
log "STEP" "Creating Windows VM with virt-install using SATA controller..."
virt-install \
  --name "$VM_NAME" \
  --memory 4096 \
  --vcpus 4 \
  --os-variant win2k22 \
  --disk path="$DISK_PATH",format=qcow2,bus=sata \
  --cdrom "$ISO_PATH" \
  --disk "$AUTOUNATTEND_ISO",device=cdrom \
  --network default \
  --graphics vnc,listen=0.0.0.0 \
  --boot cdrom,hd \
  --noautoconsole

# Display VNC information
VNC_DISPLAY=$(virsh vncdisplay "$VM_NAME" 2>/dev/null || echo "unknown")
log "INFO" "Windows VM created successfully with SATA controller!"
log "INFO" "Connect to the VM using VNC: $VNC_DISPLAY"
log "INFO" "The installation should proceed automatically"
log "INFO" "If you see 'No images available', try moving the autounattend.iso up in boot order"
log "INFO" "To switch autounattend XML, press Shift+F10 in Windows setup and explore D: or E: drive"
