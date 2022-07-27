#!/bin/sh

cat << EOF
// Code generated by "gen_main_test.sh"; DO NOT EDIT.
package test

import (
	"testing"
	_ "embed"
)
EOF

function test_case_call {
	echo "tc := testCase(b, \"$1\", source_$1)"
}

function gen_test_case {
	name="$1"
	path="$2"
cat << EOF

// Test case $name ($path)
//go:embed $path
var source_$name string

func BenchmarkGen${name}_IS(b *testing.B) {
	$(test_case_call "$name")
	tc.Run(b, TestCaseConfig{Engine: EngineInterp, Parallel: false})
}

func BenchmarkGen${name}_IP(b *testing.B) {
	$(test_case_call "$name")
	tc.Run(b, TestCaseConfig{Engine: EngineInterp, Parallel: true})
}

func BenchmarkGen${name}_VS(b *testing.B) {
	$(test_case_call "$name")
	tc.Run(b, TestCaseConfig{Engine: EngineVM, Parallel: false})
}

func BenchmarkGen${name}_VP(b *testing.B) {
	$(test_case_call "$name")
	tc.Run(b, TestCaseConfig{Engine: EngineVM, Parallel: true})
}

func TestGen${name}_IS(t *testing.T) {
	b := t
	$(test_case_call "$name")
	tc.Test(t, TestCaseConfig{Engine: EngineInterp, Parallel: false})
}

func TestGen${name}_IP(t *testing.T) {
	b := t
	$(test_case_call "$name")
	tc.Test(t, TestCaseConfig{Engine: EngineInterp, Parallel: true})
}

func TestGen${name}_VS(t *testing.T) {
	b := t
	$(test_case_call "$name")
	tc.Test(t, TestCaseConfig{Engine: EngineVM, Parallel: false})
}

func TestGen${name}_VP(t *testing.T) {
	b := t
	$(test_case_call "$name")
	tc.Test(t, TestCaseConfig{Engine: EngineVM, Parallel: true})
}
EOF
}

for path in $(ls -1 tests/*.coa); do
	name="${path#*/}"
	name="${name%.coa}"
	gen_test_case "$name" "$path"
done