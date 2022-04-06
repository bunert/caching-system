import os
import logging
import argparse
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import common
import math


def read_trace(name):
    filepath = os.path.join(os.getcwd(), 'traces', name)
    df = pd.read_csv(filepath, header=None, sep='\t', names=['id', 'delay', 'method', 'object'])

    return df

def setupArgs():
    parser = argparse.ArgumentParser(description="generator")
    return parser, parser.parse_args()

def plot_single(axes, position, df, binwidth, title):
    x_position = int(position / 2)
    y_position = position % 2

    axes[x_position, y_position].set_axisbelow(True)
    axes[x_position, y_position].grid()
    axes[x_position, y_position].patch.set_visible(False)
    sns.histplot(data=df, x="delay", binwidth=binwidth, ax=axes[x_position, y_position], element="step", alpha=0.3)
    if (y_position == 0):
        axes[x_position, y_position].set_ylabel('count')
    else:
        axes[x_position, y_position].set_ylabel('')
    axes[x_position, y_position].set_title(title)

def plot_distributions(duration, df_list):
    sns.set_context("paper")
    # sns.set(style="ticks")
    colors = sns.color_palette()
    # color_dict = dict({'S3': colors[0],
    #               'lambda': colors[1],
    #               'redis': colors[2],
    #               'EC': colors[2]})
    elements = len(df_list)
    logging.info(f"number of datasets: {elements}")
    rows = math.ceil((elements / 2))
    columns = 2

    fig, axes = plt.subplots(rows, columns, sharex=True, constrained_layout=True)
    
    runtime = duration * 60
    binwidth = runtime/50.0

    for idx, (title, df) in enumerate(df_list):
        plot_single(axes, idx, df, binwidth, title)

    axes[rows-1, 0].set_xlabel('Time [s]')
    axes[rows-1, 1].set_xlabel('Time [s]')

    title = f"Request distributions for the traces used for the simulation\nbindwidth: {binwidth} seconds"
    fig.suptitle(title)


if __name__ == '__main__':
    common.setupLogging()

    parser, args = setupArgs()

    logging.info("evaluate given trace distributions")
    df_list = []

    df_list.append(("poisson process using lambda 3.0",read_trace("30-poisson-3.txt"))) # 3 requests / minute
    df_list.append(("poisson process using lambda 6.0", read_trace("30-poisson-6.txt"))) # 6 requests / minute 

    df_list.append(("ec2log trace for the rank 1 object",read_trace("30-ec2log.txt"))) # rank 1
    df_list.append(("ec2log trace for the rank 5 object",read_trace("30-5-ec2log.txt"))) # rank 5

    df_list.append(("normal distribution N(900, 50)",read_trace("30-spike-50-60.txt"))) # N(900,50)
    df_list.append(("normal distribution N(900, 100)",read_trace("30-spike-100-60.txt"))) # N(900,100)

    df_list.append(("normal distribution N(900, 150)",read_trace("30-spike-150-60.txt"))) # N(900,150)
    df_list.append(("normal distribution N(900, 200)",read_trace("30-spike-200-60.txt"))) # N(900,200)

    plot_distributions(30, df_list)

    filename = "30-trace-overview.pdf"
    output = os.path.join(os.getcwd(), 'evaluation', filename)
    plt.savefig(output, bbox_inches='tight')

    logging.info(f"evaluation finished: {filename}")


