// Copyright 2021 The go-ethereum Authors
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

package misc

import (
	"fmt"

	"github.com/ethereum/go-ethereum/params"
)

// VerifyGaslimit verifies the header gas limit according increase/decrease
// in relation to the parent gas limit.
// VerifyGaslimit 验证区块头的 Gas Limit 是否相对于父区块的 Gas Limit 在允许的范围内变化。
func VerifyGaslimit(parentGasLimit, headerGasLimit uint64) error {
	// Verify that the gas limit remains within allowed bounds
	// 验证 Gas Limit 是否保持在允许的范围内。
	diff := int64(parentGasLimit) - int64(headerGasLimit)
	if diff < 0 {
		diff *= -1 // 取绝对值，计算父区块与当前区块 Gas Limit 的差值。
	}
	limit := parentGasLimit / params.GasLimitBoundDivisor // 计算允许的最大变化范围。

	// 检查差值是否超过允许的变化范围。
	if uint64(diff) >= limit {
		return fmt.Errorf("invalid gas limit: have %d, want %d +-= %d", headerGasLimit, parentGasLimit, limit-1)
	}

	// 检查当前区块的 Gas Limit 是否低于最小值。
	if headerGasLimit < params.MinGasLimit {
		return fmt.Errorf("invalid gas limit below %d", params.MinGasLimit)
	}

	// 如果通过所有检查，则返回 nil 表示验证成功。
	return nil
}
