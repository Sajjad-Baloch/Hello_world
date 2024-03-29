/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package historyleveldb

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/hyperledger/fabric/common/ledger/blkstorage"
	"github.com/hyperledger/fabric/common/ledger/blkstorage/fsblkstorage"
	"github.com/hyperledger/fabric/common/metrics/disabled"
	"github.com/hyperledger/fabric/core/ledger/kvledger/bookkeeping"
	"github.com/hyperledger/fabric/core/ledger/kvledger/history/historydb"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/privacyenabledstate"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/txmgr"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/txmgr/lockbasedtxmgr"
	"github.com/hyperledger/fabric/core/ledger/mock"
	"github.com/stretchr/testify/assert"
)

/////// levelDBLockBasedHistoryEnv //////

type levelDBLockBasedHistoryEnv struct {
	t                     testing.TB
	testBlockStorageEnv   *testBlockStoreEnv
	testDBEnv             privacyenabledstate.TestEnv
	testBookkeepingEnv    *bookkeeping.TestEnv
	txmgr                 txmgr.TxMgr
	testHistoryDBProvider historydb.HistoryDBProvider
	testHistoryDB         historydb.HistoryDB
	testHistoryDBPath     string
}

func newTestHistoryEnv(t *testing.T) *levelDBLockBasedHistoryEnv {
	testLedgerID := "TestLedger"

	blockStorageTestEnv := newBlockStorageTestEnv(t)

	testDBEnv := &privacyenabledstate.LevelDBCommonStorageTestEnv{}
	testDBEnv.Init(t)
	testDB := testDBEnv.GetDBHandle(testLedgerID)
	testBookkeepingEnv := bookkeeping.NewTestEnv(t)

	testHistoryDBPath, err := ioutil.TempDir("", "historyldb")
	if err != nil {
		t.Fatalf("Failed to create history database directory: %s", err)
	}

	txMgr, err := lockbasedtxmgr.NewLockBasedTxMgr(testLedgerID, testDB, nil, nil, testBookkeepingEnv.TestProvider, &mock.DeployedChaincodeInfoProvider{})
	assert.NoError(t, err)
	testHistoryDBProvider := NewHistoryDBProvider(testHistoryDBPath)
	testHistoryDB, err := testHistoryDBProvider.GetDBHandle("TestHistoryDB")
	assert.NoError(t, err)

	return &levelDBLockBasedHistoryEnv{
		t,
		blockStorageTestEnv,
		testDBEnv,
		testBookkeepingEnv,
		txMgr,
		testHistoryDBProvider,
		testHistoryDB,
		testHistoryDBPath,
	}
}

func (env *levelDBLockBasedHistoryEnv) cleanup() {
	env.txmgr.Shutdown()
	env.testDBEnv.Cleanup()
	env.testBlockStorageEnv.cleanup()
	env.testBookkeepingEnv.Cleanup()
	// clean up history
	env.testHistoryDBProvider.Close()
	os.RemoveAll(env.testHistoryDBPath)
}

/////// testBlockStoreEnv//////

type testBlockStoreEnv struct {
	t               testing.TB
	provider        *fsblkstorage.FsBlockstoreProvider
	blockStorageDir string
}

func newBlockStorageTestEnv(t testing.TB) *testBlockStoreEnv {

	testPath, err := ioutil.TempDir("", "historyleveldb-")
	if err != nil {
		panic(err)
	}
	conf := fsblkstorage.NewConf(testPath, 0)

	attrsToIndex := []blkstorage.IndexableAttr{
		blkstorage.IndexableAttrBlockHash,
		blkstorage.IndexableAttrBlockNum,
		blkstorage.IndexableAttrTxID,
		blkstorage.IndexableAttrBlockNumTranNum,
	}
	indexConfig := &blkstorage.IndexConfig{AttrsToIndex: attrsToIndex}

	blockStorageProvider := fsblkstorage.NewProvider(conf, indexConfig, &disabled.Provider{}).(*fsblkstorage.FsBlockstoreProvider)

	return &testBlockStoreEnv{t, blockStorageProvider, testPath}
}

func (env *testBlockStoreEnv) cleanup() {
	env.provider.Close()
	env.removeFSPath()
}

func (env *testBlockStoreEnv) removeFSPath() {
	fsPath := env.blockStorageDir
	os.RemoveAll(fsPath)
}
