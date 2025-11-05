package models

// NetboxVLAN represents a VLAN from Netbox
type NetboxVLAN struct {
	ID           int                    `json:"id"`
	VID          int                    `json:"vid"`
	Name         string                 `json:"name"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

// NetboxPrefix represents a prefix from Netbox
type NetboxPrefix struct {
	ID           int                    `json:"id"`
	Prefix       string                 `json:"prefix"`
	VLAN         *VLANReference         `json:"vlan"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

// VLANReference is a nested VLAN reference in a prefix
type VLANReference struct {
	ID   int    `json:"id"`
	VID  int    `json:"vid"`
	Name string `json:"name"`
}

// NAMVxLAN represents a VxLAN from NAM
type NAMVxLAN struct {
	ID         int         `json:"id"`
	Name       string      `json:"name"`
	Containers []Container `json:"containers"`
}

// Container represents a container in NAM VxLAN
type Container struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// NetboxResponse is the generic response structure from Netbox API
type NetboxResponse struct {
	Count   int         `json:"count"`
	Results interface{} `json:"results"`
}

// NAMResponse is the generic response structure from NAM API
type NAMResponse struct {
	Count   int         `json:"count"`
	Results interface{} `json:"results"`
}

// GetInfra safely extracts the infra custom field from Netbox objects
func (v *NetboxVLAN) GetInfra() string {
	if v.CustomFields != nil {
		if infra, ok := v.CustomFields["infra"].(string); ok {
			return infra
		}
	}
	return ""
}

// GetInfra safely extracts the infra custom field from Netbox prefix
func (p *NetboxPrefix) GetInfra() string {
	if p.CustomFields != nil {
		if infra, ok := p.CustomFields["infra"].(string); ok {
			return infra
		}
	}
	return ""
}

// GetContainerName returns the first container name if it exists
func (v *NAMVxLAN) GetContainerName() string {
	if len(v.Containers) > 0 {
		return v.Containers[0].Name
	}
	return ""
}
