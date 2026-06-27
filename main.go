// Command pybackendplugin is a JOBS build plugin: it reads {call:{pyproject:<bytes>}}
// on stdin, resolves the sdist's declared build backend against the curated
// build-system set, and writes the pinned backend-wheel import specs on stdout.
// Network-free, statically linked. (Slice 2b-i.)
package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/jobs-build/plugin-pybackend/internal/importdef"
	"github.com/fxamacker/cbor/v2"
)

type request struct {
	Call   map[string]any `cbor:"call"`
	Source string         `cbor:"source"`
}
type inputSpec struct {
	Kind       string `cbor:"kind"`
	Definition []byte `cbor:"definition"`
}
type pkgOut struct {
	Name    string    `cbor:"name"`
	Version string    `cbor:"version"`
	Input   inputSpec `cbor:"input"`
}

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "pybackendplugin:", err)
		os.Exit(1)
	}
}

func run(stdin io.Reader, stdout io.Writer) error {
	in, err := io.ReadAll(stdin)
	if err != nil {
		return fmt.Errorf("read request: %w", err)
	}
	var req request
	if err := cbor.Unmarshal(in, &req); err != nil {
		return fmt.Errorf("decode request: %w", err)
	}
	var pp []byte
	switch v := req.Call["pyproject"].(type) {
	case []byte:
		pp = v
	case string:
		pp = []byte(v)
	case nil:
		pp = nil // absent -> legacy fallback
	default:
		return fmt.Errorf("pyproject kwarg not bytes/string (got %T)", req.Call["pyproject"])
	}
	wheels, err := backendsFor(pp, archToken(req.Call["platform"]))
	if err != nil {
		return err
	}
	out := make([]pkgOut, 0, len(wheels))
	for _, w := range wheels {
		spec, err := wheelInput(w)
		if err != nil {
			return fmt.Errorf("backend %s: %w", w.Name, err)
		}
		out = append(out, pkgOut{Name: w.Name, Version: w.Version, Input: spec})
	}
	resp, err := cbor.Marshal(out)
	if err != nil {
		return fmt.Errorf("encode response: %w", err)
	}
	_, err = stdout.Write(resp)
	return err
}

// archToken maps the recipe-supplied platform kwarg (e.g. "linux/arm64") to the
// musl/rust arch token ("x86_64" / "aarch64"). When the kwarg is absent it falls
// back to the plugin's own GOARCH — the plugin is compiled for and runs on the
// target platform, so its arch is authoritative. Unknown values pass through
// unchanged so backendsFor reports a clear "no maturin wheel for arch" error.
func archToken(platform any) string {
	goarch := runtime.GOARCH
	if s, ok := platform.(string); ok && s != "" {
		// platform is "<os>/<arch>"; take the arch half.
		if i := strings.LastIndex(s, "/"); i >= 0 {
			goarch = s[i+1:]
		} else {
			goarch = s
		}
	}
	switch goarch {
	case "amd64", "x86_64":
		return "x86_64"
	case "arm64", "aarch64":
		return "aarch64"
	default:
		return goarch
	}
}

func wheelInput(w pinnedWheel) (inputSpec, error) {
	params, err := importdef.CanonicalParams(map[string]any{
		"url": w.URL, "filename": path.Base(w.URL), "sha256": w.Sha256, "name": w.Name, "version": w.Version,
	})
	if err != nil {
		return inputSpec{}, err
	}
	def, err := importdef.Definition{Fetcher: "pypi", Params: params}.Canonical()
	if err != nil {
		return inputSpec{}, err
	}
	return inputSpec{Kind: "import", Definition: def}, nil
}
