package main

import (
    "fmt"
    "log"

    "go/types"
    "golang.org/x/tools/go/analysis/singlechecker"
    "golang.org/x/tools/go/analysis"
    "golang.org/x/tools/go/packages"
    "golang.org/x/tools/go/ssa"
)

var Analyzer = &analysis.Analyzer{
    Name: "endpoints",
    Doc: "reports API endpoints and the parameters they utilize",
    Run: run,
}

func main() {
    singlechecker.Main(Analyzer)
}

func run(pass *analysis.Pass) (any, error) {
    fmt.Printf("===== beginning discovery: %s =====\n", pass.String())
    cfg := &packages.Config{
        Mode: packages.NeedName |
            packages.NeedSyntax |
            packages.NeedTypes |
            packages.NeedTypesInfo |
            packages.NeedDeps |
            packages.NeedImports,
        Fset: pass.Fset,
        Tests: false,
    }

    pkgs, err := packages.Load(cfg, pass.Pkg.Path())
    if err != nil {
        log.Fatal(err)
    }
    if packages.PrintErrors(pkgs) > 0 {
        log.Fatal("package contains errors")
    }

    prog := ssa.NewProgram(pass.Fset, ssa.SanityCheckFunctions)

    // The single-static assignment package to analyze in this pass
    var ssaPkg *ssa.Package

    // Load referenced packages transitively
    visited := map[string]bool{}
    var visit func(pkg *packages.Package)
    visit = func(pkg *packages.Package) {
        if pkg == nil || visited[pkg.PkgPath] {
            return
        }
        visited[pkg.PkgPath] = true
        for _, imp := range pkg.Imports {
            visit(imp)
        }
        if pkg.Types != nil && pkg.Syntax != nil && pkg.TypesInfo != nil {
            created := prog.CreatePackage(
                pkg.Types, pkg.Syntax, pkg.TypesInfo, true)
            if pkg.PkgPath == pass.Pkg.Path() {
                ssaPkg = created
            }
        }
    }
    for _, pkg := range pkgs {
        visit(pkg)
    }
    prog.Build()

    for _, mem := range ssaPkg.Members {
        fn, ok := mem.(*ssa.Function)
        if !ok {
            continue
        }

        for _, block := range fn.Blocks {
            for _, instr := range block.Instrs {
                call, ok := instr.(ssa.CallInstruction)
                if !ok {
                    continue
                }
                common := call.Common()
                if common == nil {
                    continue
                }
                callee := common.StaticCallee()
                if callee == nil {
                    continue
                }

                if isGorillaMuxRouterHandleFunc(callee) {
                    target, err := getGorillaMuxEndpointFromHandleFunc(common)
                    if err != nil {
                        pass.Reportf(call.Pos(),
                            "unsupported gorilla/mux callsite: %s",
                            err.Error())
                    } else {
                        pass.Reportf(call.Pos(),
                            "gorilla/mux endpoint detected: %s", target)
                    }
                } else if isDecoderDecodeFunction(callee) {
                    fields, err := getDecoderDecodeTargetFields(common)
                    if err != nil {
                        pass.Reportf(call.Pos(),
                            "unsupported json.Decoder.Decode callsite: %s",
                            err.Error())
                    } else {
                        pass.Reportf(call.Pos(),
                            "inferred parameters for json.Decoder.Decode: %v",
                            fields)
                    }
                }
            }
        }
    }
    
    return nil, nil
}

func isDecoderDecodeFunction(fn *ssa.Function) bool {
    if fn == nil {
        return false
    }
    sig := fn.Signature

    recv := sig.Recv()
    if recv == nil {
        return false
    }
    recvType := recv.Type()

    // Accept pointer or value receiver of Decoder
    var named *types.Named
    switch t := recvType.(type) {
    case *types.Pointer:
        var ok bool
        named, ok = t.Elem().(*types.Named)
        if !ok {
            return false
        }
    case *types.Named:
        named = t
    default:
        return false
    }

    obj := named.Obj()
    return obj.Pkg() != nil &&
    obj.Pkg().Path() == "encoding/json" &&
    obj.Name() == "Decoder" &&
    fn.Name() == "Decode"
}

func getDecoderDecodeTargetFields(common *ssa.CallCommon) ([]string, error) {
    args := common.Args
    if len(args) != 2 {
        err := fmt.Errorf(
            "unexpected argument count for json parameter decoding: %d",
            len(args))
        return nil, err 
    }

    arg1 := common.Args[1]

    var targetVal ssa.Value = nil
    var targetType types.Type = nil
    switch v := arg1.(type) {
    case *ssa.MakeInterface:
        targetVal = v.X
        tt := targetVal.Type()
        if ptr, ok := tt.(*types.Pointer); ok {
            tt = ptr.Elem()
        }
        namedType, ok := tt.(*types.Named)
        if !ok {
            return nil, fmt.Errorf("not a named type: %s", tt.String())
        }
        targetType = namedType.Underlying()
    default:
        err := fmt.Errorf(
            "unsupported argument type for json.Decoder.Decode: %T", v)
        return nil, err
    }

    fieldTags := map[string]string{}
    switch t := targetType.(type) {
    case *types.Struct:
        for i := 0; i < t.NumFields(); i++ {
            field := t.Field(i)
            tag := t.Tag(i)
            fieldTags[field.Name()] = tag
        }
    default:
        return nil, fmt.Errorf("unsupported Decoder.Decode target type: %s",
            t.String())
    }

    targets := []string{}
    for fieldName, tagName := range fieldTags {
        if tagName == "" {
            targets = append(targets, fieldName)
        } else {
            targets = append(targets, tagName)
        }
    }

    return targets, nil
}

func isGorillaMuxRouterHandleFunc(fn *ssa.Function) bool {
    if fn == nil {
        return false
    }

    sig := fn.Signature

    recv := sig.Recv()
    if recv == nil {
        return false
    }
    recvType := recv.Type()

    // Accept pointer or value receiver
    var named *types.Named
    switch t := recvType.(type) {
    case *types.Pointer:
        var ok bool
        named, ok = t.Elem().(*types.Named)
        if !ok {
            return false
        }
    case *types.Named:
        named = t
    default:
        return false
    }

    obj := named.Obj()
    return obj.Pkg() != nil &&
        obj.Pkg().Path() == "github.com/gorilla/mux" &&
        obj.Name() == "Router" &&
        fn.Name() == "HandleFunc"
}

func getGorillaMuxEndpointFromHandleFunc(
    call *ssa.CallCommon) (string, error) {

    args := call.Args
    if len(args) < 2 {
        err := fmt.Errorf(
            "unexpected argument count for gorilla/mux endpoint handler: %d",
            len(args))
        return "", err 
    }

    argValue := args[1]

    if argValue.Type().String() != "string" {
        err := fmt.Errorf(
            "endpoint argument for gorilla/mux is not a string: %s",
            argValue.Type().String())
        return "", err
    }

    return argValue.String(), nil
}


