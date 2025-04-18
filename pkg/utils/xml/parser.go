package xml

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"

	"github.com/beevik/etree"
)

// ParseXML parses XML data into a structured object
func ParseXML(data []byte, v interface{}) error {
	return xml.Unmarshal(data, v)
}

// ParseXMLFile parses an XML file into a structured object
func ParseXMLFile(filePath string, v interface{}) error {
	// Read the file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read XML file %s: %w", filePath, err)
	}

	// Parse the XML
	return ParseXML(data, v)
}

// GetElementByXPath retrieves an XML element using XPath
func GetElementByXPath(doc *etree.Document, xpath string) (*etree.Element, error) {
	elements := doc.FindElements(xpath)
	if len(elements) == 0 {
		return nil, fmt.Errorf("no elements found for XPath: %s", xpath)
	}
	return elements[0], nil
}

// LoadXMLDocument loads an XML document from a file
func LoadXMLDocument(filePath string) (*etree.Document, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(filePath); err != nil {
		return nil, fmt.Errorf("failed to read XML document %s: %w", filePath, err)
	}
	return doc, nil
}

// LoadXMLDocumentFromString loads an XML document from a string
func LoadXMLDocumentFromString(xml string) (*etree.Document, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xml); err != nil {
		return nil, fmt.Errorf("failed to parse XML string: %w", err)
	}
	return doc, nil
}

// GetElementValue gets the text value of an XML element
func GetElementValue(doc *etree.Document, xpath string) (string, error) {
	element, err := GetElementByXPath(doc, xpath)
	if err != nil {
		return "", err
	}
	return element.Text(), nil
}

// GetElementAttribute gets the value of an attribute from an XML element
func GetElementAttribute(doc *etree.Document, xpath string, attribute string) (string, error) {
	element, err := GetElementByXPath(doc, xpath)
	if err != nil {
		return "", err
	}

	attr := element.SelectAttr(attribute)
	if attr == nil {
		return "", fmt.Errorf("attribute %s not found on element %s", attribute, xpath)
	}

	return attr.Value, nil
}

// SetElementValue sets the text value of an XML element
func SetElementValue(doc *etree.Document, xpath string, value string) error {
	element, err := GetElementByXPath(doc, xpath)
	if err != nil {
		return err
	}

	element.SetText(value)
	return nil
}

// SetElementAttribute sets an attribute value on an XML element
func SetElementAttribute(doc *etree.Document, xpath string, attribute string, value string) error {
	element, err := GetElementByXPath(doc, xpath)
	if err != nil {
		return err
	}

	element.CreateAttr(attribute, value)
	return nil
}

// SaveXMLDocument saves an XML document to a file
func SaveXMLDocument(doc *etree.Document, filePath string) error {
	return doc.WriteToFile(filePath)
}

// XMLToString converts an XML document to a string
func XMLToString(doc *etree.Document) string {
	doc.Indent(2)
	xmlString, _ := doc.WriteToString()
	return xmlString
}

// PrettyPrintXML formats XML data with proper indentation
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
