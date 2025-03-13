package properorder

import (
	"github.com/golang-collections/collections/stack"
	"go/token"
	"strings"

	"go/ast"
	"go/types"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

func New() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "properorder",
		Doc:      "Checks code style part about the structure of the code in the file",
		Run:      run,
		Requires: []*analysis.Analyzer{inspect.Analyzer}, // optimization: golangci-lint will run inspect.Analyzer only once for all analyzers
	}
}

func run(pass *analysis.Pass) (any, error) {
	var lastFile *token.File
	v := validator{
		Pass:  pass,
		Stack: stack.New(),
	}

	nodeTypes := []ast.Node{
		(*ast.FuncDecl)(nil), // func Foo() or func (r *Receiver) Foo()
		(*ast.TypeSpec)(nil), // type smt (e.g. struct, int, array etc)
	}
	enterNode := func(n ast.Node) bool {
		currentFile := pass.Fset.File(n.Pos()) // clearing the stack in each new file
		if lastFile == nil || currentFile != lastFile {
			v.flushStack()
		}
		lastFile = currentFile

		switch node := n.(type) {
		case *ast.TypeSpec:
			v.CheckTypeSpecPosition(node)
			v.Stack.Push(node)
		case *ast.FuncDecl:
			v.TraverseStack(node)
			v.Stack.Push(node)
		}
		return false
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.WithStack(nodeTypes, func(n ast.Node, push bool, stack []ast.Node) (proceed bool) {
		if push {
			return enterNode(n)
		}
		return true
	})

	return nil, nil
}

var constructorPrefixes = []string{"New", "Parse", "new", "parse"}

type validator struct {
	Pass  *analysis.Pass
	Stack *stack.Stack
}

func (v *validator) TraverseStack(funcDecl *ast.FuncDecl) {
	interfaceValue := v.Stack.Peek()
	if interfaceValue == nil {
		return
	}

	switch value := interfaceValue.(type) {
	case *ast.FuncDecl:
		v.handleFuncDecl(value, funcDecl)
	case *ast.TypeSpec:
		v.handleTypeSpec(value, funcDecl)
	}
}

func (v *validator) CheckTypeSpecPosition(typeSpec *ast.TypeSpec) {
	if interfaceValue := v.Stack.Peek(); interfaceValue != nil {
		if funcDecl, ok := interfaceValue.(*ast.FuncDecl); ok {
			if isFunc(funcDecl) && isFuncHasResults(funcDecl) {
				result := funcDecl.Type.Results.List[0]
				if isTypeNamedAsExpected(unstar(v.Pass.TypesInfo.TypeOf(result.Type)), typeSpec.Name.Name) {
					v.reportInvalidOrder(result.Type.Pos(), "the constructor must be positioned after the type is defined.")
					_ = v.Stack.Pop()
				}
			}
		}
	}
}

func (v *validator) handleFuncDecl(prev, next *ast.FuncDecl) {
	switch {
	case isFunc(prev) && !isFuncHasResults(prev):
		v.validateFuncWithoutResult(prev, next)
	case isFunc(next) && isFuncHasResults(next) && isMethod(prev):
		v.validateMethodOrderRelativeToConstructor(prev, next)
	case isMethod(next):
		v.validateReceiverTypeConsistency(prev, next)
	}

	if isFunc(next) && isFuncHasResults(next) && isConstructorName(next.Name.Name) {
		v.flushStack()
	}
}

func (v *validator) handleTypeSpec(prev *ast.TypeSpec, next *ast.FuncDecl) {
	expectedName := prev.Name.Name
	if v.isExpectedTypeNameInFieldList(next.Type.Params, expectedName) {
		return
	}

	if isMethod(next) {
		newReceiver := next.Recv.List[0]
		if !isTypeNamedAsExpected(unstar(v.Pass.TypesInfo.TypeOf(newReceiver.Type)), expectedName) {
			v.reportInvalidOrder(newReceiver.Pos(), "the method has a different type than the declared type above.")
		}
	} else if isFunc(next) && isFuncHasResults(next) {
		result := next.Type.Results.List[0]
		if !isTypeNamedAsExpected(unstar(v.Pass.TypesInfo.TypeOf(result.Type)), expectedName) && isConstructorName(next.Name.Name) {
			v.reportInvalidOrder(result.Pos(), "the return value of the function has a different type than the declared type above.")
		}
	}
}

func (v *validator) validateFuncWithoutResult(prev, next *ast.FuncDecl) {
	var (
		prevPrev *ast.FuncDecl
		ok       bool
	)
	stackLen := v.Stack.Len()
	if stackLen == 0 {
		return
	}
	for i := 0; i < stackLen; i++ {
		_ = v.Stack.Pop()
		if prevPrev, ok = v.Stack.Peek().(*ast.FuncDecl); ok {
			if isFunc(prevPrev) && !isFuncHasResults(prevPrev) {
				continue
			}
			break
		}
		return
	}
	if prevPrev == nil {
		return
	}

	switch {
	case isMethod(prevPrev) && isMethod(next):
		nextReceiver := next.Recv.List[0]
		prevPrevReceiver := prevPrev.Recv.List[0]
		if v.isTypesMatch(nextReceiver.Type, prevPrevReceiver.Type) {
			v.reportInvalidOrder(prev.Pos(), "the function is located inside a block of consecutive methods.")
		}
	case isFunc(prevPrev) && isMethod(next):
		nextReceiver := next.Recv.List[0]
		result := prevPrev.Type.Results.List[0]
		if v.isTypesMatch(nextReceiver.Type, result.Type) && isConstructorName(prevPrev.Name.Name) {
			v.reportInvalidOrder(prev.Pos(), "the function is located inside the constructor and method block.")
		}
	}
}

func (v *validator) validateReceiverTypeConsistency(prev, next *ast.FuncDecl) {
	newReceiver := next.Recv.List[0]
	switch {
	case isMethod(prev):
		prevReceiver := prev.Recv.List[0]
		if !v.isTypesMatch(newReceiver.Type, prevReceiver.Type) {
			v.reportInvalidOrder(newReceiver.Pos(), "the method has a different type than the method above.")
		}
	case prev.Type.Results != nil:
		result := prev.Type.Results.List[0]
		if !v.isTypesMatch(result.Type, newReceiver.Type) {
			v.reportInvalidOrder(newReceiver.Pos(), "the method has a different type than the return value of the function above.")
		}
	}
}

func (v *validator) validateMethodOrderRelativeToConstructor(prev, next *ast.FuncDecl) {
	prevReceiver := prev.Recv.List[0]
	result := next.Type.Results.List[0]

	if v.isTypesMatch(result.Type, prevReceiver.Type) && isConstructorName(next.Name.Name) {
		v.reportInvalidOrder(prevReceiver.Pos(), "the method must be located below the constructor function.")
	}
}

func (v *validator) isExpectedTypeNameInFieldList(fieldList *ast.FieldList, expectedName string) bool {
	if fieldList == nil {
		return false
	}
	for _, field := range fieldList.List {
		if isTypeNamedAsExpected(v.Pass.TypesInfo.TypeOf(field.Type), expectedName) {
			return true
		}
	}
	return false
}

func (v *validator) isTypesMatch(t1, t2 ast.Expr) bool {
	return types.Identical(unstar(v.Pass.TypesInfo.TypeOf(t1)), unstar(v.Pass.TypesInfo.TypeOf(t2)))
}

func (v *validator) reportInvalidOrder(pos token.Pos, message string) {
	v.Pass.Report(analysis.Diagnostic{
		Pos:     pos,
		Message: message,
	})
}

func (v *validator) flushStack() {
	stackCount := v.Stack.Len()
	for i := 0; i <= stackCount; i++ {
		_ = v.Stack.Pop()
	}
}

func isConstructorName(name string) bool {
	for _, prefix := range constructorPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func isTypeNamedAsExpected(unstarredType types.Type, expectedName string) bool {
	if named, ok := unstarredType.(*types.Named); ok {
		return named.Obj().Name() == expectedName
	}
	return false
}

func isMethod(f *ast.FuncDecl) bool {
	return f.Recv != nil
}

func isFunc(f *ast.FuncDecl) bool {
	return f.Recv == nil
}

func isFuncHasResults(f *ast.FuncDecl) bool {
	return f.Type.Results != nil
}

func unstar(t types.Type) types.Type {
	if p, ok := t.(*types.Pointer); ok {
		return p.Elem()
	}
	return t
}
