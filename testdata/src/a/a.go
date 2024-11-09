package a

type Expr interface {
	sealedExpr()
}

type Var struct {
	Name string
}

func (Var) sealedExpr() {}

var _ Expr = Var{}

type Number struct {
	Val int
}

func (Number) sealedExpr() {}

type Add struct {
	X, Y Expr
}

func (Add) sealedExpr() {}

func Eval(expr Expr) int {
	switch expr := expr.(type) {
	case Number:
		return expr.Val
	case Add:
		return Eval(expr.X) + Eval(expr.Y)
	default:
		panic("invalid")
	}
}
