package eval

import "calc/ast"

type Evaluator struct {
	env *Env
}

type Value = int

type Env struct {
	parent *Env
	table  map[string]int
}

func (e *Env) get(name string) int {
	if v, ok := e.table[name]; ok {
		return v
	}
	if e.parent == nil {
		panic("undefined variable: " + name)
	}
	return e.parent.get(name)
}

func newEnv(parent *Env) *Env {
	return &Env{parent: parent, table: make(map[string]int)}
}

func NewEvaluator() *Evaluator {
	return &Evaluator{env: newEnv(nil)}
}

func (e *Evaluator) Eval(expr ast.Expr) int {
	switch expr := expr.(type) {
	case ast.Var:
		return e.env.get(expr.Name)
	case ast.Int:
		return expr.Value
	case ast.Add:
		return runPrim(e, expr)
	// case ast.Sub:
	// 	return runPrim(e, expr)
	default:
		panic("invalid")
	}
}

func runPrim[O ast.Operator](e *Evaluator, expr O) int {
	switch expr := ast.Expr(expr).(type) {
	case ast.Add:
		return e.Eval(expr.X) + e.Eval(expr.Y)
	case ast.Sub:
		return e.Eval(expr.X) - e.Eval(expr.Y)
	default:
		panic("invalid")
	}
}
