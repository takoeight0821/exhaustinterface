package ast

type Expr interface {
	sealedExpr()
}

type Var struct {
	Name string
}

func (Var) sealedExpr() {}

var _ Expr = Var{}

type Int struct {
	Value int
}

func (Int) sealedExpr() {}

type Add struct {
	X, Y Expr
}

func (Add) sealedExpr() {}

type Sub struct {
	X, Y Expr
}

func (Sub) sealedExpr() {}

type Operator interface {
	Add | Sub
}
