package xmlutils

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	testVMName = "test-vm"
)

// Sample XML data for testing.
const testXML = `<?xml version="1.0" encoding="UTF-8"?>
<domain type="kvm">
  <name>` + testVMName + `</name>
  <uuid>12345678-1234-1234-1234-123456789012</uuid>
  <memory unit="KiB">2097152</memory>
  <vcpu placement="static">2</vcpu>
  <devices>
    <disk type="file" device="disk">
      <driver name="qemu" type="qcow2"/>
      <source file="/var/lib/libvirt/images/test-vm.qcow2"/>
      <target dev="vda" bus="virtio"/>
    </disk>
    <interface type="bridge">
      <source bridge="virbr0"/>
      <mac address="52:54:00:11:22:33"/>
      <model type="virtio"/>
    </interface>
  </devices>
</domain>`

// Test structure for XML unmarshaling.
type TestDomain struct {
	XMLName xml.Name `xml:"domain"`
	Type    string   `xml:"type,attr"`
	Name    string   `xml:"name"`
	UUID    string   `xml:"uuid"`
	Memory  struct {
		Unit  string `xml:"unit,attr"` // 16 bytes (string)
		Value int    `xml:",chardata"` // 8 bytes (int)
	} `xml:"memory"`
	VCPU struct {
		Placement string `xml:"placement,attr"` // 16 bytes (string)
		Value     int    `xml:",chardata"`      // 8 bytes (int)
	} `xml:"vcpu"`
}

func TestParseXML(t *testing.T) {
	// Test parsing XML into struct
	var domain TestDomain
	err := ParseXML([]byte(testXML), &domain)
	if err != nil {
		t.Fatalf("ParseXML failed: %v", err)
	}

	// Verify the parsed values
	if domain.Type != "kvm" {
		t.Errorf("Expected Type to be 'kvm', got '%s'", domain.Type)
	}

	if domain.Name != testVMName {
		t.Errorf("Expected Name to be '%s', got '%s'", testVMName, domain.Name)
	}

	if domain.UUID != "12345678-1234-1234-1234-123456789012" {
		t.Errorf("Expected UUID to be '12345678-1234-1234-1234-123456789012', got '%s'", domain.UUID)
	}

	if domain.Memory.Value != 2097152 {
		t.Errorf("Expected Memory.Value to be 2097152, got %d", domain.Memory.Value)
	}

	if domain.Memory.Unit != "KiB" {
		t.Errorf("Expected Memory.Unit to be 'KiB', got '%s'", domain.Memory.Unit)
	}

	if domain.VCPU.Value != 2 {
		t.Errorf("Expected VCPU.Value to be 2, got %d", domain.VCPU.Value)
	}

	// Test parsing invalid XML
	err = ParseXML([]byte("invalid XML"), &domain)
	if err == nil {
		t.Errorf("Expected error when parsing invalid XML, got nil")
	}
}

func TestParseXMLFile(t *testing.T) {
	// Create a temporary XML file
	tmpDir := t.TempDir()
	xmlFilePath := filepath.Join(tmpDir, "test.xml")
	err := os.WriteFile(xmlFilePath, []byte(testXML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test XML file: %v", err)
	}

	// Test parsing XML file
	var domain TestDomain
	err = ParseXMLFile(xmlFilePath, &domain)
	if err != nil {
		t.Fatalf("ParseXMLFile failed: %v", err)
	}

	// Verify the parsed values
	if domain.Name != testVMName {
		t.Errorf("Expected Name to be '%s', got '%s'", testVMName, domain.Name)
	}

	// Test parsing non-existent file
	err = ParseXMLFile("/nonexistent/file.xml", &domain)
	if err == nil {
		t.Errorf("Expected error when parsing non-existent file, got nil")
	}
}

func TestETreeFunctions(t *testing.T) {
	// Load XML document
	doc, err := LoadXMLDocumentFromString(testXML)
	if err != nil {
		t.Fatalf("LoadXMLDocumentFromString failed: %v", err)
	}

	// Test FindElement
	element := FindElement(doc, "//domain/name")
	if element == nil {
		t.Fatalf("FindElement failed to find domain/name")
	}
	if GetElementText(element) != testVMName {
		t.Errorf("Expected element text to be '%s', got '%s'", testVMName, GetElementText(element))
	}

	// Test FindElement with non-existent path
	nonExistent := FindElement(doc, "//domain/nonexistent")
	if nonExistent != nil {
		t.Errorf("Expected nil when finding non-existent path, got element")
	}

	// Test FindElement for UUID
	uuidElement := FindElement(doc, "//domain/uuid")
	if uuidElement == nil {
		t.Fatalf("FindElement failed to find domain/uuid")
	}
	if GetElementText(uuidElement) != "12345678-1234-1234-1234-123456789012" {
		t.Errorf("Expected value to be '12345678-1234-1234-1234-123456789012', got '%s'", GetElementText(uuidElement))
	}

	// Test GetElementAttribute
	domainElement := FindElement(doc, "//domain")
	if domainElement == nil {
		t.Fatalf("FindElement failed to find domain")
	}
	attrValue := GetElementAttribute(domainElement, "type")
	if attrValue != "kvm" {
		t.Errorf("Expected attribute value to be 'kvm', got '%s'", attrValue)
	}

	// Test GetElementAttribute with non-existent attribute
	nonExistentAttr := GetElementAttribute(domainElement, "nonexistent")
	if nonExistentAttr != "" {
		t.Errorf("Expected empty string when getting non-existent attribute, got '%s'", nonExistentAttr)
	}

	// Test modifying element text
	nameElement := FindElement(doc, "//domain/name")
	if nameElement != nil {
		nameElement.SetText("modified-vm")
		if GetElementText(nameElement) != "modified-vm" {
			t.Errorf("Expected modified value to be 'modified-vm', got '%s'", GetElementText(nameElement))
		}
	}

	// Test SetElementAttribute
	if domainElement != nil {
		SetElementAttribute(domainElement, "modified", "true")
		modifiedAttr := GetElementAttribute(domainElement, "modified")
		if modifiedAttr != "true" {
			t.Errorf("Expected modified attribute to be 'true', got '%s'", modifiedAttr)
		}
	}
}

func TestLoadSaveXMLDocument(t *testing.T) {
	// Create a temporary XML file
	tmpDir := t.TempDir()
	inputXMLPath := filepath.Join(tmpDir, "input.xml")
	err := os.WriteFile(inputXMLPath, []byte(testXML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test XML file: %v", err)
	}

	// Load XML document from file
	doc, err := LoadXMLDocument(inputXMLPath)
	if err != nil {
		t.Fatalf("LoadXMLDocument failed: %v", err)
	}

	// Modify the document
	nameElement := FindElement(doc, "//domain/name")
	if nameElement != nil {
		nameElement.SetText("saved-vm")
	}

	// Save modified document
	outputXMLPath := filepath.Join(tmpDir, "output.xml")
	err = SaveXMLDocument(doc, outputXMLPath)
	if err != nil {
		t.Fatalf("SaveXMLDocument failed: %v", err)
	}

	// Load the saved document
	savedDoc, err := LoadXMLDocument(outputXMLPath)
	if err != nil {
		t.Fatalf("LoadXMLDocument failed for saved document: %v", err)
	}

	// Verify the modification
	savedNameElement := FindElement(savedDoc, "//domain/name")
	if savedNameElement == nil {
		t.Fatalf("Could not find name element in saved document")
	}
	savedValue := GetElementText(savedNameElement)
	if savedValue != "saved-vm" {
		t.Errorf("Expected saved value to be 'saved-vm', got '%s'", savedValue)
	}
}

func TestXMLToString(t *testing.T) {
	// Load XML document
	doc, err := LoadXMLDocumentFromString(testXML)
	if err != nil {
		t.Fatalf("LoadXMLDocumentFromString failed: %v", err)
	}

	// Convert to string
	xmlString := XMLToString(doc)

	// Check that the string contains expected elements
	if !strings.Contains(xmlString, "<name>test-vm</name>") {
		t.Errorf("XMLToString result doesn't contain expected content")
	}

	// Verify it's properly indented
	if !strings.Contains(xmlString, "  <name>") {
		t.Errorf("XMLToString result doesn't appear to be properly indented")
	}
}

func TestPrettyPrintXML(t *testing.T) {
	// Create unformatted XML (no indentation)
	unformatted := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<domain type="kvm"><name>test-vm</name><memory unit="KiB">2097152</memory><vcpu placement="static">2</vcpu></domain>`)

	// Pretty print
	formatted, err := PrettyPrintXML(unformatted)
	if err != nil {
		t.Fatalf("PrettyPrintXML failed: %v", err)
	}

	// Verify it's properly indented
	formattedStr := string(formatted)
	if !strings.Contains(formattedStr, "  <name>") {
		t.Errorf("PrettyPrintXML result doesn't appear to be properly indented: %s", formattedStr)
	}

	// Test with invalid XML - use clearly malformed XML
	_, err = PrettyPrintXML([]byte("<root><unclosed>"))
	if err == nil {
		t.Errorf("Expected error when pretty printing invalid XML, got nil")
	}
}
