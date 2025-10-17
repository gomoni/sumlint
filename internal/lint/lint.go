package lint

import (
	"flag"
	"go/ast"
	"go/token"
	"go/types"
	"log/slog"
	"os"
	"sort"
	"strings"

	"golang.org/x/tools/go/analysis"
)

func init() {
	var verboseFlag bool

	fs1 := flag.NewFlagSet(Sum.Name, flag.ExitOnError)
	fs1.BoolVar(&verboseFlag, "verbose", false, "enable verbose output")
	Sum.Flags = *fs1
	fs2 := flag.NewFlagSet(Oneof.Name, flag.ExitOnError)
	fs2.BoolVar(&verboseFlag, "verbose", false, "enable verbose output")
	Oneof.Flags = *fs2
}

// InterfaceImplementorsFact stores (packagePath.InterfaceName) -> implementor type names (qualified)
type InterfaceImplementorsFact struct {
	Implementors []string // fully qualified (pkgpath.TypeName)
}

func (f *InterfaceImplementorsFact) AFact() {}

// Provide stable, concise string form for analysistest fact matching.
func (f *InterfaceImplementorsFact) String() string {
	return strings.Join(f.Implementors, ",")
}

var Sum = &analysis.Analyzer{
	Name: "sumlint",
	Doc:  "checks exhaustive type switches over Sum* interfaces with single unexported marker method",
	Run:  analyzer{prefix: "Sum"}.run,
	FactTypes: []analysis.Fact{
		new(InterfaceImplementorsFact),
	},
}

var Oneof = &analysis.Analyzer{
	Name: "oneoflint",
	Doc:  "checks exhaustive type switches over oneof fields in proto files",
	Run:  analyzer{prefix: "is"}.run,
	FactTypes: []analysis.Fact{
		new(InterfaceImplementorsFact),
	},
}

type analyzer struct {
	prefix string
}

func (a analyzer) run(pass *analysis.Pass) (any, error) {
	flagVerbose := getFlag(pass.Analyzer.Flags, "verbose", false)
	if flagVerbose {
		sumLog := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
		sumLog = sumLog.With("analyzer", pass.Analyzer.Name)
		slog.SetDefault(sumLog)
	}

	sumIfaces := a.discoverSumInterfaces(pass)
	if len(sumIfaces) > 0 {
		slog.Debug("discovered interfaces", "count", len(sumIfaces))
		exportImplementorFacts(pass, sumIfaces)
	}
	ifaceImpls := loadAllInterfaceFacts(pass)
	if flagVerbose {
		count := 0
		for _, impls := range ifaceImpls {
			count += len(impls)
		}
		if count > 0 {
			slog.Debug("discovered implementations", "count", count)
		}
	}

	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSwitchStmt)
			if !ok {
				return true
			}
			checkTypeSwitch(pass, ts, ifaceImpls)
			return true
		})
	}

	return nil, nil
}

func (a analyzer) discoverSumInterfaces(pass *analysis.Pass) map[*types.TypeName]*types.Interface {
	out := make(map[*types.TypeName]*types.Interface)
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				name := ts.Name.Name
				if !strings.HasPrefix(name, a.prefix) {
					continue
				}
				obj, ok := pass.TypesInfo.Defs[ts.Name]
				if !ok {
					continue
				}
				named, ok := obj.Type().(*types.Named)
				if !ok {
					continue
				}
				iface, ok := named.Underlying().(*types.Interface)
				if !ok {
					continue
				}
				if !isValidSumInterface(name, iface) {
					continue
				}
				out[obj.(*types.TypeName)] = iface
			}
		}
	}
	return out
}

func isValidSumInterface(name string, iface *types.Interface) bool {
	iface = iface.Complete()
	if iface.NumEmbeddeds() != 0 {
		return false
	}
	if iface.NumMethods() != 1 {
		return false
	}
	m := iface.Method(0)
	expected := lowerFirst(name)
	if m.Name() != expected {
		return false
	}
	sig, ok := m.Type().(*types.Signature)
	if !ok {
		return false
	}
	if sig.Params().Len() != 0 || sig.Results().Len() != 0 {
		return false
	}
	return true
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func exportImplementorFacts(pass *analysis.Pass, sumIfaces map[*types.TypeName]*types.Interface) {
	byIface := make(map[*types.TypeName]map[string]struct{})
	for ifaceObj := range sumIfaces {
		byIface[ifaceObj] = make(map[string]struct{})
	}
	scope := pass.Pkg.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		tn, ok := obj.(*types.TypeName)
		if !ok {
			continue
		}
		named, ok := tn.Type().(*types.Named)
		if !ok {
			continue
		}
		if _, isIface := named.Underlying().(*types.Interface); isIface {
			continue
		}
		for ifaceObj, ifaceType := range sumIfaces {
			if types.Implements(named, ifaceType) || types.Implements(types.NewPointer(named), ifaceType) {
				key := qualifiedTypeName(tn)
				byIface[ifaceObj][key] = struct{}{}

			}
		}
	}
	for ifaceObj, implSet := range byIface {
		if len(implSet) == 0 {
			continue
		}
		list := make([]string, 0, len(implSet))
		for k := range implSet {
			list = append(list, k)
			slog.Debug("exporting implementation", "interface", ifaceObj.Name(), "name", k)
		}
		sort.Strings(list) // deterministic ordering for tests
		pass.ExportObjectFact(ifaceObj, &InterfaceImplementorsFact{Implementors: list})
	}
}

func loadAllInterfaceFacts(pass *analysis.Pass) map[*types.TypeName]map[string]struct{} {
	res := make(map[*types.TypeName]map[string]struct{})
	scope := pass.Pkg.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		tn, ok := obj.(*types.TypeName)
		if !ok {
			continue
		}
		_, ok = tn.Type().Underlying().(*types.Interface)
		if !ok {
			continue
		}
		var fact InterfaceImplementorsFact
		if pass.ImportObjectFact(tn, &fact) {
			set := make(map[string]struct{}, len(fact.Implementors))
			for _, t := range fact.Implementors {
				slog.Debug("loadAllInterfaceFacts: scopeNames", "tn", tn, "t", t)
				set[t] = struct{}{}
			}
			res[tn] = set
		}
	}
	for ident, obj := range pass.TypesInfo.Uses {
		_ = ident
		tn, ok := obj.(*types.TypeName)
		if !ok {
			continue
		}
		if _, ok := tn.Type().Underlying().(*types.Interface); !ok {
			continue
		}
		if _, exists := res[tn]; exists {
			continue
		}
		var fact InterfaceImplementorsFact
		if pass.ImportObjectFact(tn, &fact) {
			set := make(map[string]struct{}, len(fact.Implementors))
			for _, t := range fact.Implementors {
				slog.Debug("loadAllInterfaceFacts: ImportObjectFact", "tn", tn, "t", t)
				set[t] = struct{}{}
			}
			res[tn] = set
		}
	}
	return res
}
func checkTypeSwitch(pass *analysis.Pass, ts *ast.TypeSwitchStmt, ifaceImpls map[*types.TypeName]map[string]struct{}) {
	var asserted types.Type
	switch ass := ts.Assign.(type) {
	case *ast.AssignStmt:
		if len(ass.Rhs) != 1 {
			return
		}
		ta, ok := ass.Rhs[0].(*ast.TypeAssertExpr)
		if !ok || ta.Type != nil {
			return
		}
		asserted = pass.TypesInfo.TypeOf(ta.X) // e.g. msg.GetPayload()
	case *ast.ExprStmt:
		ta, ok := ass.X.(*ast.TypeAssertExpr)
		if !ok || ta.Type != nil {
			return
		}
		asserted = pass.TypesInfo.TypeOf(ta.X) // e.g. msg.GetPayload()
	default:
		return
	}

	if asserted == nil {
		return
	}
	if ptr, ok := asserted.(*types.Pointer); ok {
		asserted = ptr.Elem()
	}
	if _, ok := asserted.Underlying().(*types.Interface); !ok {
		return
	}

	var ifaceObj *types.TypeName
	if named, ok := asserted.(*types.Named); ok {
		ifaceObj = named.Obj()
	} else {
		// Try match by underlying interface identity
		for tn := range ifaceImpls {
			if tn.Type().Underlying() == asserted.Underlying() {
				ifaceObj = tn
				break
			}
		}
	}
	if ifaceObj == nil {
		return
	}

	implSet, hasFact := ifaceImpls[ifaceObj]
	// If this package did not discover the interface (e.g. unexported in another package),
	// attempt to import its fact now that we have the *types.TypeName from the type switch expression.
	if !hasFact || len(implSet) == 0 {
		var fact InterfaceImplementorsFact
		if pass.ImportObjectFact(ifaceObj, &fact) && len(fact.Implementors) > 0 {
			newSet := make(map[string]struct{}, len(fact.Implementors))
			for _, implObj := range fact.Implementors {
				newSet[implObj] = struct{}{}
			}
			ifaceImpls[ifaceObj] = newSet
			implSet = newSet
			hasFact = true
		}
	}
	if !hasFact || len(implSet) == 0 {
		return
	}

	covered := make(map[string]struct{})
	var hasDefault bool
	for _, cc := range ts.Body.List {
		clause, _ := cc.(*ast.CaseClause)
		if clause == nil {
			continue
		}
		if clause.List == nil {
			hasDefault = true
			continue
		}
		for _, expr := range clause.List {
			t := pass.TypesInfo.TypeOf(expr)
			if t == nil {
				continue
			}
			if pt, ok := t.(*types.Pointer); ok {
				t = pt.Elem()
			}
			named, ok := t.(*types.Named)
			if !ok {
				continue
			}
			tn := named.Obj()
			key := qualifiedTypeName(tn)
			covered[key] = struct{}{}
		}
	}

	var missing []string
	for impl := range implSet {
		if _, ok := covered[impl]; !ok {
			missing = append(missing, impl)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		pass.Reportf(ts.Switch, "non-exhaustive type switch on %s: missing cases for: %s",
			ifaceObj.Name(), strings.Join(missing, ", "))
	}
	if !hasDefault {
		pass.Reportf(ts.Switch, "missing default case on %s: code cannot handle nil interface", ifaceObj.Name())
	}
}

func qualifiedTypeName(tn *types.TypeName) string {
	if tn.Pkg() == nil {
		return tn.Name()
	}
	return tn.Pkg().Path() + "." + tn.Name()
}

func getFlag[T any](fs flag.FlagSet, key string, def T) T {
	if fl := fs.Lookup(key); fl != nil {
		if g, ok := fl.Value.(flag.Getter); ok {
			if t, ok := g.Get().(T); ok {
				return t
			}
		}
	}
	return def
}
