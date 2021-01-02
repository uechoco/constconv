package main

import "go/constant"

type Value struct {
	Name        string // The name with trimmed prefix
	PkgName     string // The specified package name of the constant.
	TypeName    string // The specified type name of the constant.
	RepTypeName string // The specified representation of the type name.

	Str      string        // The string representation given by the "go/constant" package.
	ExactStr string        // The exact string representation given by the "go/constant" package.
	Kind     constant.Kind // The kind of constant given by the "go/constant" package.
}

func (v Value) String() string {
	return v.Str
}

// IsBool returns true if the kind of value is constant.Bool
func (v Value) IsBool() bool {
	return v.Kind == constant.Bool
}

// IsString returns true if the kind of value is constant.String
func (v Value) IsString() bool {
	return v.Kind == constant.String
}

// IsInt returns true if the kind of value is constant.Int
func (v Value) IsInt() bool {
	return v.Kind == constant.Int
}

// IsFloat returns true if the kind of value is constant.Float
func (v Value) IsFloat() bool {
	return v.Kind == constant.Float
}

// IsComplex returns true if the kind of value is constant.Complex
func (v Value) IsComplex() bool {
	return v.Kind == constant.Complex
}