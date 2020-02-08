package astutil

import (
	"fmt"

	"github.com/lighttiger2505/sqls/ast"
	"github.com/lighttiger2505/sqls/dialect"
	"github.com/lighttiger2505/sqls/token"
)

type NodeMatcher struct {
	NodeTypeMatcherFunc func(node interface{}) bool
	ExpectTokens        []token.Kind
	ExpectSQLType       []dialect.KeywordKind
	ExpectKeyword       []string
}

func (nm *NodeMatcher) IsMatchNodeType(node interface{}) bool {
	if nm.NodeTypeMatcherFunc != nil {
		if nm.NodeTypeMatcherFunc(node) {
			return true
		}
	}
	return false
}

func (nm *NodeMatcher) IsMatchTokens(tok *ast.SQLToken) bool {
	if nm.ExpectTokens != nil {
		for _, expect := range nm.ExpectTokens {
			if tok.MatchKind(expect) {
				return true
			}
		}
	}
	return false
}

func (nm *NodeMatcher) IsMatchSQLType(tok *ast.SQLToken) bool {
	if nm.ExpectSQLType != nil {
		for _, expect := range nm.ExpectSQLType {
			if tok.MatchSQLKind(expect) {
				return true
			}
		}
	}
	return false
}

func (nm *NodeMatcher) IsMatchKeyword(tok *ast.SQLToken) bool {
	if nm.ExpectKeyword != nil {
		for _, expect := range nm.ExpectKeyword {
			if tok.MatchSQLKeyword(expect) {
				return true
			}
		}
	}
	return false
}

func (nm *NodeMatcher) IsMatch(node ast.Node) bool {
	// For node object
	if nm.IsMatchNodeType(node) {
		return true
	}
	if _, ok := node.(ast.TokenList); ok {
		return false
	}
	// For token object
	tok, ok := node.(ast.Token)
	if !ok {
		panic(fmt.Sprintf("invalid type. not has Token, got=(type: %T, value: %#v)", node, node.String()))
	}
	sqlTok := tok.GetToken()
	if nm.IsMatchTokens(sqlTok) || nm.IsMatchSQLType(sqlTok) || nm.IsMatchKeyword(sqlTok) {
		return true
	}
	return false
}

func isWhitespace(node ast.Node) bool {
	tok, ok := node.(ast.Token)
	if !ok {
		return false
	}
	if tok.GetToken().MatchKind(token.Whitespace) {
		return true
	}
	return false
}

type NodeReader struct {
	Node    ast.TokenList
	CurNode ast.Node
	Index   uint
}

func NewNodeReader(list ast.TokenList) *NodeReader {
	return &NodeReader{
		Node: list,
	}
}

func (nr *NodeReader) CopyReader() *NodeReader {
	return &NodeReader{
		Node:  nr.Node,
		Index: nr.Index,
	}
}

func (nr *NodeReader) NodesWithRange(startIndex, endIndex uint) []ast.Node {
	return nr.Node.GetTokens()[startIndex:endIndex]
}

func (nr *NodeReader) hasNext() bool {
	return nr.Index < uint(len(nr.Node.GetTokens()))
}

func (nr *NodeReader) NextNode(ignoreWhiteSpace bool) bool {
	if !nr.hasNext() {
		return false
	}
	nr.CurNode = nr.Node.GetTokens()[nr.Index]
	nr.Index++

	if ignoreWhiteSpace && isWhitespace(nr.CurNode) {
		return nr.NextNode(ignoreWhiteSpace)
	}
	return true
}

func (nr *NodeReader) CurNodeIs(nm NodeMatcher) bool {
	if nr.CurNode != nil {
		if nm.IsMatch(nr.CurNode) {
			return true
		}
	}
	return false
}

func (nr *NodeReader) PeekNode(ignoreWhiteSpace bool) (uint, ast.Node) {
	tmpReader := nr.CopyReader()
	for tmpReader.hasNext() {
		index := tmpReader.Index
		node := tmpReader.Node.GetTokens()[index]

		if ignoreWhiteSpace {
			if !isWhitespace(node) {
				return index, node
			}
		} else {
			return index, node
		}
		tmpReader.NextNode(false)
	}
	return 0, nil
}

func (nr *NodeReader) PeekNodeIs(ignoreWhiteSpace bool, nm NodeMatcher) bool {
	_, node := nr.PeekNode(ignoreWhiteSpace)
	if node != nil {
		if nm.IsMatch(node) {
			return true
		}
	}
	return false
}

func (nr *NodeReader) FindNode(ignoreWhiteSpace bool, nm NodeMatcher) (*NodeReader, ast.Node) {
	tmpReader := nr.CopyReader()
	for tmpReader.hasNext() {
		node := tmpReader.Node.GetTokens()[tmpReader.Index]

		// For node object
		if nm.IsMatchNodeType(node) {
			return tmpReader, node
		}
		if _, ok := tmpReader.CurNode.(ast.TokenList); ok {
			continue
		}
		// For token object
		tok, _ := nr.CurNode.(ast.Token)
		sqlTok := tok.GetToken()
		if nm.IsMatchTokens(sqlTok) || nm.IsMatchSQLType(sqlTok) || nm.IsMatchKeyword(sqlTok) {
			return tmpReader, node
		}
		tmpReader.NextNode(ignoreWhiteSpace)
	}
	return nil, nil
}
