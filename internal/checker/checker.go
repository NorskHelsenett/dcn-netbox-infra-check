package checker

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/NorskHelsenett/dcn-netbox-infra-check/internal/config"
	"github.com/NorskHelsenett/dcn-netbox-infra-check/internal/models"
)

// Result holds the check results for a DC
type Result struct {
	DCName             string
	Infra              string
	Output             string
	HasMismatches      bool
	MovedVLANs         []MovedVLAN
	MisconfiguredVLANs []models.NAMVxLAN
	NameMismatches     []models.NAMVxLAN
	WrongPrefixes      []WrongPrefix
}

// MovedVLAN represents a VLAN that was moved but not updated
type MovedVLAN struct {
	VxLAN      models.NAMVxLAN
	NetboxVLAN models.NetboxVLAN
}

// WrongPrefix represents a prefix with incorrect infra
type WrongPrefix struct {
	VLAN   models.NAMVxLAN
	Prefix models.NetboxPrefix
}

// Check performs all VLAN checks for a given DC
func Check(
	dcName string,
	infra string,
	netboxVLANs []models.NetboxVLAN,
	netboxPrefixes []models.NetboxPrefix,
	namVxLANs []models.NAMVxLAN,
	config *config.Config,
) *Result {
	result := &Result{
		DCName: dcName,
		Infra:  infra,
	}

	// Filter VxLANs for this DC
	dcVxLANs := filterDCVxLANs(namVxLANs, dcName)

	// Filter VLANs for this infra
	infraVLANs := filterInfraVLANs(netboxVLANs, infra)

	// Perform checks
	result.MovedVLANs = checkMovedVLANs(dcVxLANs, infraVLANs)
	result.MisconfiguredVLANs = checkMisconfiguredVLANs(dcVxLANs, infraVLANs, infra)
	result.NameMismatches = checkNameMismatches(dcVxLANs, infraVLANs, result.MisconfiguredVLANs)
	result.WrongPrefixes = checkWrongPrefixes(dcVxLANs, netboxPrefixes, infra)

	// Set HasMismatches before generating output
	result.HasMismatches = len(result.MovedVLANs) > 0 ||
		len(result.MisconfiguredVLANs) > 0 ||
		len(result.NameMismatches) > 0 ||
		len(result.WrongPrefixes) > 0

	// Generate output
	result.Output = generateOutput(result, config)

	return result
}

// filterDCVxLANs filters VxLANs for a specific DC
func filterDCVxLANs(vxlans []models.NAMVxLAN, dcName string) []models.NAMVxLAN {
	var filtered []models.NAMVxLAN
	for _, vxlan := range vxlans {
		if vxlan.GetContainerName() == dcName {
			filtered = append(filtered, vxlan)
		}
	}
	return filtered
}

// filterInfraVLANs filters VLANs for a specific infra
func filterInfraVLANs(vlans []models.NetboxVLAN, infra string) []models.NetboxVLAN {
	var filtered []models.NetboxVLAN
	for _, vlan := range vlans {
		if vlan.GetInfra() == infra {
			filtered = append(filtered, vlan)
		}
	}
	return filtered
}

// checkMovedVLANs finds VLANs moved to nam-03 but not updated in NAM
func checkMovedVLANs(dcVxLANs []models.NAMVxLAN, infraVLANs []models.NetboxVLAN) []MovedVLAN {
	var moved []MovedVLAN
	for _, vxlan := range dcVxLANs {
		for _, vlan := range infraVLANs {
			if strings.Contains(vlan.Name, "nam-03") &&
				vlan.VID == vxlan.ID &&
				normalizeName(vxlan.Name) == normalizeName(strings.Replace(vlan.Name, "nam-03", "nam-01", -1)) {
				moved = append(moved, MovedVLAN{
					VxLAN:      vxlan,
					NetboxVLAN: vlan,
				})
			}
		}
	}
	return moved
}

// checkMisconfiguredVLANs finds VxLANs missing or misconfigured in Netbox
func checkMisconfiguredVLANs(dcVxLANs []models.NAMVxLAN, infraVLANs []models.NetboxVLAN, infra string) []models.NAMVxLAN {
	var misconfigured []models.NAMVxLAN
	for _, vxlan := range dcVxLANs {
		found := false
		for _, vlan := range infraVLANs {
			if vlan.VID == vxlan.ID && vlan.GetInfra() == infra {
				found = true
				break
			}
		}
		if !found {
			misconfigured = append(misconfigured, vxlan)
		}
	}
	return misconfigured
}

// checkNameMismatches finds VxLANs with name mismatches
func checkNameMismatches(dcVxLANs []models.NAMVxLAN, infraVLANs []models.NetboxVLAN, misconfigured []models.NAMVxLAN) []models.NAMVxLAN {
	var mismatches []models.NAMVxLAN

	// Create a map of misconfigured VLANs for quick lookup
	misconfiguredMap := make(map[int]bool)
	for _, v := range misconfigured {
		misconfiguredMap[v.ID] = true
	}

	for _, vxlan := range dcVxLANs {
		// Skip if already in misconfigured list
		if misconfiguredMap[vxlan.ID] {
			continue
		}

		found := false
		for _, vlan := range infraVLANs {
			if vlan.VID == vxlan.ID && normalizeName(vxlan.Name) == normalizeName(vlan.Name) {
				found = true
				break
			}
		}
		if !found {
			mismatches = append(mismatches, vxlan)
		}
	}
	return mismatches
}

// checkWrongPrefixes finds prefixes with wrong infra setting
func checkWrongPrefixes(dcVxLANs []models.NAMVxLAN, prefixes []models.NetboxPrefix, infra string) []WrongPrefix {
	var wrong []WrongPrefix
	for _, vxlan := range dcVxLANs {
		for _, prefix := range prefixes {
			if prefix.VLAN != nil &&
				prefix.VLAN.VID == vxlan.ID &&
				prefix.VLAN.Name == vxlan.Name &&
				prefix.GetInfra() != infra {
				wrong = append(wrong, WrongPrefix{
					VLAN:   vxlan,
					Prefix: prefix,
				})
			}
		}
	}
	return wrong
}

// normalizeName normalizes a name for comparison
func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// generateOutput creates formatted output text
func generateOutput(result *Result, config *config.Config) string {
	var buf bytes.Buffer

	if len(result.MovedVLANs) > 0 {
		buf.WriteString(strings.Repeat("=", 75))
		buf.WriteString("\n")
		buf.WriteString(fmt.Sprintf("Vxlans i '%s' som ikke er oppdatert i NAM etter flytting til nam-03 for '%s'\n", result.DCName, result.Infra))
		buf.WriteString(strings.Repeat("=", 75))
		buf.WriteString("\n")
		for _, mv := range result.MovedVLANs {
			buf.WriteString(fmt.Sprintf("✗ [NAM VLAN ID %d] Netbox='%s' -> NAM='%s'\n",
				mv.VxLAN.ID, mv.NetboxVLAN.Name, mv.VxLAN.Name))
		}
		buf.WriteString("\n")
	}

	if len(result.MisconfiguredVLANs) > 0 {
		buf.WriteString(strings.Repeat("=", 75))
		buf.WriteString("\n")
		buf.WriteString(fmt.Sprintf("Vxlans i '%s' som mangler eller ikke er registrert som '%s' i Netbox (%s)\n", result.DCName, result.Infra, config.NetboxURL))
		buf.WriteString(strings.Repeat("=", 75))
		buf.WriteString("\n")
		for _, vxlan := range result.MisconfiguredVLANs {
			buf.WriteString(fmt.Sprintf("✗ [NAM VLAN ID %d]: -> %s\n", vxlan.ID, vxlan.Name))
		}
		buf.WriteString("\n")
	}

	if len(result.NameMismatches) > 0 {
		buf.WriteString(strings.Repeat("=", 75))
		buf.WriteString("\n")
		buf.WriteString(fmt.Sprintf("Vxlans i '%s' som ikke har samme navn i Netbox (%s)\n", result.DCName, config.NetboxURL))
		buf.WriteString(strings.Repeat("=", 75))
		buf.WriteString("\n")
		for _, vxlan := range result.NameMismatches {
			buf.WriteString(fmt.Sprintf("✗ [NAM VLAN ID %d]: -> %s\n", vxlan.ID, vxlan.Name))
		}
		buf.WriteString("\n")
	}

	if len(result.WrongPrefixes) > 0 {
		buf.WriteString(strings.Repeat("=", 75))
		buf.WriteString("\n")
		buf.WriteString(fmt.Sprintf("Prefixes i '%s' som har feil 'infra i Netbox (%s)'\n", result.DCName, config.NetboxURL))
		buf.WriteString(strings.Repeat("=", 75))
		buf.WriteString("\n")
		for _, wp := range result.WrongPrefixes {
			buf.WriteString(fmt.Sprintf("✗ [NAM VLAN ID %d] -> %s har 'infra' = '%s'\n",
				wp.VLAN.ID, wp.Prefix.Prefix, wp.Prefix.GetInfra()))
		}
		buf.WriteString("\n")
	}

	if !result.HasMismatches {
		buf.WriteString("✓ Ingen avvik funnet!\n")
	}

	return buf.String()
}
