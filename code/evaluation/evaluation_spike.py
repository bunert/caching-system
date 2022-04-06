import os
import logging
import argparse
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import common
import random
import click
import math


def read_trace(path):
    filepath = os.path.join(os.getcwd(), path)
    df = pd.read_csv(filepath, header=None, sep='\t', names=['id', 'delay', 'method', 'object'])

    df['origin'] = "spike"
    return df

# check output file for simulator
def check_output(path):
    # checks that output file does not exists
    filename = os.path.basename(path)
    name = os.path.splitext(filename)[0]
    filename = name+'.pdf'
    filepath = os.path.join(os.getcwd(), 'evaluation', filename)
    if os.path.exists(filepath):
        if click.confirm("The corresponding output file already exists, do you want to overwrite it?", default=False):
            logging.info("overwrite output file")
        else:
            logging.warning("if you don't overwrite the output file, resolve manually before evaluation can be started")
            exit()
    return name

def setupArgs():
    parser = argparse.ArgumentParser(description="spike evaluation")
    parser.add_argument("-i", "--input",
        dest="filename", required=True, type=common.simulate_check_input,
        help="input file with access trace", metavar="FILE")
    return parser, parser.parse_args()

def plot_single(axes, position, df, binwidth, title):
    colors = sns.color_palette()
    color_dict = dict({'spike': colors[0],
                  'poisson': colors[1]})
    axes[position].set_axisbelow(True)
    # axes[position].grid()
    axes[position].patch.set_visible(False)
    sns.histplot(data=df, x="delay", binwidth=binwidth, palette=color_dict, hue='origin', ax=axes[position], element="step", alpha=0.3, multiple="stack")
    axes[position].set_ylabel('count')
    axes[position].set_title(title)

def plot_distributions(duration, df_list, lambd):
    sns.set_context("paper")

    elements = len(df_list)
    logging.info(f"number of datasets: {elements}")
    rows = 3

    fig, axes = plt.subplots(rows, 1, sharex=True, constrained_layout=True, gridspec_kw={'height_ratios': [1, 1, 2]})
    
    runtime = duration * 60
    binwidth = runtime/50.0

    for idx, (title, df) in enumerate(df_list):
        plot_single(axes, idx, df, binwidth, title)

    capacity = (1.25 * lambd) # requests/minute the base system can handle
    capacity_bin = (capacity/60.0)*binwidth
    avg = math.ceil(capacity_bin)
    logging.info(f"base instance handling: {capacity} requests/minute")
    txt = "base"
    axes[2].axhline(avg, color=sns.color_palette()[3], linewidth=1.0, linestyle='--', label=txt, alpha=1.0)
    title_txt = f"total distribution (base instance capacity: {capacity} requests/minute)"
    axes[2].set_title(title_txt)
    # axes[2].legend()

    axes[2].set_xlabel('Time [s]')

    title = f"Modeling how a spike exceeds the capacity of a base system.\nbindwidth: {binwidth} seconds"
    fig.suptitle(title)

def get_poisson(duration, lambd):
    duration_sec = duration*60
    cur = 0

    data = []
    while(True):
        # https://stackoverflow.com/questions/1155539/how-do-i-generate-a-poisson-process
        delay = random.expovariate(lambd/60.0)

        cur = round(cur+delay, 0)
        if cur > duration_sec:
            break
        data.append([cur, "GET", "index.html", "poisson"])
        # df = df.append({'id': idx, 'delay': cur, 'method': "GET", 'object': "index.html"}, ignore_index = True)

    df = pd.DataFrame(data, columns=['delay', 'method', 'object', 'origin'])
    # df = df.sort_values(by=['delay'])
    # df = df.reset_index(drop=True)
    return df

if __name__ == '__main__':
    common.setupLogging()

    parser, args = setupArgs()

    name = check_output(args.filename)

    logging.info("make possible spike derivation")
    df_list = []

    spike_df = read_trace(args.filename).drop(columns='id')
    filename = os.path.basename(args.filename)
    name = os.path.splitext(filename)[0] # remove file extension
    name_parts = name.split('-')
    samples = name_parts[-1]
    variance = name_parts[-2]
    duration = int(name_parts[0])

    spike_txt = f"normal distribution (variance: {variance}, samples: {samples})"
    df_list.append((spike_txt, spike_df))


    lambd = 40
    poisson_df = get_poisson(duration, lambd)
    df_list.append((f"poisson process ({lambd} requests/minute)", poisson_df))

    df_total = pd.concat([spike_df, poisson_df])
    df_total = df_total.sort_values(by=['origin'], ascending=False)
    df_total = df_total.reset_index(drop=True)

    df_list.append(("total distribution", df_total))

    plot_distributions(duration, df_list, lambd)

    output_filename = name+".pdf"
    output = os.path.join(os.getcwd(), 'evaluation', output_filename)
    plt.savefig(output, bbox_inches='tight')

    logging.info(f"evaluation finished: {output_filename}")


