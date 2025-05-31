package xml

import (
	"encoding/xml"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

// Sample XML data for testing
const testXML = `<?xml version="1.0" encoding="UTF-8"?>
<domain type="kvm">
  <name>test-vm</name>
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

// Test structure for XML unmarshaling
type TestDomain struct {
	XMLName xml.Name `xml:"domain"`
	Type    string   `xml:"type,attr"`
	Name    string   `xml:"name"`
	UUID    string   `xml:"uuid"`
	Memory  struct {
		Value int    `xml:",chardata"`
		Unit  string `xml:"unit,attr"`
	} `xml:"memory"`
	VCPU struct {
		Value     int    `xml:",chardata"`
		Placement string `xml:"placement,attr"`
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

	if domain.Name != "test-vm" {
		t.Errorf("Expected Name to be 'test-vm', got '%s'", domain.Name)
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
	err := ioutil.WriteFile(xmlFilePath, []byte(testXML), 0644)
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
	if domain.Name != "test-vm" {
		t.Errorf("Expected Name to be 'test-vm', got '%s'", domain.Name)
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

	// Test GetElementByXPath
	element, err := GetElementByXPath(doc, "//domain/name")
	if err != nil {
		t.Fatalf("GetElementByXPath failed: %v", err)
	}
	if element.Text() != "test-vm" {
		t.Errorf("Expected element text to be 'test-vm', got '%s'", element.Text())
	}

	// Test GetElementByXPath with non-existent path
	_, err = GetElementByXPath(doc, "//domain/nonexistent")
	if err == nil {
		t.Errorf("Expected error when getting non-existent path, got nil")
	}

	// Test GetElementValue
	value, err := GetElementValue(doc, "//domain/uuid")
	if err != nil {
		t.Fatalf("GetElementValue failed: %v", err)
	}
	if value != "12345678-1234-1234-1234-123456789012" {
		t.Errorf("Expected value to be '12345678-1234-1234-1234-123456789012', got '%s'", value)
	}

	// Test GetElementAttribute
	attrValue, err := GetElementAttribute(doc, "//domain", "type")
	if err != nil {
		t.Fatalf("GetElementAttribute failed: %v", err)
	}
	if attrValue != "kvm" {
		t.Errorf("Expected attribute value to be 'kvm', got '%s'", attrValue)
	}

	// Test GetElementAttribute with non-existent attribute
	_, err = GetElementAttribute(doc, "//domain", "nonexistent")
	if err == nil {
		t.Errorf("Expected error when getting non-existent attribute, got nil")
	}

	// Test SetElementValue
	err = SetElementValue(doc, "//domain/name", "modified-vm")
	if err != nil {
		t.Fatalf("SetElementValue failed: %v", err)
	}
	modifiedValue, err := GetElementValue(doc, "//domain/name")
	if err != nil {
		t.Fatalf("GetElementValue failed after modification: %v", err)
	}
	if modifiedValue != "modified-vm" {
		t.Errorf("Expected modified value to be 'modified-vm', got '%s'", modifiedValue)
	}

	// Test SetElementAttribute
	err = SetElementAttribute(doc, "//domain", "modified", "true")
	if err != nil {
		t.Fatalf("SetElementAttribute failed: %v", err)
	}
	modifiedAttr, err := GetElementAttribute(doc, "//domain", "modified")
	if err != nil {
		t.Fatalf("GetElementAttribute failed after modification: %v", err)
	}
	if modifiedAttr != "true" {
		t.Errorf("Expected modified attribute to be 'true', got '%s'", modifiedAttr)
	}
}

func TestLoadSaveXMLDocument(t *testing.T) {
	// Create a temporary XML file
	tmpDir := t.TempDir()
	inputXMLPath := filepath.Join(tmpDir, "input.xml")
	err := ioutil.WriteFile(inputXMLPath, []byte(testXML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test XML file: %v", err)
	}

	// Load XML document from file
	doc, err := LoadXMLDocument(inputXMLPath)
	if err != nil {
		t.Fatalf("LoadXMLDocument failed: %v", err)
	}

	// Modify the document
	err = SetElementValue(doc, "//domain/name", "saved-vm")
	if err != nil {
		t.Fatalf("SetElementValue failed: %v", err)
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
	savedValue, err := GetElementValue(savedDoc, "//domain/name")
	if err != nil {
		t.Fatalf("GetElementValue failed for saved document: %v", err)
	}
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
