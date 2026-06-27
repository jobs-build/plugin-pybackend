package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

type pyproject struct {
	BuildSystem buildSystem `toml:"build-system"`
}
type buildSystem struct {
	Requires []string `toml:"requires"`
}

// normalize applies PEP 503 name normalization (lowercase; runs of -_. -> -).
func normalize(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	prevDash := false
	for _, r := range name {
		if r == '-' || r == '_' || r == '.' {
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
			continue
		}
		b.WriteRune(r)
		prevDash = false
	}
	return b.String()
}

// reqName extracts the PEP 503-normalized package name from a PEP 508 spec
// (e.g. "hatchling>=1.0 ; python_version>='3.8'" -> "hatchling").
func reqName(spec string) string {
	for i, r := range spec {
		if strings.ContainsRune("<>=!~ ;[(@", r) {
			return normalize(spec[:i])
		}
	}
	return normalize(spec)
}

// backendsFor parses pyproject's [build-system].requires and returns the
// resolved install closure (pinned wheels) from the curated set. An absent /
// empty [build-system] falls back to setuptools+wheel (PEP 517 legacy). A
// declared backend not in the curated set is a hard error. arch is the musl/rust
// arch token ("x86_64" / "aarch64") used to select the architecture-specific
// maturin wheel; pure-Python backends are arch-independent.
func backendsFor(pyprojectTOML []byte, arch string) ([]pinnedWheel, error) {
	var pp pyproject
	if len(pyprojectTOML) > 0 {
		if err := toml.Unmarshal(pyprojectTOML, &pp); err != nil {
			return nil, fmt.Errorf("parse pyproject.toml: %w", err)
		}
	}
	reqs := pp.BuildSystem.Requires
	if len(reqs) == 0 {
		reqs = []string{"setuptools", "wheel"} // PEP 517 legacy fallback
	}
	seen := map[string]bool{}
	var out []pinnedWheel
	for _, spec := range reqs {
		name := reqName(spec)
		closure, ok := curated[name]
		if !ok {
			return nil, fmt.Errorf("build backend %q is not in the curated build-system set (slice 2b-i supports a fixed set; regenerate curated.go to add it)", name)
		}
		for _, w := range closure {
			if seen[w.Name] {
				continue
			}
			seen[w.Name] = true
			// maturin is arch-specific (bundles a compiled binary): swap the
			// registry placeholder for the wheel matching the target arch.
			if w.Name == "maturin" {
				mw, ok := maturinByArch[arch]
				if !ok {
					return nil, fmt.Errorf("no maturin wheel for arch %q (supported: x86_64, aarch64)", arch)
				}
				w = mw
			}
			out = append(out, w)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
