package sdiffer

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"reflect"
	"regexp"
	"testing"
)

type Person struct {
	Name    string
	Age     int
	Loc     *Location
	StrArr  []string
	Parents []*Person
}

type Location struct {
	Name     string
	Province *Location
}

type Building struct {
	BuildingMap map[string]string
}

func newLoc(name string) *Location {
	return &Location{Name: name}
}

type DiffTestSuite struct {
	suite.Suite
}

func TestDiff(t *testing.T) {
	suite.Run(t, new(DiffTestSuite))
}

func (*DiffTestSuite) Test() {
	me := &Person{
		Name:   "sjl",
		Age:    20,
		Loc:    &Location{"Ji'An", newLoc("JiangXi")},
		StrArr: []string{"hello", "world", "hi"},
	}
	he := &Person{
		Name:   "kxc",
		Age:    21,
		Loc:    nil,
		StrArr: []string{"world", "hello", "hi"},
	}
	me.Parents = append(me.Parents, &Person{Name: "me father", Parents: []*Person{{Name: "me grandFather"}}})
	he.Parents = append(he.Parents, &Person{Name: "he father", Parents: []*Person{{Name: "he grandFather"}}})
	differ := NewDiffer().Compare(me, he)
	fmt.Println(differ.String())
	for _, diff := range differ.Diffs() {
		fmt.Println(diff.Tag())
	}
}

func (suite *DiffTestSuite) TestType() {
	arr1 := []int{1, 2, 3, 4, 5}
	arr2 := []int64{5, 4, 3, 2, 1}
	suite.True(allowPanic(func() {
		fmt.Println(NewDiffer().Compare(arr1, arr2).String())
	}))
}

func (suite *DiffTestSuite) TestSlice() {
	arr1 := []int{1, 2, 3, 4, 5}
	arr2 := []int{5, 4, 3, 2, 1}
	fmt.Println(NewDiffer().Compare(arr1, arr2).String())
}

func (suite *DiffTestSuite) TestMap() {
	b1 := &Building{map[string]string{"1": "1", "2": "2"}}
	b2 := &Building{map[string]string{"1": "2", "2": "1"}}
	fmt.Println(NewDiffer().Compare(b1, b2).String())
}

func (suite *DiffTestSuite) TestIgnore() {
	b1 := &Building{map[string]string{"1": "1", "2": "2"}}
	b2 := &Building{map[string]string{"1": "2", "2": "1"}}
	fmt.Println(NewDiffer().Ignore("Building.BuildingMap[[0-9]*]").Compare(b1, b2).String())
}

func (suite *DiffTestSuite) TestTag() {
	b1 := &Building{map[string]string{"1": "1", "2": "2"}}
	b2 := &Building{map[string]string{"1": "2", "2": "1"}}
	differ := NewDiffer().Compare(b1, b2)
	fmt.Println(differ.String())
	diffs := differ.Diffs()
	for _, diff := range diffs {
		fmt.Println(diff.Tag())
	}
}

type parentsComparator struct{}

func (*parentsComparator) Match(path string) bool {
	return path == "Person.Parents"
}

func (*parentsComparator) Equals(a, b interface{}) (dt DiffType, msgA, msgB interface{}) {
	pa, pb := a.([]*Person), b.([]*Person)
	if len(pa) != len(pb) {
		return LengthDiff, len(pa), len(pb)
	}
	if pa[0].Name != pb[0].Name {
		return ElemDiff, "hello", "world"
	}
	return NoDiff, nil, nil
}

func (suite *DiffTestSuite) TestComparator() {
	me := &Person{
		Name:   "sjl",
		Age:    20,
		Loc:    &Location{"Ji'An", newLoc("JiangXi")},
		StrArr: []string{"hello", "world", "hi"},
	}
	he := &Person{
		Name:   "kxc",
		Age:    21,
		Loc:    nil,
		StrArr: []string{"world", "hello", "hi"},
	}
	me.Parents = append(me.Parents, &Person{Name: "me father", Parents: []*Person{{Name: "me grandFather"}}})
	he.Parents = append(he.Parents, &Person{Name: "he father", Parents: []*Person{{Name: "he grandFather"}}})

	differ := NewDiffer().WithComparator(new(parentsComparator)).Compare(me, he)
	fmt.Println(differ.String())
}

func (suite *DiffTestSuite) TestTrimSpaces() {
	me := &Person{
		Name: "   sjl",
	}
	me2 := &Person{
		Name: " sjl     ",
	}
	differ := NewDiffer()
	println(differ.WithTrimSpace("Person.Name").Compare(me, me2).String())
	println(differ.Reset().WithTrim("Person.Name", " ").Compare(me, me2).String())
}

func (suite *DiffTestSuite) TestSort() {
	type Integer struct {
		val int
	}
	arr := []Integer{{5}, {4}, {3}, {2}, {1}}
	qsort(reflect.ValueOf(arr), func(a, b interface{}) bool {
		i, j := a.(Integer), b.(Integer)
		return i.val < j.val
	})
	fmt.Println(arr)
}

type pSorter struct {
	match *regexp.Regexp
}

func (ps *pSorter) Match(fieldPath string) bool {
	return ps.match.MatchString(fieldPath)
}

func (ps *pSorter) Less(a, b interface{}) bool {
	pa, pb := a.(*Person), b.(*Person)
	return pa.Age < pb.Age
}

func (suite *DiffTestSuite) TestDisorderedCompare() {
	me := &Person{
		Name:    "me",
		Age:     20,
		Parents: []*Person{{Name: "p1", Age: 30}, {Name: "p2", Age: 40}, {Name: "p3", Age: 45}},
	}
	he := &Person{
		Name:    "he",
		Age:     21,
		Parents: []*Person{{Name: "p2", Age: 40}, {Name: "p1", Age: 30}, {Name: "p3", Age: 50}},
	}

	differ := NewDiffer().Compare(me, he)
	_, ok := differ.FindDiff("Person.Name")
	suite.True(ok)
	_, ok = differ.FindDiff("Person.Age")
	suite.True(ok)
	dfs := differ.FindDiffFuzzily("Person.Parents[[0-9]+]")
	suite.NotZero(len(dfs))

	differ.Reset()
	differ.WithSorter(&pSorter{regexp.MustCompile("Person.Parents")}).Compare(me, he)
	suite.Equal(40, he.Parents[0].Age)
	_, ok = differ.FindDiff("Person.Parents[[0-9]+].Name")
	suite.False(ok)
}

func (suite *DiffTestSuite) TestChore() {
	// ...
}