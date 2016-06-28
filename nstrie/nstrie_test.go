// +build small

package nstrie

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTrie(t *testing.T) {
	Convey("Get and Put", t, func() {
		trie := New()

		trie.Put([]string{"foo", "bar", "baz", "qux"}, 1)
		trie.Put([]string{"foo", "bar", "baz"}, 2)
		trie.Put([]string{"foo", "bar", "qux", "baz"}, 3)

		Convey("test basic lookups", func() {
			val, ok := trie.Get([]string{"foo", "bar", "baz", "qux"})
			So(ok, ShouldBeTrue)
			So(val.(int), ShouldEqual, 1)
			val, ok = trie.Get([]string{"foo", "bar", "baz"})
			So(ok, ShouldBeTrue)
			So(val.(int), ShouldEqual, 2)
			val, ok = trie.Get([]string{"foo", "bar", "qux", "baz"})
			So(ok, ShouldBeTrue)
			So(val.(int), ShouldEqual, 3)
		})

		Convey("test overwriting of values", func() {
			trie.Put([]string{"foo", "bar", "baz"}, 4)
			val, ok := trie.Get([]string{"foo", "bar", "baz"})
			So(ok, ShouldBeTrue)
			So(val.(int), ShouldEqual, 4)
		})

		Convey("test that a node with no value returns ok: false", func() {
			_, ok := trie.Get([]string{"foo"})
			So(ok, ShouldBeFalse)
		})

		Convey("test that a namespace which exceeds the depth of the trie does not cause a bounds exceeded panic", func() {
			_, ok := trie.Get([]string{"foo", "bar", "baz", "qux", "foo", "bar", "baz"})
			So(ok, ShouldBeFalse)
		})

		Convey("test behavior against an empty string query", func() {
			_, ok := trie.Get([]string{""})
			So(ok, ShouldBeFalse)
		})

		Convey("test behavior against an empty query", func() {
			_, ok := trie.Get([]string{})
			So(ok, ShouldBeFalse)
		})
	})

	Convey("Fetch", t, func() {
		trie := New()

		Convey("test basic lookups", func() {
			trie.Put([]string{"foo", "bar", "baz", "qux"}, 1)
			trie.Put([]string{"foo", "bar", "baz"}, 2)
			trie.Put([]string{"foo", "bar", "qux", "baz"}, 3)

			childVals1, ok1 := trie.Fetch([]string{"foo", "bar"})
			fetchAssert(t, childVals1, ok1)

			childVals2, ok2 := trie.Fetch([]string{""})
			fetchAssert(t, childVals2, ok2)

			childVals3, ok3 := trie.Fetch([]string{})
			fetchAssert(t, childVals3, ok3)
		})

		Convey("test namespace not found", func() {
			_, err := trie.Fetch([]string{"foo", "baz"})
			So(err, ShouldNotBeNil)
		})

		Convey("test lookup at leaf", func() {
			trie.Put([]string{"foo", "bar", "baz", "qux"}, 1)
			leaf, ok := trie.Fetch([]string{"foo", "bar", "baz", "qux"})
			So(ok, ShouldBeTrue)
			So(len(leaf), ShouldEqual, 1)
			So(leaf[0].(int), ShouldEqual, 1)
		})
	})

	Convey("TestAndPut", t, func() {
		trie := New()

		Convey("test that it does not overwrite existing values and returns false", func() {
			trie.Put([]string{"foo", "bar", "baz"}, 1)
			ok := trie.TestAndPut([]string{"foo", "bar", "baz"}, 2)
			So(ok, ShouldBeFalse)
			val, ok := trie.Get([]string{"foo", "bar", "baz"})
			So(ok, ShouldBeTrue)
			So(val.(int), ShouldEqual, 1)
		})
		Convey("test that it does create new values and returns true", func() {
			trie.Put([]string{"foo", "bar", "baz"}, 1)
			ok := trie.TestAndPut([]string{"foo", "bar"}, 2)
			So(ok, ShouldBeTrue)
			val, ok := trie.Get([]string{"foo", "bar"})
			So(ok, ShouldBeTrue)
			So(val.(int), ShouldEqual, 2)
		})
	})

}

func fetchAssert(t *testing.T, childVals []interface{}, ok bool) {
	So(ok, ShouldBeTrue)
	So(len(childVals), ShouldEqual, 3)
	So(contains(childVals, 1), ShouldBeTrue)
	So(contains(childVals, 2), ShouldBeTrue)
	So(contains(childVals, 3), ShouldBeTrue)
}

func contains(slice []interface{}, val int) bool {
	for _, i := range slice {
		if i.(int) == val {
			return true
		}
	}
	return false
}
