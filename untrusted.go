package fuzzreq

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// This file answers ONE question — what makes a parameter untrusted input —
// and fuzzreq.go answers what makes a declaration an entry point. The split is
// where yze/filesize's finding pointed, and it is the seam the package comment
// already describes.

// untrustedName describes which untrusted shape a parameter carries.
type untrustedName string

// untrustedParam reports the first parameter carrying a BARE untrusted type:
// []byte, string, or an io type that admits a stream. A named domain type is
// deliberate vocabulary and does not count.
func untrustedParam(pass *analysis.Pass, decl *ast.FuncDecl) (untrustedName, bool) {
	for _, field := range decl.Type.Params.List {
		if input, ok := untrustedType(pass.TypesInfo.TypeOf(field.Type)); ok {
			return input, true
		}
	}
	return "", false
}

// untrustedType classifies a bare untrusted parameter type.
//
// A VARIADIC parameter arrives here as the SLICE its body sees and is not
// judged by its element, so `raw ...byte` is []byte and reported while `raw
// ...string` is []string and is not.
//
// THAT IS A HOLE, not a property. Adding `...` to a string parameter costs one
// token, leaves every call site byte-identical, and takes the function out of
// the rule — the cheapest evasion in it. It is open because closing it was
// measured and costs more: judging the element adds 293 findings across 262
// fleet modules, 286 of them on `args ...domain.Argument`, the variadic string
// every domain Run takes because go-app's Runner contract requires it, whose
// author cannot make Argument a defined type without leaving that contract and
// so can answer only with a baseline. Neither side is right; what is needed is
// a discriminator, and there is none yet. Recorded as
// `fuzzreq.variadic-escape-needs-a-discriminator` so it stays a debt rather
// than a decision.
//
// The type is UNALIASED first: an alias declares no type, so `type Payload = []byte` is a
// spelling of []byte and carries none of the vocabulary the named-type
// exemption exists for — types.Identical(Payload, []byte) holds, which is why
// returnsError already sees through an alias of error.
func untrustedType(t types.Type) (untrustedName, bool) {
	switch resolved := coreType(types.Unalias(t)).(type) {
	case *types.Basic:
		if resolved.Kind() == types.String {
			return "string", true
		}
	case *types.Slice:
		if isByte(resolved.Elem()) {
			return "[]byte", true
		}
	case *types.Named:
		return ioStreamName(resolved)
	}
	return "", false
}

// constraintDepth bounds the walk through nested constraint interfaces. Go
// forbids an embedding cycle, so the chain is finite in any package that
// type-checked — the bound is what keeps a malformed or future representation
// from spinning, and TestCoreTypeStopsAtTheConstraintDepth pins both sides of
// it.
const constraintDepth = 4

// coreType is the one type a type parameter's constraint admits, or the type
// itself. `func Parse[T ~[]byte](raw T) error` is []byte at every call site and
// inference leaves every caller unchanged, so a type parameter is a spelling of
// what it constrains — the same reason an alias is. Every spelling of that one
// type resolves, because go/types represents them differently and an author
// choosing between them is choosing nothing: a union term (`~[]byte`), a bare
// embedded type (`[]byte`), an alias of one, and a named constraint interface
// embedding any of those.
//
// A constraint carrying a METHOD is not resolved. `interface{ ~[]byte; Len()
// int }` admits only defined types with a Len method — which is the deliberate
// vocabulary this rule exempts — and no caller can pass a bare []byte, so
// reporting it would fire where the reason does not hold.
func coreType(t types.Type) types.Type {
	param, ok := t.(*types.TypeParam)
	if !ok {
		return t
	}
	sole := param.Constraint()
	for range constraintDepth {
		iface, isInterface := sole.Underlying().(*types.Interface)
		if !isInterface {
			return types.Unalias(sole)
		}
		next, ok := soleEmbedded(iface)
		if !ok {
			return t
		}
		sole = next
	}
	return t
}

// soleEmbedded is the one type a constraint interface embeds, when it embeds
// exactly one and declares no method of its own. What it returns may still be
// an alias — `~Blob` is a union term over the alias itself — and coreType
// unaliases on the way back in rather than here, so there is one place that
// does it and a mutation of that place fails a case.
func soleEmbedded(iface *types.Interface) (types.Type, bool) {
	if iface.NumEmbeddeds() != 1 || iface.NumExplicitMethods() != 0 {
		return nil, false
	}
	embedded := types.Unalias(iface.EmbeddedType(0))
	union, isUnion := embedded.(*types.Union)
	if !isUnion {
		return embedded, true
	}
	if union.Len() != 1 {
		return nil, false
	}
	return union.Term(0).Type(), true
}

// isByte reports the byte element type, unaliased for the same reason
// untrustedType unaliases: `type B = byte` makes []B a spelling of []byte.
func isByte(t types.Type) bool {
	basic, ok := types.Unalias(t).(*types.Basic)
	return ok && basic.Kind() == types.Byte
}

// ioStreamName reports a type DECLARED IN PACKAGE io that admits an untrusted
// stream — io.Reader and every interface embedding it, io.ReadCloser
// and io.ReadSeeker among them. Reading the name alone exempted them, and
// widening a parameter from Reader to ReadCloser is one word that changes
// nothing about what the function consumes.
func ioStreamName(named *types.Named) (untrustedName, bool) {
	obj := named.Obj()
	if obj.Pkg() == nil || obj.Pkg().Path() != "io" {
		return "", false
	}
	reader := obj.Pkg().Scope().Lookup("Reader")
	if reader == nil {
		return "", false
	}
	stream, ok := reader.Type().Underlying().(*types.Interface)
	if !ok || !types.Implements(named, stream) {
		return "", false
	}
	return untrustedName("io." + obj.Name()), true
}
