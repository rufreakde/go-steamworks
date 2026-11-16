// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The go-steamworks Authors

package steamworks

import (
	"debug/elf"
	"debug/macho"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"
)

// registeredNames reads the flatAPI_* constants of steamworks.go: every
// name registerFunctions asks the Steam library for.
func registeredNames(t *testing.T) []string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "steamworks.go", nil, 0)
	if err != nil {
		t.Fatalf("parse steamworks.go: %v", err)
	}
	var names []string
	ast.Inspect(file, func(n ast.Node) bool {
		spec, ok := n.(*ast.ValueSpec)
		if !ok {
			return true
		}
		for i, name := range spec.Names {
			if !strings.HasPrefix(name.Name, "flatAPI_") || i >= len(spec.Values) {
				continue
			}
			lit, ok := spec.Values[i].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				t.Fatalf("%s is not a string literal", name.Name)
			}
			value, err := strconv.Unquote(lit.Value)
			if err != nil {
				t.Fatalf("%s: %v", name.Name, err)
			}
			names = append(names, value)
		}
		return true
	})
	if len(names) < 30 {
		t.Fatalf("found only %d flatAPI_ names in steamworks.go", len(names))
	}
	return names
}

func missing(needed []string, has map[string]bool) []string {
	var lacks []string
	for _, name := range needed {
		if !has[name] {
			lacks = append(lacks, name)
		}
	}
	return lacks
}

// A library of another SDK version lacks names this binding registers, and
// the program then cannot use Steam at all. So the two libraries in this
// repository must hold every name. (The Windows DLL is not in this
// repository; the game that ships it checks it the same way.)
func TestTheMacOSLibraryHasEveryRegisteredName(t *testing.T) {
	fat, err := macho.OpenFat("libsteam_api.dylib")
	if err != nil {
		t.Fatalf("open libsteam_api.dylib: %v", err)
	}
	defer fat.Close()
	if len(fat.Arches) == 0 {
		t.Fatal("libsteam_api.dylib holds no architecture")
	}
	needed := registeredNames(t)
	for _, arch := range fat.Arches {
		if arch.Symtab == nil {
			t.Fatalf("%v: no symbol table", arch.Cpu)
		}
		has := map[string]bool{}
		for _, sym := range arch.Symtab.Syms {
			has[strings.TrimPrefix(sym.Name, "_")] = true
		}
		if lacks := missing(needed, has); len(lacks) > 0 {
			t.Errorf("libsteam_api.dylib (%v) lacks %d registered names: %v", arch.Cpu, len(lacks), lacks)
		}
	}
}

func TestTheLinuxLibraryHasEveryRegisteredName(t *testing.T) {
	file, err := elf.Open("libsteam_api64.so")
	if err != nil {
		t.Fatalf("open libsteam_api64.so: %v", err)
	}
	defer file.Close()
	symbols, err := file.DynamicSymbols()
	if err != nil {
		t.Fatalf("read the symbols of libsteam_api64.so: %v", err)
	}
	has := map[string]bool{}
	for _, sym := range symbols {
		has[sym.Name] = true
	}
	if lacks := missing(registeredNames(t), has); len(lacks) > 0 {
		t.Errorf("libsteam_api64.so lacks %d registered names: %v", len(lacks), lacks)
	}
}
