package transfer

import (
	"context"
	"errors"
)

type emptyFileTransfer struct {
}

var ErrEmptyFileTransfer = errors.New("请设置文件转移方式")

func (e *emptyFileTransfer) Transfer(_ context.Context, _ string, _ string) (string, error) {
	return "", ErrEmptyFileTransfer
}

func (e *emptyFileTransfer) DeleteTransferFiles(_ context.Context, _ ...string) error {
	return ErrEmptyFileTransfer
}

func (e *emptyFileTransfer) GetTransferFile(_ context.Context, _ string) (string, error) {
	return "", ErrEmptyFileTransfer
}
