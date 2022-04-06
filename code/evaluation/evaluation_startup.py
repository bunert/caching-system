import os
import logging
import pandas as pd
import seaborn as sns
import matplotlib.pyplot as plt

import evaluation.helper.stats as stats
import common

def read_data(filepath):
    if not os.path.isfile(filepath):
        logging.error("file does not exist?")
        exit()

    df = pd.read_csv(filepath, header=None, sep='\s+', usecols=[5]) 

    return df

def getFilename(method):
    match method:
        case "10B":
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


def process_lambda(df, lst, name):
    for i in range(30):
            idx = i*3
            start = df.iloc[idx+1][5]
            end = df.iloc[idx+2][5]
            latency = end-start
            lst.append([name, latency])


def process_redis(df, lst, name):
    for i in range(30):
        idx = i*9
        start = df.iloc[idx+7][5]
        end = df.iloc[idx+8][5]
        latency = end-start
        lst.append([name, latency])

def processExperiments(tests):
    directory = os.path.join(os.getcwd(), 'simulation/startup_simulations')
    size_dict = {}

    df_coldStart = pd.DataFrame(columns=['variable', 'value'])
    df_warmStart = pd.DataFrame(columns=['variable', 'value'])
    df_redis = pd.DataFrame(columns=['variable', 'value'])
    size_dict[tests[0]] = df_coldStart
    size_dict[tests[1]] = df_warmStart
    size_dict[tests[2]] = df_redis

    for test in tests:
        lst = []
        base_name = "startup_simulations_"
        
        match test:
            case "coldStart":
                filepath = os.path.join(directory, base_name+'100B_'+test+'.txt')
                df = read_data(filepath)
                process_lambda(df, lst, "100B")

                filepath = os.path.join(directory, base_name+'1KB_'+test+'.txt')
                df = read_data(filepath)
                process_lambda(df, lst, "1KB")

                filepath = os.path.join(directory, base_name+'1MB_'+test+'.txt')
                df = read_data(filepath)
                process_lambda(df, lst, "1MB")

            case "warmStart":
                filepath = os.path.join(directory, base_name+'100B_'+test+'.txt')
                df = read_data(filepath)
                process_lambda(df, lst, "100B")

                filepath = os.path.join(directory, base_name+'1KB_'+test+'.txt')
                df = read_data(filepath)
                process_lambda(df, lst, "1KB")

                filepath = os.path.join(directory, base_name+'1MB_'+test+'.txt')
                df = read_data(filepath)
                process_lambda(df, lst, "1MB")
            case "Redis":
                filepath = os.path.join(directory, base_name+'100B_'+test+'.txt')
                df = read_data(filepath)
                process_redis(df, lst, "100B")

                filepath = os.path.join(directory, base_name+'1KB_'+test+'.txt')
                df = read_data(filepath)
                process_redis(df, lst, "1KB")

                filepath = os.path.join(directory, base_name+'1MB_'+test+'.txt')
                df = read_data(filepath)
                process_redis(df, lst, "1MB")

        # # coldStart:
        # print("cold start")
        # filepath = os.path.join(directory, base_name+'100B_'+test+'.txt')
        # coldStart_df = read_data(filepath)
        # for i in range(30):
        #     idx = i*3
        #     start = coldStart_df.iloc[idx+1][5]
        #     end = coldStart_df.iloc[idx+2][5]
        #     latency = end-start
        #     list.append(["Cold-Start", latency])

        # # warmStart:
        # print("warm start")
        # filepath = os.path.join(directory, base_name+'_warmStart.txt')
        # warmStart_df = read_data(filepath)
        # for i in range(30):
        #     idx = i*3
        #     start = warmStart_df.iloc[idx+1][5]
        #     end = warmStart_df.iloc[idx+2][5]
        #     latency = end-start
        #     list.append(["Warm-Start", latency])

        # # Redis:
        # print("Redis")
        # filepath = os.path.join(directory, base_name+'_Redis.txt')
        # redis_df = read_data(filepath)
        # for i in range(30):
        #     idx = i*9
        #     start = redis_df.iloc[idx+7][5]
        #     end = redis_df.iloc[idx+8][5]
        #     latency = end-start
        #     list.append(["Redis", latency])

        size_dict[test] = pd.DataFrame(lst, columns =['variable', 'value'])
            
    return size_dict

def getTitle(method):
    match method:
        case "coldStart":
            return "AWS Lambda cold-start"
        case "warmStart":
            return "AWS Lambda warm-start"
        case "Redis":
            return "Self-Hosted Redis"


def evaluate(data_dict, tests, sizes):
    sns.set_context("paper")
    colors = sns.color_palette()

    fig, axes = plt.subplots(3, 1, sharex=False,constrained_layout=True, gridspec_kw={'height_ratios': [1,1,1], 'width_ratios': [1]})
    # fig.suptitle(name)
    row = 0
    for test in tests:
        # axes[row].grid()
        # axes[row].grid()
        axes[row].set_axisbelow(True)
        axes[row].set_axisbelow(True)
        
        df = data_dict[test]
        # pd.set_option("display.max_rows", None, "display.max_columns", None)
        boxplot = sns.boxplot(x="variable", y="value", data=df, ax=axes[row])
        boxplot.set_xlabel(None)
        boxplot.set_title(getTitle(test))
        boxplot.set_ylabel("Startup Time [ms]")

        # data_redis = df.loc[df['variable'].isin([tests[2]])]
        # print(data_redis)
        # boxplot2 = sns.boxplot(x="variable", y="value", data=data_redis, ax=axes[row, 1])
        # boxplot2.set_xlabel(None)
        # boxplot2.set_title("Self-Hosted Redis")
        # boxplot2.set_ylabel(None)

        row+=1

    filepath = os.path.join(os.getcwd(), 'evaluation', 'startup.pdf')
    plt.savefig(filepath, bbox_inches='tight')

    return

if __name__ == '__main__':
    common.setupLogging()
    logging.info("Evaluate startup simulation")

    tests = ['coldStart', 'warmStart', 'Redis']
    sizes = ['100B', '1KB', '1MB']

    size_dict = processExperiments(tests)

    logging.info("processed experiment data")

    evaluate(size_dict, tests, sizes)


    logging.info("Latency Evaluation done")

