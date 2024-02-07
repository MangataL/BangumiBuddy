package transfer

import "sync"

var (
	fileTransferMu  sync.RWMutex
	fileTransferFac map[string]FileTransfer = make(map[string]FileTransfer)
)

func RegisterFileTransfer(name string, fileTransfer FileTransfer) {
	fileTransferMu.Lock()
	defer fileTransferMu.Unlock()
	fileTransferFac[name] = fileTransfer
}

func GetFileTransfer(name string) FileTransfer {
	fileTransferMu.RLock()
	defer fileTransferMu.RUnlock()
	fileTransfer, ok := fileTransferFac[name]
	if !ok {
		return &emptyFileTransfer{}
	}
	return fileTransfer
}
