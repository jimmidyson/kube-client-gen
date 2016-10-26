package astutils_test

import (
	"go/ast"
	"go/token"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/jimmidyson/kube-client-gen/pkg/loader/astutils"
)

var sortedCommentGroup = []*ast.CommentGroup{
	&ast.CommentGroup{[]*ast.Comment{{Slash: token.Pos(1)}}},
	&ast.CommentGroup{[]*ast.Comment{{Slash: token.Pos(2)}}},
	&ast.CommentGroup{[]*ast.Comment{{Slash: token.Pos(3)}}},
	&ast.CommentGroup{[]*ast.Comment{{Slash: token.Pos(4)}}},
	&ast.CommentGroup{[]*ast.Comment{{Slash: token.Pos(5)}}},
}

var _ = Describe("SortComments", func() {
	Context("when there are no comment groups", func() {
		It("should return empty slice", func() {
			Expect(SortCommentsByPos(nil)).Should(BeEmpty())
			Expect(SortCommentsByPos([]*ast.CommentGroup{})).Should(BeEmpty())
		})
	})

	DescribeTable("sorted comment groups",
		func(actual []*ast.CommentGroup, expected []*ast.CommentGroup) {
			postSort := SortCommentsByPos(actual)
			Expect(postSort).Should(HaveLen(len(expected)))
			for i, cg := range postSort {
				Expect(cg).Should(Equal(expected[i]))
			}
		},
		Entry("single entry comment group", []*ast.CommentGroup{sortedCommentGroup[2]}, sortedCommentGroup[2:3]),
		Entry("multiple entry comment group", []*ast.CommentGroup{sortedCommentGroup[3], sortedCommentGroup[2]}, sortedCommentGroup[2:4]),
		Entry("multiple entry comment group take 2", []*ast.CommentGroup{sortedCommentGroup[4], sortedCommentGroup[3], sortedCommentGroup[0], sortedCommentGroup[1], sortedCommentGroup[2]}, sortedCommentGroup),
	)
})
