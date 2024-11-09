package exhaustinterface

import (
	"go/ast"
	"go/types"
	"log"
	"reflect"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "exhaustinterface checks if type-switch statement exhausts all types of target interface. Target interface must have at least one package-internal method."

var SealedInterfaceFinder = &analysis.Analyzer{
	Name: "sealedinterfacefinder",
	Doc:  doc,
	Run:  findSealedInterfaces,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	ResultType: reflect.TypeOf([]*types.Interface{}),
}

func findSealedInterfaces(pass *analysis.Pass) (any, error) {
	log.Printf("START findSealedInterfaces")
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	sealedInterfaces := make([]*types.Interface, 0)

	nodeFilter := []ast.Node{
		(*ast.TypeSpec)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		if ts, ok := n.(*ast.TypeSpec); ok {
			iface := pass.TypesInfo.TypeOf(ts.Type)
			log.Printf("name: %v, type: %v", ts.Name.Name, iface)

			if iface, ok := iface.(*types.Interface); ok {
				for i := 0; i < iface.NumMethods(); i++ {
					m := iface.Method(i)
					if !m.Exported() {
						log.Printf("sealed interface found: %v", iface)
						sealedInterfaces = append(sealedInterfaces, iface)
						break
					}
				}
			}
		}
	})

	return sealedInterfaces, nil
}

var SealedInstanceFinder = &analysis.Analyzer{
	Name: "sealedinstancefinder",
	Doc:  doc,
	Run:  findSealedInstances,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
		SealedInterfaceFinder,
	},
	ResultType: reflect.TypeOf(make(map[*types.Interface][]types.Type)),
}

func findSealedInstances(pass *analysis.Pass) (any, error) {
	log.Printf("START findSealedInstances")
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	sealdedInterfaces := pass.ResultOf[SealedInterfaceFinder].([]*types.Interface)

	sealedInstances := make(map[*types.Interface][]types.Type)

	nodeFilter := []ast.Node{
		(*ast.TypeSpec)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		if ts, ok := n.(*ast.TypeSpec); ok {
			// check if ts.Type implements sealed interface
			typ := pass.TypesInfo.Defs[ts.Name].Type()
			log.Printf("name: %v, type: %v", ts.Name.Name, typ)
			for _, iface := range sealdedInterfaces {
				if types.Identical(typ.Underlying(), iface) {
					log.Printf("skip sealed interface itself: %v", iface)
				} else if types.Implements(typ, iface) {
					log.Printf("implements %v: %v", iface, typ)
					sealedInstances[iface] = append(sealedInstances[iface], typ)
				}
			}
		}
	})

	return sealedInstances, nil
}

var Analyzer = &analysis.Analyzer{
	Name: "exhaustinterface",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
		SealedInstanceFinder,
	},
}

func run(pass *analysis.Pass) (any, error) {
	log.Printf("START run")
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.TypeSwitchStmt)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.TypeSwitchStmt:
			switch assign := n.Assign.(type) {
			case *ast.AssignStmt:
				if len(assign.Rhs) != 1 {
					log.Printf("invalid type switch: %v", n)
					return
				}

				if _, ok := assign.Rhs[0].(*ast.TypeAssertExpr); !ok {
					log.Printf("invalid type switch: %v", n)
					return
				}

				log.Printf("type switch: %#v", assign.Rhs[0])

				typ := pass.TypesInfo.TypeOf(assign.Rhs[0].(*ast.TypeAssertExpr).X)
				log.Printf("type switch: %v", typ)

				for iface, instances := range pass.ResultOf[SealedInstanceFinder].(map[*types.Interface][]types.Type) {
					if types.Identical(typ.Underlying(), iface) {
						log.Printf("found sealed interface: %v", iface)

						// check if all instances are handled
						handled := make(map[types.Type]bool)
						for _, instance := range instances {
							handled[instance] = false
						}

						for _, stmt := range n.Body.List {
							if c, ok := stmt.(*ast.CaseClause); ok {
								if len(c.List) == 0 {
									// default case
									continue
								}

								for _, expr := range c.List {
									exprType := pass.TypesInfo.TypeOf(expr)
									for _, instance := range instances {
										if types.Identical(exprType, instance) {
											log.Printf("handled: %v", exprType)
											handled[instance] = true
										}
									}
								}
							}
						}

						for instance, handled := range handled {
							if !handled {
								pass.Reportf(n.Pos(), "instance %v for %v is not handled", instance, iface)
							}
						}
					}
				}
			case *ast.ExprStmt:
				log.Printf("expr: %v", assign)
			}
		}
	})

	return nil, nil
}
