package config

// GetKeyDefinitions returns the effective key map including underlay and overlay pages.
func (p Page) GetKeyDefinitions(cfg File) map[int]KeyDefinition {
	var (
		defMaps []map[int]KeyDefinition
		result  = make(map[int]KeyDefinition)
	)

	// First process underlay if defined
	if p.Underlay != "" {
		defMaps = append(defMaps, cfg.Pages[p.Underlay].Keys)
	}

	// Process current definition
	defMaps = append(defMaps, p.Keys)

	// Last process overlay if defined
	if p.Overlay != "" {
		defMaps = append(defMaps, cfg.Pages[p.Overlay].Keys)
	}

	// Assemble combination of keys
	for _, pageDef := range defMaps {
		for idx, kd := range pageDef {
			if kd.Display.Type == "" {
				continue
			}

			result[idx] = kd
		}
	}

	return result
}
