// Copyright 2015 The go-ethereum Authors
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

package abi

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

// FunctionType represents different types of functions a contract might have.
// FunctionType 表示合约可能拥有的不同类型的函数。
type FunctionType int

const (
	// Constructor represents the constructor of the contract.
	// The constructor function is called while deploying a contract.
	// Constructor 表示合约的构造函数。构造函数在部署合约时被调用。
	Constructor FunctionType = iota
	// Fallback represents the fallback function.
	// This function is executed if no other function matches the given function
	// signature and no receive function is specified.
	// Fallback 表示回退函数。当没有其他函数匹配给定的函数签名且未指定接收函数时，执行此函数。
	Fallback
	// Receive represents the receive function.
	// This function is executed on plain Ether transfers.
	// Receive 表示接收函数。在纯以太币转账时执行此函数。
	Receive
	// Function represents a normal function.
	// Function 表示普通函数。
	Function
)

// Method represents a callable given a `Name` and whether the method is a constant.
// 如果方法是 `Const`，则无需为此特定方法创建事务。它可以通过本地虚拟机轻松模拟。
// 例如，`Balance()` 方法只需要从存储中检索某些内容，因此不需要发送事务到网络。
// 而像 `Transact` 这样的方法则需要事务，因此会被标记为 `false`。
// Input 指定了该方法所需的输入参数。
type Method struct {
	// Name is the method name used for internal representation. It's derived from
	// the raw name and a suffix will be added in the case of a function overload.
	// 名称是用于内部表示的方法名称。它是从原始名称派生的，在函数重载的情况下会添加后缀。
	Name    string
	RawName string // RawName 是从 ABI 中解析出的原始方法名称。

	// Type indicates whether the method is a special fallback introduced in solidity v0.6.0.
	// 类型指示该方法是否为 Solidity v0.6.0 引入的特殊回退函数。
	Type FunctionType

	// StateMutability indicates the mutability state of method,
	// the default value is nonpayable. It can be empty if the abi
	// is generated by legacy compiler.
	// StateMutability 表示方法的状态可变性，默认值为非支付（nonpayable）。如果 ABI 由旧版编译器生成，则可以为空。
	StateMutability string

	// Legacy indicators generated by compiler before v0.6.0
	// 由 v0.6.0 之前的编译器生成的遗留指标。
	Constant bool
	Payable  bool

	Inputs  Arguments // 输入参数列表。
	Outputs Arguments // 输出参数列表。
	str     string    // 方法的字符串表示形式。
	// Sig returns the methods string signature according to the ABI spec.
	// 根据 ABI 规范返回方法的字符串签名。
	Sig string
	// ID returns the canonical representation of the method's signature used by the
	// abi definition to identify method names and types.
	// 返回方法签名的标准表示形式，ABI 定义使用它来标识方法名称和类型。
	ID []byte
}

// NewMethod creates a new Method.
// A method should always be created using NewMethod.
// It also precomputes the sig representation and the string representation
// of the method.
// NewMethod 创建一个新的 Method 对象。应始终使用 NewMethod 创建方法。
// 它还会预先计算方法的签名表示和字符串表示。
func NewMethod(name string, rawName string, funType FunctionType, mutability string, isConst, isPayable bool, inputs Arguments, outputs Arguments) Method {
	var (
		types       = make([]string, len(inputs))  // 输入参数类型列表。
		inputNames  = make([]string, len(inputs))  // 输入参数名称列表。
		outputNames = make([]string, len(outputs)) // 输出参数名称列表。
	)
	for i, input := range inputs {
		inputNames[i] = fmt.Sprintf("%v %v", input.Type, input.Name) // 格式化输入参数为 "类型 名称"。
		types[i] = input.Type.String()                               // 获取输入参数的类型字符串。
	}
	for i, output := range outputs {
		outputNames[i] = output.Type.String() // 获取输出参数的类型字符串。
		if len(output.Name) > 0 {
			outputNames[i] += fmt.Sprintf(" %v", output.Name) // 如果输出参数有名称，则附加名称。
		}
	}
	// 计算签名和方法 ID。注意只有普通函数才有有意义的签名和 ID。
	var (
		sig string
		id  []byte
	)
	if funType == Function {
		sig = fmt.Sprintf("%v(%v)", rawName, strings.Join(types, ","))
		id = crypto.Keccak256([]byte(sig))[:4]
	}
	identity := fmt.Sprintf("function %v", rawName) // 方法标识符。
	switch funType {
	case Fallback:
		identity = "fallback" // 回退函数。
	case Receive:
		identity = "receive" // 接收函数。
	case Constructor:
		identity = "constructor" // 构造函数。
	}
	var str string
	// 提取 Solidity 方法的有意义的状态可变性。
	// 如果它是空字符串或默认值 "nonpayable"，则不打印它。
	if mutability == "" || mutability == "nonpayable" {
		str = fmt.Sprintf("%v(%v) returns(%v)", identity, strings.Join(inputNames, ", "), strings.Join(outputNames, ", "))
	} else {
		str = fmt.Sprintf("%v(%v) %s returns(%v)", identity, strings.Join(inputNames, ", "), mutability, strings.Join(outputNames, ", "))
	}

	return Method{
		Name:            name,
		RawName:         rawName,
		Type:            funType,
		StateMutability: mutability,
		Constant:        isConst,
		Payable:         isPayable,
		Inputs:          inputs,
		Outputs:         outputs,
		str:             str,
		Sig:             sig,
		ID:              id,
	}
}

func (method Method) String() string {
	return method.str // 返回方法的字符串表示形式。
}

// IsConstant returns the indicator whether the method is read-only.
// IsConstant 返回方法是否为只读的指示器。
func (method Method) IsConstant() bool {
	return method.StateMutability == "view" || method.StateMutability == "pure" || method.Constant
}

// IsPayable returns the indicator whether the method can process
// plain ether transfers.
// IsPayable 返回方法是否可以处理纯以太币转账的指示器。
func (method Method) IsPayable() bool {
	return method.StateMutability == "payable" || method.Payable
}
