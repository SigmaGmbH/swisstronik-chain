package stream

import (
	"context"
	"fmt"
	"sync"

	"swisstronik/rpc/types"
	evmtypes "swisstronik/x/evm/types"

	"cosmossdk.io/log"
	cmtquery "github.com/cometbft/cometbft/libs/pubsub/query"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

const (
	streamSubscriberName = "swisstronik-json-rpc"
	subscribBufferSize   = 1024

	headerStreamSegmentSize = 128
	headerStreamCapacity    = 128 * 32
	txStreamSegmentSize     = 1024
	txStreamCapacity        = 1024 * 32
	logStreamSegmentSize    = 2048
	logStreamCapacity       = 2048 * 32
)

var (
	txEvents  = tmtypes.QueryForEvent(tmtypes.EventTx).String()
	evmEvents = cmtquery.MustCompile(fmt.Sprintf("%s='%s' AND %s.%s='%s'",
		tmtypes.EventTypeKey,
		tmtypes.EventTx,
		sdk.EventTypeMessage,
		sdk.AttributeKeyModule, evmtypes.ModuleName)).String()
	blockEvents  = tmtypes.QueryForEvent(tmtypes.EventNewBlock).String()
	evmTxHashKey = fmt.Sprintf("%s.%s", evmtypes.TypeMsgEthereumTx, evmtypes.AttributeKeyEthereumTxHash)
)

type RPCHeader struct {
	EthHeader *ethtypes.Header
	Hash      common.Hash
}

// RPCStream provides data streams for newHeads, logs, and pendingTransactions.
type RPCStream struct {
	evtClient rpcclient.EventsClient
	logger    log.Logger
	txDecoder sdk.TxDecoder

	headerStream *Stream[RPCHeader]
	txStream     *Stream[common.Hash]
	logStream    *Stream[*ethtypes.Log]

	wg sync.WaitGroup
}

func NewRPCStreams(evtClient rpcclient.EventsClient, logger log.Logger, txDecoder sdk.TxDecoder) (*RPCStream, error) {
	s := &RPCStream{
		evtClient: evtClient,
		logger:    logger,
		txDecoder: txDecoder,

		headerStream: NewStream[RPCHeader](headerStreamSegmentSize, headerStreamCapacity),
		txStream:     NewStream[common.Hash](txStreamSegmentSize, txStreamCapacity),
		logStream:    NewStream[*ethtypes.Log](logStreamSegmentSize, logStreamCapacity),
	}

	ctx := context.Background()

	chBlocks, err := s.evtClient.Subscribe(ctx, streamSubscriberName, blockEvents, subscribBufferSize)
	if err != nil {
		return nil, err
	}

	chTx, err := s.evtClient.Subscribe(ctx, streamSubscriberName, txEvents, subscribBufferSize)
	if err != nil {
		if err := s.evtClient.UnsubscribeAll(ctx, streamSubscriberName); err != nil {
			s.logger.Error("failed to unsubscribe", "err", err)
		}
		return nil, err
	}

	chLogs, err := s.evtClient.Subscribe(ctx, streamSubscriberName, evmEvents, subscribBufferSize)
	if err != nil {
		if err := s.evtClient.UnsubscribeAll(context.Background(), streamSubscriberName); err != nil {
			s.logger.Error("failed to unsubscribe", "err", err)
		}
		return nil, err
	}

	go s.start(&s.wg, chBlocks, chTx, chLogs)

	return s, nil
}

func (s *RPCStream) Close() error {
	if err := s.evtClient.UnsubscribeAll(context.Background(), streamSubscriberName); err != nil {
		return err
	}
	s.wg.Wait()
	return nil
}

func (s *RPCStream) HeaderStream() *Stream[RPCHeader] {
	return s.headerStream
}

func (s *RPCStream) TxStream() *Stream[common.Hash] {
	return s.txStream
}

func (s *RPCStream) LogStream() *Stream[*ethtypes.Log] {
	return s.logStream
}

func (s *RPCStream) start(
	wg *sync.WaitGroup,
	chBlocks <-chan coretypes.ResultEvent,
	chTx <-chan coretypes.ResultEvent,
	chLogs <-chan coretypes.ResultEvent,
) {
	wg.Add(1)
	defer func() {
		wg.Done()
		if err := s.evtClient.UnsubscribeAll(context.Background(), streamSubscriberName); err != nil {
			s.logger.Error("failed to unsubscribe", "err", err)
		}
	}()

	for {
		select {
		case ev, ok := <-chBlocks:
			if !ok {
				chBlocks = nil
				break
			}

			data, ok := ev.Data.(tmtypes.EventDataNewBlock)
			if !ok {
				s.logger.Error("event data type mismatch", "type", fmt.Sprintf("%T", ev.Data))
				continue
			}

			baseFee := types.BaseFeeFromEvents(data.ResultFinalizeBlock.Events)

			// TODO: fetch bloom from events
			header := types.EthHeaderFromTendermint(data.Block.Header, ethtypes.Bloom{}, baseFee)
			s.headerStream.Add(RPCHeader{EthHeader: header, Hash: common.BytesToHash(data.Block.Header.Hash())})
		case ev, ok := <-chTx:
			if !ok {
				chTx = nil
				break
			}

			data, ok := ev.Data.(tmtypes.EventDataTx)
			if !ok {
				s.logger.Error("event data type mismatch", "type", fmt.Sprintf("%T", ev.Data))
				continue
			}

			tx, err := s.txDecoder(data.Tx)
			if err != nil {
				s.logger.Error("fail to decode tx", "error", err.Error())
				continue
			}

			var hashes []common.Hash
			for _, msg := range tx.GetMsgs() {
				if ethTx, ok := msg.(*evmtypes.MsgHandleTx); ok {
					hashes = append(hashes, ethTx.AsTransaction().Hash())
				}
			}
			s.txStream.Add(hashes...)
		case ev, ok := <-chLogs:
			if !ok {
				chLogs = nil
				break
			}

			if _, ok := ev.Events[evmTxHashKey]; !ok {
				// ignore transaction as it's not from the evm module
				continue
			}

			// get transaction result data
			dataTx, ok := ev.Data.(tmtypes.EventDataTx)
			if !ok {
				s.logger.Error("event data type mismatch", "type", fmt.Sprintf("%T", ev.Data))
				continue
			}
			txLogs, err := evmtypes.DecodeTransactionLogs(dataTx.TxResult.Result.Data)
			if err != nil {
				s.logger.Error("fail to decode evm tx response", "error", err.Error())
				continue
			}

			// Convert swistronik type tranasction log into ethereum tx log
			ethTxLogs := evmtypes.ConvertLogToEthereumType(txLogs)
			s.logStream.Add(ethTxLogs...)
		}

		if chBlocks == nil && chTx == nil && chLogs == nil {
			break
		}
	}
}
