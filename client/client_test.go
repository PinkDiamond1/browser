package client

import (
	"github.com/browser/types"
	"testing"
)

func TestClient(t *testing.T) {
	chainConfig, err := GetChainConfig()
	if err != nil {
		panic(err)
	}
	t.Log(chainConfig.ChainID)
}

func TestSendRawTransaction(t *testing.T) {
	txRawData := "0xf87e80843b9aca00f876f87482020503808c7a6f6d6269657878787878788c6b696e6734343434353535358401406f40018080f84b80f848f84681eca03889c1f6806915060853e43901176c9dd95034aa7266e33894ae3da04ff249e5a066149cbedd90211a8ddf3f43f4a7393ea30fa3d66f0e19a14a3b49fef3ea8095c180"
	hash, err := SendRawTransaction(txRawData)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(hash)
}

func TestGetCurrentBlockInfo(t *testing.T) {
	data := &types.RpcBlock{}
	err := GetData(methodCurrentBlock, data, false)
	if err != nil {
		t.Failed()
	}
}
