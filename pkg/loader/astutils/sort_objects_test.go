package astutils_test

import (
	"go/ast"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/jimmidyson/kube-client-gen/pkg/loader/astutils"
)

var sortedObjects = []*ast.Object{
	{Name: "1", Decl: &ast.Field{Names: []*ast.Ident{{Name: "1", NamePos: 1}}}},
	{Name: "2", Decl: &ast.TypeSpec{Name: &ast.Ident{Name: "2", NamePos: 2}}},
	{Name: "3", Decl: &ast.AssignStmt{Lhs: []ast.Expr{&ast.Ident{Name: "3", NamePos: 3}}}},
	{Name: "4", Decl: &ast.Field{Names: []*ast.Ident{{Name: "4", NamePos: 4}}}},
}

var unsortedMap = map[string]*ast.Object{
	"3": sortedObjects[2],
	"1": sortedObjects[0],
	"4": sortedObjects[3],
	"2": sortedObjects[1],
}

var _ = Describe("SortObjects", func() {
	Context("when there are no objects", func() {
		It("should return empty slice", func() {
			Expect(SortObjectsByPos(nil)).Should(BeEmpty())
			Expect(SortObjectsByPos(map[string]*ast.Object{})).Should(BeEmpty())
		})
	})

	DescribeTable("sorted objects",
		func(unsorted map[string]*ast.Object, expected []*ast.Object) {
			postSort := SortObjectsByPos(unsorted)
			Expect(postSort).Should(HaveLen(len(expected)))
			for i, cg := range postSort {
				Expect(cg).Should(Equal(expected[i]))
			}
		},
		Entry("single entry object", map[string]*ast.Object{"3": sortedObjects[2]}, sortedObjects[2:3]),
		Entry("multiple entry objects", unsortedMap, sortedObjects),
	)
})
