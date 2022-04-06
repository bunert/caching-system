import time
import pandas as pd
import logging

redis_list = []
lambda_list = []

redis_idx = 0
lambda_idx = 0

lambda_redis_tradeoff = 9 # if 5 request within next 90 seconds, lambda and redis cheaper than lambda for each request
lambda_runtime = 5.0


state = 'none'

def removeIsolatedOnes(df, number):

    return df

def checkRequests(df, idx, numberOfRequests):
    nextIdx = idx+numberOfRequests
    timestamp = df.loc[idx]['delay']
    max_timestamp = timestamp+90.0

    # if curr index not in dataframe, test one smaller 
    if (nextIdx > len(df)-1):
        return checkRequests(df, idx, numberOfRequests-1)
    else:
        nextTimestamp = df.loc[nextIdx]['delay']
        # return True if x-th requests withing 90 seconds
        return nextTimestamp < max_timestamp

def checkNextRequest(df, idx, duration):
    nextIdx = idx+1
    timestamp = df.loc[idx]['delay']
    if (nextIdx > len(df)-1):
        return False
    else:
        nextTimestamp = df.loc[nextIdx]['delay']
        return nextTimestamp > (timestamp+duration)

def checkLambdaTimestamp(df, index, timestamp, lambda_end):

    return timestamp > lambda_end

def checkIfStart(df, index, timestamp):
    global lambda_idx, state, lambda_list
    if checkRequests(df, index, lambda_redis_tradeoff):
        # start lambda & ec2-redis cheaper
        # logging.info("start lambda & ec2-redis cheaper")
        lambda_list.append([timestamp, timestamp+30.0])
        lambda_idx += 1
        redis_list.append([timestamp+30.0, timestamp+90.0])
        state = 'redis'
    else:
        # logging.info("start lambda")
        lambda_list.append([timestamp, timestamp+lambda_runtime])
        state = 'lambda'

def optimizeSystem(df, start, end):
    global lambda_idx, redis_idx, state, lambda_list, redis_list
    lambda_list = []
    redis_list = []
    redis_idx = 0
    lambda_idx = 0
    state = 'none'

    # s3_requests = len(df[df['origin'] == "S3"])
    # logging.info(f"total requests: {len(df)}")
    # logging.info(f"s3_requests: {s3_requests}")
    # df = removeIsolatedOnes(df, s3_requests)
    # logging.info(f"remaining requests: {len(df)}")
    # print(df)

    for index, row in df.iterrows():
        # logging.info(f"state: {state} \t ({index})")
        if state == 'none':
            checkIfStart(df, index, row['delay'])
            continue
        elif state == 'lambda':
            # logging.info(f"{row['delay']} > {lambda_list[lambda_idx][1]}")
            if row['delay'] > lambda_list[lambda_idx][1]:
                # lambda not running
                # logging.info("lambda not running")
                lambda_idx+=1
                checkIfStart(df, index, row['delay'])
            else:
                # lambda still running
                # logging.info("lambda still running")
                if checkRequests(df, index, lambda_redis_tradeoff):
                    # logging.info("start lambda & ec2-redis cheaper")
                    end = lambda_list[lambda_idx][1]
                    lambda_list[lambda_idx][1] = end+30.0
                    lambda_idx += 1
                    redis_list.append([end+30.0, end+90.0])
                    state = 'redis'
                else:
                    # logging.info("extend lambda")
                    lambda_list[lambda_idx][1] += lambda_runtime
                    state = 'lambda'
            continue
        elif state == 'redis':
            if row['delay'] < redis_list[redis_idx][1]:
                # redis running
                # redis_list[redis_idx][1] = row['delay']
                continue
            else:
                # redis not running
                # logging.info("redis not running")
                if checkNextRequest(df, index, 104.6):
                    # logging.info("redis shutdown at last request")
                    redis_idx+=1
                    checkIfStart(df, index, row['delay'])
                else:
                    # logging.info("keep redis running")
                    redis_list[redis_idx][1] = row['delay']

            continue
    

    ec2redis_best_df = pd.DataFrame(redis_list, columns=['start', 'end'])
    lambda_best_df = pd.DataFrame(lambda_list, columns=['start', 'end'])
    # print("ec dataframe:")
    # print(ec2redis_best_df)
    # print("lambda df:")
    # print(lambda_best_df)

    return ec2redis_best_df, lambda_best_df