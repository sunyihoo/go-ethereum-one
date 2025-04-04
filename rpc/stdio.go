// Copyright 2018 The go-ethereum Authors
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

package rpc

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"time"
)

// DialStdIO creates a client on stdin/stdout.
// DialStdIO 在标准输入/输出上创建一个客户端。
//
// 在某些测试场景或特定的工具中，可能需要与本地运行的以太坊节点或其他服务进行基于命令行的 RPC 交互，这时可以使用标准输入/输出作为通信通道。
func DialStdIO(ctx context.Context) (*Client, error) {
	return DialIO(ctx, os.Stdin, os.Stdout)
}

// DialIO creates a client which uses the given IO channels
// DialIO 创建一个使用给定 IO 通道的客户端
func DialIO(ctx context.Context, in io.Reader, out io.Writer) (*Client, error) {
	cfg := new(clientConfig)
	return newClient(ctx, cfg, newClientTransportIO(in, out))
}

func newClientTransportIO(in io.Reader, out io.Writer) reconnectFunc {
	return func(context.Context) (ServerCodec, error) {
		return NewCodec(stdioConn{
			in:  in,
			out: out,
		}), nil
	}
}

type stdioConn struct {
	in  io.Reader
	out io.Writer
}

func (io stdioConn) Read(b []byte) (n int, err error) {
	return io.in.Read(b)
}

func (io stdioConn) Write(b []byte) (n int, err error) {
	return io.out.Write(b)
}

func (io stdioConn) Close() error {
	return nil
}

func (io stdioConn) RemoteAddr() string {
	return "/dev/stdin"
}

func (io stdioConn) SetWriteDeadline(t time.Time) error {
	return &net.OpError{Op: "set", Net: "stdio", Source: nil, Addr: nil, Err: errors.New("deadline not supported")}
}
