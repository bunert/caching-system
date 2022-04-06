from datetime import datetime
import os
import logging
import argparse
import math

from numpy import float64

import common
import traces.helper.ec2logGenerator as ec2log
import traces.helper.simpleGenerator as simple
import traces.helper.poissonGenerator as poisson
import traces.helper.spikeGenerator as spike
import traces.helper.latencyGenerator as latency

## output format:
## id, delay, method, object

def setupArgs():
    parser = argparse.ArgumentParser(description="generator")
    # general parameter
    parser.add_argument("-t", "--trace",
        dest="trace", required=True, choices={"simple", "ec2log", "poisson", "spike", "spike-simple", "latency"},
        help="specifiy trace generator")
    # optional parameter (only optional for spike)
    parser.add_argument("-d", "--duration",
        dest="duration", required=False, type=int, choices=range(0,180), metavar='PARAMETER',
        help="duration of simulation (in minutes), max 180 minutes")
    # ec2log parameter
    parser.add_argument("-r", "--rank",
        dest="rank", required=False, type=int, choices=range(0,180), metavar='PARAMETER',
        help="trace of the x-th most requested object from the ec2log file")
    # poisson parameter
    parser.add_argument("-l", "--lambda",
        dest="lambd", required=False, type=float64, metavar='PARAMETER',
        help="poisson process parameter lambda (average requests/minute)")
    # spike parameter
    parser.add_argument("-v", "--variance",
        dest="variance", required=False, type=int, choices=range(0,500), metavar='PARAMETER',
        help="variance for normal distribution modeling a spike")
    parser.add_argument("-s", "--samples",
        dest="samples", required=False, type=int, choices=range(0,800), metavar='PARAMETER',
        help="number of samples from the normal distribution")
    return parser, parser.parse_args()

if __name__ == '__main__':
    common.setupLogging()

    parser, args = setupArgs()

    match args.trace:
        case "simple":
            if args.duration is None:
                parser.error("-t simple requires --duration")
            filename = str(args.duration)+"-simple.txt"
            filepath = os.path.join(os.getcwd(), 'traces', filename)
            logging.info(f"generate new simple trace: {filename}")
            common.generate_check_output(filename)
            simple.generate_simple(filepath, args.duration)
        case "ec2log":
            if args.duration is None:
                parser.error("-t ec2log requires --duration")
            if args.rank is None:
                parser.error("-t ec2log requires --rank")

            filename = str(args.duration)+"-"+str(args.rank)+"-ec2log.txt"
            filepath = os.path.join(os.getcwd(), 'traces', filename)
            logging.info(f"generate new ec2log trace: {filename}")
            common.generate_check_output(filename)

            ec2log.generate(filepath, args.duration, args.rank)
        case "poisson":
            if args.duration is None:
                parser.error("-t poisson requires --duration")
            if args.lambd is None:
                parser.error("-t poisson requires --lambda")

            filename = str(args.duration)+"-poisson-"+str(args.lambd)+".txt"
            filepath = os.path.join(os.getcwd(), 'traces', filename)
            logging.info(f"generate new poisson trace: {filename}")
            common.generate_check_output(filename)
            logging.info(f"using poisson parameter (lambda): {args.lambd}")

            poisson.generate_poisson(filepath, args.duration, args.lambd)
        case "spike":
            if args.duration is None:
                parser.error("-t poisson requires --duration")
            if args.variance is None:
                parser.error("-t spike requires --variance")
            if args.samples is None:
                parser.error("-t spike requires --samples")

            filename = str(args.duration)+"-spike-"+ str(args.variance)+"-"+str(args.samples)+ ".txt"
            filepath = os.path.join(os.getcwd(), 'traces', filename)
            logging.info(f"generate new spike trace: {filename}")
            common.generate_check_output(filename)
            logging.info(f"using variance: {args.variance} seconds")

            spike.generate_spike(filepath, args.duration, 1, args.variance, args.samples)
        case "spike-simple":
            if args.variance is None:
                parser.error("-t spike requires --variance")
            if args.samples is None:
                parser.error("-t spike requires --samples")


            filename = spike.generate_spike_simple(1, args.variance, args.samples)
        case "latency":
            if args.samples is None:
                parser.error("-t spike requires --samples")

            duration = math.ceil((args.samples * 10)/60) # duration in minutes (round up)
            filename = str(duration)+"-latency-" + str(args.samples)+".txt"
            filepath = os.path.join(os.getcwd(), 'traces', filename)
            logging.info(f"generate new latency trace: {filename}")
            common.generate_check_output(filename)
            latency.generate_latency(filepath, args.samples)


    logging.info(f"trace successfully generated: {filename}")


