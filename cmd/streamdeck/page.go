package main

func (p page) GetKeyDefinitions(cfg config) map[int]keyDefinition {
	var (
		defMaps []map[int]keyDefinition
		result  = make(map[int]keyDefinition)
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
