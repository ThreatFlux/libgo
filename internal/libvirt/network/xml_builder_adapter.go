package network

import (
	"github.com/threatflux/libgo/pkg/logger"
	"github.com/threatflux/libgo/pkg/utils/xml"
)

// XMLBuilderAdapter adapts a TemplateXMLBuilder to work with a TemplateLoader
type XMLBuilderAdapter struct {
	builder *TemplateXMLBuilder
	loader  *xml.TemplateLoader // not used, just for compatibility
	logger  logger.Logger
}

// TemplateXMLBuilderWithLoader creates a new XMLBuilderAdapter
// This provides compatibility with the expected signature while using the existing implementation
func TemplateXMLBuilderWithLoader(loader *xml.TemplateLoader, logger logger.Logger) *XMLBuilderAdapter {
	return &XMLBuilderAdapter{
		builder: NewTemplateXMLBuilder(logger),
		loader:  loader,
		logger:  logger,
	}
}

// BuildNetworkXML delegates to the underlying builder
func (a *XMLBuilderAdapter) BuildNetworkXML(name string, bridgeName string, cidr string, dhcp bool) (string, error) {
	return a.builder.BuildNetworkXML(name, bridgeName, cidr, dhcp)
}
