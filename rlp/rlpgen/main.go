// Copyright 2022 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/types"
	"os"

	"golang.org/x/tools/go/packages"
)

const pathOfPackageRLP = "github.com/ethereum/go-ethereum/rlp"

func main() {
	var (
		pkgdir     = flag.String("dir", ".", "input package")                   // 指定输入包的目录，默认值为当前目录
		output     = flag.String("out", "-", "output file (default is stdout)") // 指定输出文件路径，默认输出到 stdout
		genEncoder = flag.Bool("encoder", true, "generate EncodeRLP?")          // 是否生成 RLP 编码方法，默认为 true
		genDecoder = flag.Bool("decoder", false, "generate DecodeRLP?")         // 是否生成 RLP 解码方法，默认为 false
		typename   = flag.String("type", "", "type to generate methods for")    // 指定要生成方法的类型名称，默认为空
	)
	flag.Parse() // 解析命令行参数

	cfg := Config{
		Dir:             *pkgdir,
		Type:            *typename,
		GenerateEncoder: *genEncoder,
		GenerateDecoder: *genDecoder,
	}
	code, err := cfg.process() // 调用 process 方法生成代码
	if err != nil {
		fatal(err)
	}
	if *output == "-" { // 如果输出为 stdout，直接写入标准输出
		os.Stdout.Write(code)
	} else if err := os.WriteFile(*output, code, 0600); err != nil { // 否则写入指定文件，若失败则退出
		fatal(err)
	}
}

func fatal(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
}

// Config 结构体用于配置代码生成或解析的参数
type Config struct {
	Dir  string // input package directory 输入包的目录路径
	Type string // 要处理的类型名称

	GenerateEncoder bool // 是否生成编码器
	GenerateDecoder bool // 是否生成解码器
}

// process generates the Go code.
// process 生成 Go 代码
// 加载指定目录中的包，查找目标结构体类型，并生成编码器/解码器代码，最终返回生成的代码字节切片。
func (cfg *Config) process() (code []byte, err error) {
	// Load packages.
	// 加载包
	pcfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes, // 指定加载模式，需要包名和类型信息
		Dir:  cfg.Dir,                                // 使用配置中的目录路径
	}
	ps, err := packages.Load(pcfg, pathOfPackageRLP, ".") // 加载指定路径的包和当前目录的包
	if err != nil {
		return nil, err
	}
	if len(ps) == 0 {
		return nil, fmt.Errorf("no Go package found in %s", cfg.Dir)
	}
	packages.PrintErrors(ps) // 打印包加载过程中的错误信息

	// Find the packages that were loaded.
	// 查找已加载的包
	var (
		pkg        *types.Package // 用户指定的包
		packageRLP *types.Package // RLP 包
	)
	for _, p := range ps {
		if len(p.Errors) > 0 {
			return nil, fmt.Errorf("package %s has errors", p.PkgPath)
		}
		if p.PkgPath == pathOfPackageRLP {
			packageRLP = p.Types // 找到 RLP 包
		} else {
			pkg = p.Types // 找到用户包
		}
	}
	bctx := newBuildContext(packageRLP) // 创建基于 RLP 包的构建上下文

	// Find the type and generate.
	// 查找类型并生成代码
	typ, err := lookupStructType(pkg.Scope(), cfg.Type) // 在包作用域中查找结构体类型
	if err != nil {
		return nil, fmt.Errorf("can't find %s in %s: %v", cfg.Type, pkg, err)
	}
	code, err = bctx.generate(typ, cfg.GenerateEncoder, cfg.GenerateDecoder) // 生成代码
	if err != nil {
		return nil, err
	}

	// Add build comments.
	// This is done here to avoid processing these lines with gofmt.
	// 添加构建注释
	// 此处添加注释以避免 gofmt 处理这些行
	var header bytes.Buffer
	fmt.Fprint(&header, "// Code generated by rlpgen. DO NOT EDIT.\n\n") // 添加生成代码的头部注释
	return append(header.Bytes(), code...), nil
}

// 在指定作用域中查找指定名称的结构体类型，返回结构体类型对象或错误
// scope: 类型作用域，用于查找类型的上下文
// name: 要查找的类型名称
// 返回值: *types.Named 表示找到的结构体类型对象
//
// 在给定作用域中查找特定名称的类型，并确保该类型是一个结构体类型
func lookupStructType(scope *types.Scope, name string) (*types.Named, error) {
	typ, err := lookupType(scope, name) // 调用 lookupType 在作用域中查找类型
	if err != nil {
		return nil, err
	}
	_, ok := typ.Underlying().(*types.Struct) // 检查底层类型是否为结构体，获取 typ 的底层类型。*types.Named 表示一个命名类型（如 type MyStruct struct{}），但其底层类型可能是结构体、整数等。
	if !ok {
		return nil, errors.New("not a struct type")
	}
	return typ, nil
}

// 在指定作用域中查找指定名称的类型，返回类型对象或错误
// scope: 类型作用域，用于查找类型的上下文。这是一个指向 go/types 包中 Scope 结构体的指针，表示一个类型的作用域。作用域通常用于存储变量、函数、类型等定义。Scope 是一个符号表，用于管理标识符和它们的定义。
// name: 要查找的类型名称，例如某个以太坊相关的类型标识符。
// 返回值: *types.Named 表示找到的类型对象，表示一个命名的类型（例如 type MyType int 中的 MyType）。
func lookupType(scope *types.Scope, name string) (*types.Named, error) {
	obj := scope.Lookup(name) // 通过名称在作用域中查找对象
	if obj == nil {           // 如果 Lookup 返回空，说明作用域中不存在该名称的对象。
		return nil, errors.New("no such identifier")
	}
	typ, ok := obj.(*types.TypeName) // 将对象断言为类型名称
	if !ok {
		return nil, errors.New("not a type")
	}
	return typ.Type().(*types.Named), nil // 返回类型名称对应的命名类型对象
}
