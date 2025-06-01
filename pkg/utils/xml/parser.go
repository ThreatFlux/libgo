package xml

import (
	"encoding/xml"
	"fmt"
	"os"

	"github.com/beevik/etree"
)

// ParseXML parses XML data into a structured object.
func ParseXML(data []byte, v interface{}) error {
	return xml.Unmarshal(data, v)
}

// ParseXMLFile parses an XML file into a structured object.
func ParseXMLFile(filePath string, v interface{}) error {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read XML file %s: %w", filePath, err)
	}

	// Parse the XML
	return ParseXML(data, v)
}

// CreateXMLDocument creates a new XML document with root element.
func CreateXMLDocument(rootElement string) *etree.Document {
	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
	doc.CreateElement(rootElement)
	return doc
}

// AddElement adds an element to a parent with text content.
func AddElement(parent *etree.Element, name, text string) *etree.Element {
	elem := parent.CreateElement(name)
	if text != "" {
		elem.SetText(text)
	}
	return elem
}

// AddElementWithAttributes adds an element with attributes.
func AddElementWithAttributes(parent *etree.Element, name string, attrs map[string]string) *etree.Element {
	elem := parent.CreateElement(name)
	for key, value := range attrs {
		elem.CreateAttr(key, value)
	}
	return elem
}

// SetElementAttribute sets an attribute on an element.
func SetElementAttribute(elem *etree.Element, name, value string) {
	elem.CreateAttr(name, value)
}

// GetElementText gets the text content of an element.
func GetElementText(elem *etree.Element) string {
	return elem.Text()
}

// GetElementAttribute gets an attribute value from an element.
func GetElementAttribute(elem *etree.Element, name string) string {
	attr := elem.SelectAttr(name)
	if attr != nil {
		return attr.Value
	}
	return ""
}

// FindElement finds the first element matching the path.
func FindElement(doc *etree.Document, path string) *etree.Element {
	return doc.FindElement(path)
}

// FindElements finds all elements matching the path.
func FindElements(doc *etree.Document, path string) []*etree.Element {
	return doc.FindElements(path)
}

// ParseXMLFromString parses XML from a string.
func ParseXMLFromString(xmlString string, v interface{}) error {
	return xml.Unmarshal([]byte(xmlString), v)
}

// LoadXMLDocument loads an XML document from file.
func LoadXMLDocument(filePath string) (*etree.Document, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(filePath); err != nil {
		return nil, fmt.Errorf("failed to load XML document from %s: %w", filePath, err)
	}
	return doc, nil
}

// LoadXMLDocumentFromString loads an XML document from string.
func LoadXMLDocumentFromString(xmlString string) (*etree.Document, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xmlString); err != nil {
		return nil, fmt.Errorf("failed to load XML document from string: %w", err)
	}
	return doc, nil
}

// SaveXMLDocument saves an XML document to a file.
func SaveXMLDocument(doc *etree.Document, filePath string) error {
	return doc.WriteToFile(filePath)
}

// XMLToString converts an XML document to a string.
func XMLToString(doc *etree.Document) string {
	doc.Indent(2)
	xmlString, err := doc.WriteToString()
	if err != nil {
		return ""
	}
	return xmlString
}

// PrettyPrintXML formats XML data with proper indentation.
func PrettyPrintXML(xmlData []byte) ([]byte, error) {
	// Parse the XML
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(xmlData); err != nil {
		return nil, fmt.Errorf("failed to parse XML for pretty printing: %w", err)
	}

	// Set indentation
	doc.Indent(2)

	// Convert back to bytes
	return doc.WriteToBytes()
}
