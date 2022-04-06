import os
import argparse
import csv
import click
import logging

def simulate_check_input(x):
    # Type for argparse - checks that file exists but does not open.
    filepath = os.path.join(os.getcwd(), x)
    if not os.path.exists(filepath):
        # Argparse uses the ArgumentTypeError to give a rejection message like:
        # error: argument input: x does not exist
        raise argparse.ArgumentTypeError("{0} does not exist".format(x))

    return x

def evaluate_check_input(x):
    # Type for argparse - checks that file exists but does not open.
    filepath = os.path.join(os.getcwd(), x)
    if not os.path.exists(filepath):
        # Argparse uses the ArgumentTypeError to give a rejection message like:
        # error: argument input: x does not exist
        raise argparse.ArgumentTypeError("{0} does not exist".format(x))

    return x

# check output file for simulator
def generate_check_output(filename):
    # checks that output file does not exists
    filepath = os.path.join(os.getcwd(), 'traces', filename)
    if os.path.exists(filepath):
        if click.confirm("The corresponding output file already exists, do you want to overwrite it?", default=False):
            logging.info("overwrite output file")
        else:
            logging.warning("if you don't overwrite the output file, resolve manually before generation can be started")
            exit()
    return filename, filepath

# check output file for simulator
def simulate_check_output(name, path, simulator):
    # checks that output file does not exists
    filename = os.path.basename(path)
    filename = os.path.splitext(filename)[0]
    filename = simulator+'-'+filename+'-'+name+'.txt'
    filepath = os.path.join(os.getcwd(), 'simulation', filename)
    if os.path.exists(filepath):
        if click.confirm("The corresponding output file already exists, do you want to overwrite it?", default=False):
            logging.info("overwrite output file")
        else:
            logging.warning("if you don't overwrite the output file, resolve manually before simulation can be started")
            exit()
    return filename, filepath

# check output file for latency
def latency_check_output(name, path, simulator, object):
    # checks that output file does not exists
    object_name = object.split('.')[0]

    filename = os.path.basename(path)
    filename = os.path.splitext(filename)[0]
    filename = simulator+'-'+filename+'-'+name+'-'+object
    filepath = os.path.join(os.getcwd(), 'simulation', filename)
    if os.path.exists(filepath):
        if click.confirm("The corresponding output file already exists, do you want to overwrite it?", default=False):
            logging.info("overwrite output file")
        else:
            logging.warning("if you don't overwrite the output file, resolve manually before simulation can be started")
            exit()
    return filename, filepath

# check output file for evaluation
def evaluate_check_output(path):
    filename = os.path.basename(path)
    # checks that output file does not exists
    output = os.path.splitext(filename)[0]+'.pdf'
    filepath = os.path.join(os.getcwd(), 'evaluation', output)

    cost_output = os.path.splitext(filename)[0]+'-stats.txt'
    cost_filepath = os.path.join(os.getcwd(), 'evaluation', cost_output)

    if os.path.exists(filepath) or os.path.exists(cost_filepath):
        if click.confirm("The corresponding output files already exists, do you want to overwrite it?", default=False):
            logging.info("overwrite output file")
        else:
            logging.warning("if you don't overwrite the output file, resolve manually before evaluation can be started")
            exit()
    return output, filepath, cost_output, cost_filepath

# check output file for evaluation
def latency_evaluate_check_output():
    # checks that output file does not exists
    output = "latency.pdf"
    filepath = os.path.join(os.getcwd(), 'evaluation', output)

    if os.path.exists(filepath):
        if click.confirm("The corresponding output files already exists, do you want to overwrite it?", default=False):
            logging.info("overwrite output file")
        else:
            logging.warning("if you don't overwrite the output file, resolve manually before evaluation can be started")
            exit()
    return output, filepath

def write_results(results, output_filepath, simulator, start, end):
    with open(output_filepath, "w") as f:
        writer = csv.writer(f, delimiter="\t")
        writer.writerow([start, end])

        for elem in results:
            writer.writerow(elem)
    

def setupLogging():
    logging.basicConfig(
         format='%(asctime)s.%(msecs)03d %(levelname)-8s %(message)s',
         level=logging.INFO,
         datefmt='%Y-%m-%d %H:%M:%S')