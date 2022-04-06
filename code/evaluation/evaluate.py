import argparse
import os
import logging
import pandas as pd
import seaborn as sns
import matplotlib.pyplot as plt

import evaluation.helper.stats as stats
import common

def read_input(filename):
    filepath = os.path.join(os.getcwd(), filename)
    df = pd.read_csv(filepath, header=None, sep='\t', names=['id', 'delay', 'method', 'object', 'origin', 'latency'])
    
    start = 0
    end = 0
    
    # if first entry is unequal 0, we got timestamps
    if df.iloc[0, 0] != 0:
        # logging.info("timestamp available:")
        start = int(df.iloc[0, 0].item())
        # logging.info(start)
        end = int(df.iloc[0, 1].item())
        # logging.info(end)
    else:
        # logging.warning("no timestamp available")
        logging.error("no timestamp available?")
    
    df = df.drop(index=0)
    return df, start, end
    


def setupArgs():
    parser = argparse.ArgumentParser(description="evaluator")
    parser.add_argument("-i", "--input",
        dest="filename", required=True, type=common.evaluate_check_input,
        help="input file with access trace", metavar="FILE")
    return parser.parse_args()

def getSimulatorMethod(path):
    # remove file extension
    print(path)
    filename = os.path.basename(path)
    filename = os.path.splitext(filename)[0]
    # return first part of filename (indicates simulation method)
    return filename.split('-')[0]

if __name__ == '__main__':
    common.setupLogging()

    # argparser
    args = setupArgs()
    method = getSimulatorMethod(args.filename)

    logging.info(f"Evaluate simulation file: {args.filename}")
    output, output_path, cost_output, cost_output_path = common.evaluate_check_output(args.filename)

    df, start, end = read_input(args.filename)

    stats.evaluate(df, method, cost_output_path, start, end, args.filename)

    plt.savefig(output_path, bbox_inches='tight')

    logging.info(f"Saved plot: {output}")
    logging.info(f"Saved cost: {cost_output}")

