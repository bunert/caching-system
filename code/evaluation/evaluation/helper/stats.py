from evaluation.helper.optimizer import optimizeSystem
import matplotlib.pyplot as plt
import seaborn as sns
import boto3
import logging
import numpy as np
from datetime import datetime, timedelta
import time
import json
from prettytable import PrettyTable
import pytz
import os
import pandas as pd
import math

s3_request_cost = 0.00000043 # S3 per request (S3 standard)
ec_hour_cost = 0.019 # cache.t2.micro instance price per hour
ec2_hour_cost = 0.0134 # t2.micro instance price per hour
lambda_ms_cost = 0.0000000167 # 1025 MB instance per ms
lambda_req_cost = 0.0000002 # price for lambda invocations

tz = pytz.timezone('Europe/Berlin')

def getStats(f, df, start, end):

    runtime = end-start # in seconds
    total_duration = (runtime/60.0) # in minutes
    f.write('{0:<25}  {1:<15}\n'.format("Simulation (minutes):", total_duration))
    number_of_requests = len(df)
    f.write('{0:<25}  {1:<15}\n'.format("Number of requests:", number_of_requests))

    f.write('{0:<25}  {1}\n'.format("Start:", datetime.fromtimestamp(start, tz)))
    f.write('{0:<25}  {1}\n'.format("End:", datetime.fromtimestamp(end, tz)))
    f.write('{0:<40}\n'.format('='*40))

def addS3Costs(f, df):
    logging.info("add S3 costs")
    total_cost = 0.0

    # number of requests * S3 request price
    number_of_requests = df.shape[0]
    cost = number_of_requests * s3_request_cost
    total_cost += cost

    f.write('{0:<25}  {1:<15}\n'.format("S3 request costs:", np.format_float_positional(cost, trim='-')))

    f.write('{0:<40}\n'.format('-'*40))
    f.write('{0:<25}  {1:<15}\n'.format("total cost:", np.format_float_positional(total_cost, trim='-')))
    return total_cost


def addECCosts(f, df, start, end):
    logging.info("add ElastiCache costs")
    total_cost = 0.0

    # add cost for first request, should be S3 access
    if df.loc[1, 'origin'] == "S3":
        total_cost += s3_request_cost
        f.write('{0:<25}  {1:<15}\n'.format("S3 request costs:", np.format_float_positional(s3_request_cost, trim='-')))
    else:
        logging.warning("first request not S3?")
        exit()

    # add cost for ElastiCache runtime (last delay is taken as simulation duration)
    runtime = end-start # in seconds
    total_duration = (runtime/3600.0) # in hours
    ec_cost = total_duration * ec_hour_cost
    total_cost += ec_cost
    f.write('{0:<25}  {1:<15}\n'.format("ElastiCache costs:", np.format_float_positional(ec_cost, trim='-')))
    f.write('{0:<40}\n'.format('-'*40))
    f.write('{0:<25}  {1:<15}\n'.format("total cost:", np.format_float_positional(total_cost, trim='-')))
    return total_cost

def computeLambdaCosts(df):
    df['runtime'] = df['end']-df['start']

    # print(ec_df)
    lambda_invokes = len(df)
    lambda_runtime = df['runtime'].sum()
    
    return lambda_invokes, lambda_runtime

def getLambdaCosts(start, end):
    client = boto3.client('logs')

    query = 'filter @type = "REPORT" | stats sum(@billedDuration) as total_duration, count(*) as number_of_requests, sum(strcontains(@message, "Init Duration")) as cold_starts'  

    log_group = '/aws/lambda/get-s3-object'

    # TODO: added 60 seconds to end time to capture the end of lambda execution
    start_query_response = client.start_query(
        logGroupName=log_group,
        startTime=start,
        endTime=end+60,
        queryString=query,
    )

    query_id = start_query_response['queryId']

    response = None

    while response == None or response['status'] == 'Running':
        logging.info('Waiting for query to complete ...')
        time.sleep(2)
        response = client.get_query_results(
            queryId=query_id
        )
    # logging.info(json.dumps(response))

    if (response["statistics"]["recordsMatched"] >= 1):
        lambda_runtime = int(response["results"][0][0]["value"])
        lambda_requests = int(response["results"][0][1]["value"])
        lambda_coldStarts = int(response["results"][0][2]["value"])
        return lambda_runtime, lambda_requests, lambda_coldStarts
    else:
        logging.warning("CloudWatchLog query did not found any entry?")
        exit()

# @timestamp, @type, @requestId, @billedDuration, @ptr
def process_lambdalog(elem):
    t = datetime.strptime(elem[0]["value"],"%Y-%m-%d %H:%M:%S.%f")
    # convert UTC datetime to specific timezone
    t_conv = t.replace(tzinfo=pytz.utc).astimezone(tz)
    timestamp = datetime.timestamp(t_conv)
    # discard the @ptr fields (elem[3])
    # START type logs
    if len(elem) == 4:
        newElem = [timestamp, elem[1]["value"], elem[2]["value"], np.nan]
    # discard the @ptr fields (elem[4])
    # END type logs
    else:
        # convert ms billedDuration to seconds
        newElem = [timestamp, elem[1]["value"], elem[2]["value"], int(elem[3]["value"])/1000.0]

    return newElem

def getLambdaRuntimes(start, end):
    client = boto3.client('logs')

    query = 'filter @type = "START" or @type = "REPORT" | fields @timestamp, @type, @requestId, @billedDuration| sort @timestamp asc'  

    log_group = '/aws/lambda/get-s3-object'

    # TODO: added 60 seconds to end time to capture the end of lambda execution
    start_query_response = client.start_query(
        logGroupName=log_group,
        startTime=start,
        endTime=end+60,
        queryString=query,
    )

    query_id = start_query_response['queryId']

    response = None

    while response == None or response['status'] == 'Running':
        logging.info('Waiting for query to complete ...')
        time.sleep(2)
        response = client.get_query_results(
            queryId=query_id
        )
    # logging.info(json.dumps(response))
    
    if (response["statistics"]["recordsMatched"] >= 1):
        # lambda_runtime = int(response["results"][0][0]["value"])
        # lambda_requests = int(response["results"][0][1]["value"])
        # lambda_coldStarts = int(response["results"][0][2]["value"])
        # return lambda_runtime, lambda_requests, lambda_coldStarts
        # df = pd.DataFrame.from_dict(response["results"])
        result_list = response["results"]
        processed_list = map(process_lambdalog, result_list)
        df = pd.DataFrame(processed_list, columns = ['timestamp', 'type', 'requestId', 'billedDuration'])
        # print(df)

        aggregation_functions = {'timestamp': 'first', 'billedDuration': 'last'}
        df_agg = df.groupby('requestId', as_index=False).aggregate(aggregation_functions).sort_values(by=['timestamp'])
        # df_agg.index = df_agg.index.droplevel(0)
        df = df_agg.iloc[: , 1:]
        df['billedDuration'] = df.apply(lambda row: row['timestamp'] + row['billedDuration'], axis= 1)
        df.rename(columns={'timestamp': 'start', 'billedDuration': 'end'}, inplace=True)
        df.reset_index(inplace=True, drop=True)
        return df
    else:
        logging.warning("CloudWatchLog query did not find any entry?")
        exit()

def getEC2Redis(path):
    filename = os.path.basename(path)
    # TODO: read EC2-redis log file and extract cost
    filename, ext = os.path.splitext(filename)
    filename = filename+'-log.txt'
    filepath = os.path.join(os.getcwd(), 'simulation', filename)
    ec_df = pd.read_csv(filepath, header=None, sep='\t', names=['start', 'end'])
    return ec_df

def computeEC2Costs(df, end):
    # compute runtime of each EC2-redis start, if <60s round up to 60s (min billing duration)
    # except for the case where the EC2-redis instance did not shut down during the simulation duration:
    #   in this case we take the number of seconds the instance is running even if it less than 60 seconds
    # set last redis end timestamp to simulation end
    if len(df) == 0:
        logging.info("redis was never running")
        return 0, 0

    if (pd.isna(df.iloc[-1,-1])):
        logging.info("redis was running at the end of the simulation, take simulation end timestamp")
        df['runtime'] = df[:-1]['end']-df[:-1]['start']
        df['runtime'] = df['runtime'].apply(lambda x: 60.0 if x < 60.0 else x)

        df.iloc[-1, 1] = end
        df.iloc[-1, 2] = df.iloc[-1, 1]-df.iloc[-1, 0]
    else:
        df['runtime'] = df['end']-df['start']
        df['runtime'] = df['runtime'].apply(lambda x: 60.0 if x < 60.0 else x)

    # print(ec_df)
    ec2_redis_invokes = len(df)
    ec2_redis_runtime = df['runtime'].sum()
    
    return ec2_redis_invokes, ec2_redis_runtime


def getEC2Costs(filename, end):
    logging.info("get EC2-redis costs")

    # TODO: read EC2-redis log file and extract cost
    ec_df = getEC2Redis(filename)

    return computeEC2Costs(ec_df, end)

def addSystemCosts(f, df, start, end, filename):
    logging.info("add system costs")
    total_cost = 0.0

    # Lambda
    lambda_runtime, lambda_requests, lambda_coldStarts = getLambdaCosts(start, end)
    lambda_runtime_cost = lambda_runtime * lambda_ms_cost
    total_cost += lambda_runtime_cost
    f.write('{0:<25}  {1:<15}\n'.format("Lambda compute cost:", np.format_float_positional(lambda_runtime_cost, trim='-')))
    f.write('\t  {0:<25}  {1:<15}\n'.format("billed duration (ms):", lambda_runtime))
    
    lambda_request_cost = lambda_requests * lambda_req_cost
    total_cost += lambda_request_cost
    f.write('{0:<25}  {1:<15}\n'.format("Lambda request cost:", np.format_float_positional(lambda_request_cost, trim='-')))
    f.write('\t  {0:<25}  {1:<15}\n'.format("lambda invokes:", lambda_requests))

    # EC2-Redis
    # ec2_redis_runtime in seconds
    ec2_redis_invokes, ec2_redis_runtime = getEC2Costs(filename, end)
    ec2_redis_runtime_hours = (ec2_redis_runtime/3600.0) # in hours
    ec2_cost = ec2_redis_runtime_hours * ec2_hour_cost
    total_cost += ec2_cost
    f.write('{0:<25}  {1:<15}\n'.format("EC2-Redis cost:", np.format_float_positional(ec2_cost, trim='-')))
    f.write('\t  {0:<25}  {1:<15}\n'.format("runtime (s):", ec2_redis_runtime))
    f.write('\t  {0:<25}  {1:<15}\n'.format("invokes:", ec2_redis_invokes))

    # S3
    s3_requests = len(df[df['origin'] == "S3"])
    s3_cost = (s3_requests+lambda_coldStarts+ec2_redis_invokes) * s3_request_cost
    total_cost += s3_cost
    f.write('{0:<25}  {1:<15}\n'.format("S3 request costs:", np.format_float_positional(s3_cost, trim='-')))
    f.write('\t  {0:<25}  {1:<15}\n'.format("S3 requests:", s3_requests))
    f.write('\t  {0:<25}  {1:<15}\n'.format("lambda cold starts:", lambda_coldStarts))
    f.write('\t  {0:<25}  {1:<15}\n'.format("EC2-Redis starts:", ec2_redis_invokes))

    f.write('{0:<40}\n'.format('-'*40))
    f.write('{0:<25}  {1:<15}\n'.format("total cost:", np.format_float_positional(total_cost, trim='-')))

    return total_cost

def getName(method):
    match method:
        case "S3":
            return "Simulation on S3 only"
        case "system":
            return "Simulation on our System"
        case "EC":
            return "Simulation on AWS ElastiCache"

def plot(df, trace, method, total_cost, start, end):
    sns.set_context("paper")
    # sns.set(style="ticks")
    colors = sns.color_palette()
    color_dict = dict({'S3': colors[0],
                  'lambda': colors[1],
                  'redis': colors[2],
                  'EC': colors[2]})
    fig, ax = plt.subplots(1, 1, sharex=True ,constrained_layout=True, figsize=(6,3))

    # background histplot
    ax2 = ax.twinx()
    runtime = end-start # in seconds
    bins = 60 #runtime/50.0
    logging.info("binwidth: {} seconds".format(bins))
    sns.histplot(data=df, x="delay", binwidth=bins, color='grey', ax=ax2, element="step", alpha=0.3, edgecolor='w')
    ax2.set_ylabel('Count [requests / {} seconds]'.format(bins))

    sns.scatterplot(data=df, x="delay", y="latency", hue="origin", palette=color_dict, ax=ax, alpha=1.0)
    # ax.set(xlabel='Time [s]', ylabel='Latency [ms]')
    ax.set_ylabel('Latency [ms]')
    ax.set_xlabel('Time [s]')

    avg = round(df["latency"].mean(), 2)
    txt = "avg: {} ms".format(avg)
    ax.axhline(avg, color=sns.color_palette()[3], linewidth=0.7, linestyle='--', label=txt)

    cost_txt = "total cost (USD):"
    trace_txt = "simulation trace:"
    subtitle = "{:<18} {:.6f}\n{:<18} {}".format(cost_txt, total_cost, trace_txt, trace)
    ax.set_title(subtitle, loc='left')


    ax.set_zorder(2)
    ax.set_axisbelow(True)
    # ax.grid()
    ax2.set_zorder(1)
    ax.patch.set_visible(False)
    # cost label, invisible horizontal line
    # axes[0].axhline(0, color='w', linewidth=0.0, label=cost_txt)
    ax.legend()

    # plt.plot([], [], ' ', label=cost_text)
    plt.legend()

def plotSystem(df, trace, method, start, end, filename, ec2redis_best_df, lambda_best_df, total_cost):
    sns.set_context("paper")
    # sns.set(style="ticks")
    colors = sns.color_palette()
    color_dict = dict({'S3': colors[0],
                  'lambda': colors[1],
                  'redis': colors[2],
                  'EC': colors[2]})

    runtime = end-start # in seconds
    total_duration = (runtime/3600.0) # in hours

    fig, axes = plt.subplots(3, 1, sharex=True, constrained_layout=True, gridspec_kw={'height_ratios': [6, 1, 1]})
    
    ax2 = axes[0].twinx()

    # extract trace name
    trace_name = os.path.basename(trace)
    trace_names = trace_name.split('-')
    if trace_names[3] == 'simple':
        minimum_duration = (df['delay'].iloc[-1] - df['delay'].iloc[0]) # in seconds
        minimum_duration_hr = (minimum_duration)/3600.0
        refcost_txt = f"reference redis cost ({math.ceil(minimum_duration)} s):"
        refcost = (minimum_duration_hr * ec2_hour_cost) + s3_request_cost
    else:
        refcost_txt = "reference full time redis cost (USD):"
        refcost = total_duration * ec2_hour_cost + s3_request_cost
    
    # background histplot
    bins = runtime/50.0 #60
    logging.info("binwidth: {} seconds".format(bins))
    sns.histplot(data=df, x="delay", binwidth=bins, color='grey', ax=ax2, element="step", alpha=0.3, edgecolor='w')
    ax2.set_ylabel('Count [requests / {} seconds]'.format(bins))

    cost_txt = "total cost (USD):"
    s3_requests = len(df[df['origin'] == "S3"])
    percentage = "{:.2f}".format(((len(df)-s3_requests) / len(df))*100.0)
    request_txt = "in-memory:"
    # axes[0].set_title(getName(method)+f"\nsimulation log: {trace}"+f"\n{cost_txt}"+f"\n{request_txt}")
    trace_txt = "simulation trace:"
    subtitle = "{:<18} {:.3g}\n{:<18} {:.3g}\n{:<18} {}%\n{:<18} {}".format(cost_txt, total_cost, refcost_txt, refcost, request_txt, percentage, trace_txt, trace_name)
    axes[0].set_title(subtitle, loc='left')

    # scatter plot and mean horizontal line
    sns.scatterplot(data=df, x="delay", y="latency", hue="origin", palette=color_dict, ax=axes[0], alpha=1.0)
    axes[0].set_ylabel('Latency [ms]')

    avg = round(df["latency"].mean(), 2)
    txt = "avg: {} ms".format(avg)
    axes[0].axhline(avg, color=sns.color_palette()[3], linewidth=0.7, linestyle='--', label=txt, alpha=1.0)
    
    axes[0].set_zorder(2)
    axes[0].set_axisbelow(True)
    # axes[0].grid()
    ax2.set_zorder(1)
    axes[0].patch.set_visible(False)
    # cost label, invisible horizontal line
    # axes[0].axhline(0, color='w', linewidth=0.0, label=cost_txt)
    axes[0].legend()
    # axes[0].plot([], [], ' ', label="Extra label on the legend")
    # plt.legend()

    # subplot: lambda and EC2-redis runtimes:
    axes[1].set_title("Actual Execution Plan", loc='left')

    axes[1].set_ylim(0,1)
    ec2redis_df = getEC2Redis(filename)
    if len(ec2redis_df) != 0:
        if (pd.isna(ec2redis_df.iloc[-1,-1])):
            ec2redis_df.iloc[-1,-1] = end
        ec2redis_df['start'] = ec2redis_df['start'].map(lambda s: s - start)
        ec2redis_df['end'] = ec2redis_df['end'].map(lambda e: e - start)
        # print(ec2redis_df)
        for idx, row in ec2redis_df.iterrows():
            axes[1].axvspan(int(row['start']), int(row['end']), facecolor=colors[2], alpha=0.5)


    lambda_df = getLambdaRuntimes(start, end)
    lambda_df['start'] = lambda_df['start'].map(lambda s: s - start)
    lambda_df['end'] = lambda_df['end'].map(lambda e: e - start)
    # print(lambda_df)
    for idx, row in lambda_df.iterrows():
        axes[1].axvspan(int(row['start']), int(row['end']), facecolor=colors[1], alpha=0.5)

    axes[1].set_yticks([])
    # axes[1].set(xlabel='Time [s]')

    # subplot: best lambda/redis:
    lambda_invokes, lambda_runtime = computeLambdaCosts(lambda_best_df)
    ec2_redis_invokes, ec2_redis_runtime = computeEC2Costs(ec2redis_best_df, end)
    logging.info(f"lambda runtime (s): {lambda_runtime}")
    logging.info(f"EC2-Redis runtime (s): {ec2_redis_runtime}")
    ec2_redis_runtime_hours = (ec2_redis_runtime/3600.0) # in hours
    ec2_cost = ec2_redis_runtime_hours * ec2_hour_cost

    lambda_runtime_cost = lambda_runtime * (lambda_ms_cost*1000.0)

    theory_cost = ec2_cost + lambda_runtime_cost

    axes[2].set_title("Theoretical Execution Plan (USD): {:.3g}".format(theory_cost), loc='left')
    axes[2].set_ylim(0,1)
    # print(ec2redis_df)
    for idx, row in ec2redis_best_df.iterrows():
        axes[2].axvspan(int(row['start']), int(row['end']), facecolor=colors[2], alpha=0.5)


    # print(lambda_df)
    for idx, row in lambda_best_df.iterrows():
        axes[2].axvspan(int(row['start']), int(row['end']), facecolor=colors[1], alpha=0.5)

    axes[2].set_yticks([])
    axes[2].set(xlabel='Time [s]')

def evaluate(df, method, filepath, start, end, filename):
    # TODO: get costs for simulation
    # https://boto3.amazonaws.com/v1/documentation/api/latest/reference/services/ce.html#CostExplorer.Client.get_cost_and_usage

    
    with open(filepath, "w") as f:
        getStats(f, df, start, end)

        f.write('{0:<25}\n'.format("Cost Evaluation (all USD)"))
        match method:
            case "S3":
                total_cost = addS3Costs(f, df)
                plot(df, filename, method, total_cost, start, end)
            case "EC":
                total_cost = addECCosts(f, df, start, end)
                plot(df, filename, method, total_cost, start, end)
            case "system":
                total_cost = addSystemCosts(f, df, start, end, filename)
                ec2redis_best_df, lambda_best_df = optimizeSystem(df, start, end)
                plotSystem(df, filename, method, start, end, filename, ec2redis_best_df, lambda_best_df, total_cost)

 