/**
* This file was automatically generated by @cosmwasm/ts-codegen@0.35.7.
* DO NOT MODIFY IT BY HAND. Instead, modify the source JSONSchema file,
* and run the @cosmwasm/ts-codegen generate command to regenerate this file.
*/

import { selectorFamily } from "recoil";
import { cosmWasmClient } from "./chain";
import { TxEncoding, InstantiateMsg, ChannelOpenInitOptions, ExecuteMsg, CosmosMsgForEmpty, BankMsg, Uint128, Binary, IbcMsg, Timestamp, Uint64, WasmMsg, GovMsg, VoteOption, Decimal, Coin, Empty, IbcTimeout, IbcTimeoutBlock, WeightedVoteOption, QueryMsg, CallbackCounter, IbcOrder, ChannelStatus, ChannelState, IbcChannel, IbcEndpoint, Addr, ContractState, IcaInfo } from "./StorageOutpost.types";
import { StorageOutpostQueryClient } from "./StorageOutpost.client";
type QueryClientParams = {
  contractAddress: string;
};
export const queryClient = selectorFamily<StorageOutpostQueryClient, QueryClientParams>({
  key: "storageOutpostQueryClient",
  get: ({
    contractAddress
  }) => ({
    get
  }) => {
    const client = get(cosmWasmClient);
    return new StorageOutpostQueryClient(client, contractAddress);
  }
});
export const getChannelSelector = selectorFamily<ChannelState, QueryClientParams & {
  params: Parameters<StorageOutpostQueryClient["getChannel"]>;
}>({
  key: "storageOutpostGetChannel",
  get: ({
    params,
    ...queryClientParams
  }) => async ({
    get
  }) => {
    const client = get(queryClient(queryClientParams));
    return await client.getChannel(...params);
  }
});
export const getContractStateSelector = selectorFamily<ContractState, QueryClientParams & {
  params: Parameters<StorageOutpostQueryClient["getContractState"]>;
}>({
  key: "storageOutpostGetContractState",
  get: ({
    params,
    ...queryClientParams
  }) => async ({
    get
  }) => {
    const client = get(queryClient(queryClientParams));
    return await client.getContractState(...params);
  }
});
export const getCallbackCounterSelector = selectorFamily<CallbackCounter, QueryClientParams & {
  params: Parameters<StorageOutpostQueryClient["getCallbackCounter"]>;
}>({
  key: "storageOutpostGetCallbackCounter",
  get: ({
    params,
    ...queryClientParams
  }) => async ({
    get
  }) => {
    const client = get(queryClient(queryClientParams));
    return await client.getCallbackCounter(...params);
  }
});