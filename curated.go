package main

// pinnedWheel is one curated build-system wheel.
type pinnedWheel struct{ Name, Version, URL, Sha256 string }

// curated maps a normalized backend/build-dep name to its install closure
// (the package + its transitive build-time deps), all py3-none-any wheels,
// pinned by url+sha256. Generated at authoring time (see the plan); regenerate
// to add a backend. Names are PEP 503-normalized (lowercase, runs of -_. -> -).
//
// Generation commands (requires network access):
//
//	nix shell nixpkgs#python3Packages.pip -c pip download \
//	    --only-binary :all: setuptools wheel hatchling \
//	    -d /tmp/cur/ --verbose
//	# record URLs from verbose output + sha256sum of each .whl
var curated = map[string][]pinnedWheel{
	"setuptools": {
		{
			Name:    "setuptools",
			Version: "82.0.1",
			URL:     "https://files.pythonhosted.org/packages/9d/76/f789f7a86709c6b087c5a2f52f911838cad707cc613162401badc665acfe/setuptools-82.0.1-py3-none-any.whl",
			Sha256:  "a59e362652f08dcd477c78bb6e7bd9d80a7995bc73ce773050228a348ce2e5bb",
		},
	},
	"wheel": {
		{
			Name:    "wheel",
			Version: "0.47.0",
			URL:     "https://files.pythonhosted.org/packages/87/1b/9e33c09813d65e248f7f773119148a612516a4bea93e9c6f545f78455b7c/wheel-0.47.0-py3-none-any.whl",
			Sha256:  "212281cab4dff978f6cedd499cd893e1f620791ca6ff7107cf270781e587eced",
		},
		// wheel>=0.45 requires packaging>=24.0 at install time (PEP 517 build check).
		{
			Name:    "packaging",
			Version: "26.2",
			URL:     "https://files.pythonhosted.org/packages/df/b2/87e62e8c3e2f4b32e5fe99e0b86d576da1312593b39f47d8ceef365e95ed/packaging-26.2-py3-none-any.whl",
			Sha256:  "5fc45236b9446107ff2415ce77c807cee2862cb6fac22b8a73826d0693b0980e",
		},
	},
	"hatchling": {
		// hatchling + its runtime build deps (pathspec, pluggy, packaging, trove-classifiers)
		{
			Name:    "hatchling",
			Version: "1.30.1",
			URL:     "https://files.pythonhosted.org/packages/56/49/2797ec0ef88008a653a8867bb8d1e5c223cd2df8e40390dd5c6a0279cbc5/hatchling-1.30.1-py3-none-any.whl",
			Sha256:  "161eacafb3c6f91526e92116d21426369f2c36e98c36a864f11a96345ad4ee31",
		},
		{
			Name:    "packaging",
			Version: "26.2",
			URL:     "https://files.pythonhosted.org/packages/df/b2/87e62e8c3e2f4b32e5fe99e0b86d576da1312593b39f47d8ceef365e95ed/packaging-26.2-py3-none-any.whl",
			Sha256:  "5fc45236b9446107ff2415ce77c807cee2862cb6fac22b8a73826d0693b0980e",
		},
		{
			Name:    "pathspec",
			Version: "1.1.1",
			URL:     "https://files.pythonhosted.org/packages/f1/d9/7fb5aa316bc299258e68c73ba3bddbc499654a07f151cba08f6153988714/pathspec-1.1.1-py3-none-any.whl",
			Sha256:  "a00ce642f577bf7f473932318056212bc4f8bfdf53128c78bbd5af0b9b20b189",
		},
		{
			Name:    "pluggy",
			Version: "1.6.0",
			URL:     "https://files.pythonhosted.org/packages/54/20/4d324d65cc6d9205fabedc306948156824eb9f0ee1633355a8f7ec5c66bf/pluggy-1.6.0-py3-none-any.whl",
			Sha256:  "e920276dd6813095e9377c0bc5566d94c932c33b27a3e3945d8389c374dd4746",
		},
		{
			Name:    "trove-classifiers",
			Version: "2026.6.1.19",
			URL:     "https://files.pythonhosted.org/packages/7c/a4/81502f486f01db95bc8320646a8a12511f5e556cb63d5e224d91816605c4/trove_classifiers-2026.6.1.19-py3-none-any.whl",
			Sha256:  "ab4c4ec93cc4a4e7815fa759906e05e6bb3f2fbd92ea0f897288c6a43efd15b3",
		},
	},
	// maturin's closure is a single arch-specific wheel; the actual pin is chosen
	// from maturinByArch at resolve time (see backendsFor). The entry here only
	// registers "maturin" as a known backend — its placeholder pin is never emitted.
	"maturin": {
		{Name: "maturin", Version: "1.14.1"},
	},
}

// maturinByArch pins the maturin PEP-517 backend wheel per target arch. Unlike the
// pure-Python backends (py3-none-any), the maturin wheel bundles a compiled binary,
// so it is architecture-specific (musllinux_1_1_<arch>). Keyed by the musl/rust
// arch token ("x86_64" / "aarch64"), matching the recipe's platform mapping.
var maturinByArch = map[string]pinnedWheel{
	"x86_64": {
		Name:    "maturin",
		Version: "1.14.1",
		URL:     "https://files.pythonhosted.org/packages/1a/bd/9c0d5d6983905ce2c9edaa073a7e89355a9cf7f396988e05d32f1c37785d/maturin-1.14.1-py3-none-manylinux_2_12_x86_64.manylinux2010_x86_64.musllinux_1_1_x86_64.whl",
		Sha256:  "dfc54ae32e6fcb18302193ab9a30b0b25eefffba994ae13238974805533ef75e",
	},
	"aarch64": {
		Name:    "maturin",
		Version: "1.14.1",
		URL:     "https://files.pythonhosted.org/packages/e5/33/b096412bd6a7cb399652b260666f901adf88a687181a6dbd6a3f89f0a94e/maturin-1.14.1-py3-none-manylinux_2_17_aarch64.manylinux2014_aarch64.musllinux_1_1_aarch64.whl",
		Sha256:  "a131d912b5267e640bc96d70f4914e10590aed64082ec9abacba7cea52004224",
	},
}
