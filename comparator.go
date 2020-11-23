package sdiffer

type DiffType int

const (
	// LengthDiff should be returned by Comparator.Equals when length of two Arrays or Slices are different.
	LengthDiff DiffType = iota

	// NilDiff should be returned by Comparator.Equals when only one of two elements is nil.
	NilDiff

	// ElemDiff should be returned by Comparator.Equals when exact value of two elements are different.
	ElemDiff

	// NoDiff should be returned by Comparator.Equals when two elements are equal.
	NoDiff
)

// Comparator customized field comparator.
type Comparator interface {

	// Match checks if a field should use this comparator.
	Match(fieldPath string) bool

	// Equals compares two interfaces and return a DiffType among LengthDiff, NilDiff, ElemDiff
	// and NoDiff, or else Differ will throw a panic.
	//
	// Attention:
	// msgA, msgB represent the diff msg you want to record when dt is ElemDiff, which means
	// once you trying to use a customized Comparator, you have to build your own diff message
	// and take over all the compare work for the sub-fields of the two interfaces.
	// See Differ.Compare for more details.
	Equals(a, b interface{}) (dt DiffType, msgA, msgB interface{})
}
