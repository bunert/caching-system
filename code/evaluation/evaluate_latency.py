import argparse
import os
import logging
import pandas as pd
import seaborn as sns
import matplotlib.pyplot as plt

import evaluation.helper.stats as stats
import common

def read_proxy(filepath, column, skip):
    if not os.path.isfile(filepath):
        logging.error("file does not exist?")
        exit()

    df = pd.read_csv(filepath, header=None, skiprows=skip, sep='\s+', usecols=[column]) 

    return df

def read_client(filepath, skip):
    if not os.path.isfile(filepath):
        logging.error("file does not exist?")

    df = pd.read_csv(filepath, header=None, skiprows=skip+1, sep='\t', names=['id', 'delay', 'method', 'object', 'origin', 'latency'], usecols=['latency'])
    
    return df

def getFilename(method):
    match method:
        case "System S3":
            return "system-5-latency-30-s3-"
        case "System Lambda":
            return "system-5-latency-30-lambda-"
        case "System Redis":
            return "system-5-latency-30-redis-"
        case "S3":
            return "S3-5-latency-30-S3-"
        case "ElastiCache":
            return "EC-5-latency-30-EC-"

def getMS(str):
    time_unit = ''.join([i for i in str if (not i.isdigit() and i != '.')])
    latency = float(''.join([i for i in str if (i.isdigit() or i == '.')]))
    match time_unit:
        case "µs":
            return latency/1000
        case "ms":
            return latency
        case "s":
            return latency*1000




def processExperiments(tests, sizes):
    directory = os.path.join(os.getcwd(), 'simulation/latency_simulations')
    proxy_dict = {}
    client_dict = {}

    proxy_df_100B = pd.DataFrame(columns=tests)
    proxy_df_1KB = pd.DataFrame(columns=tests)
    proxy_df_1MB = pd.DataFrame(columns=tests)
    proxy_dict[sizes[0]] = proxy_df_100B
    proxy_dict[sizes[1]] = proxy_df_1KB
    proxy_dict[sizes[2]] = proxy_df_1MB

    client_df_100B = pd.DataFrame(columns=tests)
    client_df_1KB = pd.DataFrame(columns=tests)
    client_df_1MB = pd.DataFrame(columns=tests)
    client_dict[sizes[0]] = client_df_100B
    client_dict[sizes[1]] = client_df_1KB
    client_dict[sizes[2]] = client_df_1MB

    for test in tests:
        for size in sizes:
            skip = 0
            base_name = getFilename(test)
            # print(base_name+size)

            # process log file
            system = base_name.split('-')[0]
            if (system == 'system'):
                latency_column = 7
            else:
                latency_column = 6

            if (system == 'EC'):
                skip = 1

            proxy_filepath = os.path.join(directory, base_name+size+'-log.txt')
            proxy_df = read_proxy(proxy_filepath, latency_column, skip)
            proxy_df = proxy_df.apply(lambda x: getMS(x[latency_column]), axis=1)
            proxy_dict[size][test] = proxy_df

            # # process main file
            client_filepath = os.path.join(directory, base_name+size+'.txt')
            client_df = read_client(client_filepath, skip)
            client_dict[size][test] = client_df

            # df = pd.concat([proxy_df, client_df], axis=1, ignore_index=True)
            # df.columns = ['proxy', 'client']
            
    return proxy_dict, client_dict

def evaluate(data_dict, tests, size, name, filename):
    sns.set_context("paper")
    colors = sns.color_palette()

    fig, axes = plt.subplots(1, 2, figsize=(6, 2.5), sharex=False,constrained_layout=True, gridspec_kw={'height_ratios': [1], 'width_ratios': [3, 2]})
    # fig.suptitle(name)
    axes[0].set_axisbelow(True)
    axes[1].set_axisbelow(True)

    data_100B = pd.melt(data_dict[size])
    data_100B_f = data_100B.loc[data_100B['variable'].isin([tests[1], tests[2], tests[4]])]
    boxplot = sns.boxplot(x="variable", y="value", data=data_100B_f, ax=axes[0])
    boxplot.set_xlabel(None)
    boxplot.set_title("In-Memory Layer")
    boxplot.set_ylabel("Latency [ms]")

    data_100B_S = data_100B.loc[data_100B['variable'].isin([tests[0], tests[3]])]
    boxplot2 = sns.boxplot(x="variable", y="value", data=data_100B_S, ax=axes[1])
    boxplot2.set_xlabel(None)
    boxplot2.set_title("Storage Layer")
    boxplot2.set_ylabel(None)

    output = filename+'.pdf'
    filepath = os.path.join(os.getcwd(), 'evaluation', output)
    plt.savefig(filepath, bbox_inches='tight')

    return

if __name__ == '__main__':
    common.setupLogging()
    logging.info("Evaluate latency simulation")

    tests = ['System S3', 'System Lambda', 'System Redis', 'S3', 'ElastiCache']
    sizes = ['100B', '1KB', '1MB']

    proxy_dict, client_dict = processExperiments(tests, sizes)

    print("processed experiment data")

    evaluate(proxy_dict, tests, sizes[0], "Proxy Latency Measurements (size: 100B)", 'proxy_latency_100B')
    evaluate(proxy_dict, tests, sizes[1], "Proxy Latency Measurements (size: 1KB)", 'proxy_latency_1KB')
    evaluate(proxy_dict, tests, sizes[2], "Proxy Latency Measurements (size: 1MB)", 'proxy_latency_1MB')

    evaluate(client_dict, tests, sizes[0], "Client Latency Measurements (size: 100B)", 'client_latency_100B')
    evaluate(client_dict, tests, sizes[1], "Client Latency Measurements (size: 1KB)", 'client_latency_1KB')
    evaluate(client_dict, tests, sizes[2], "Client Latency Measurements (size: 1MB)", 'client_latency_1MB')

    logging.info("Latency Evaluation done")

