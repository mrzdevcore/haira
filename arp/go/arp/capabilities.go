package arp

// DefaultComponents returns the standard built-in ARP component type names.
func DefaultComponents() []string {
	return []string{
		"text",
		"status-card",
		"table",
		"code-block",
		"diff",
		"key-value",
		"progress",
		"chart",
		"form",
		"confirm",
		"choices",
		"product-cards",
		"markdown",
		"image",
	}
}

// DefaultCapabilities returns a standard ArpHello with all built-in
// components and default features.
func DefaultCapabilities() ArpHello {
	return ArpHello{
		V:    1,
		Type: TypeHello,
		Capabilities: ArpCapabilities{
			Components: DefaultComponents(),
			Features:   []string{"streaming", "input"},
		},
	}
}

// NewCapabilities creates an ArpHello with custom components and features.
func NewCapabilities(components []string, features []string) ArpHello {
	return ArpHello{
		V:    1,
		Type: TypeHello,
		Capabilities: ArpCapabilities{
			Components: components,
			Features:   features,
		},
	}
}
