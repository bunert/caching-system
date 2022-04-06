import csv
import random
import numpy as np
import logging
import os
import common

## output format:
## id, delay, method, object

# duration in minutes
def generate_spike(filepath, duration, number_of_spikes, variance, samples):
    duration_sec = duration*60

    spike_arrivals = []

    for i in range(number_of_spikes):
        # spike_arrivals.append(round(random.uniform(0, duration_sec), 0))
        spike_arrivals.append(900)

    logging.info(f"spike arrival times (mean of normal distribution): {np.sort(spike_arrivals)}")

    arrival_list = []
    for arrival in spike_arrivals:
        arrival_list.append(np.random.normal(arrival, variance, size=samples))

    all_arrivals = np.concatenate(arrival_list).ravel()
    all_arrivals = np.sort(all_arrivals)

    with open(filepath, "w") as f:
        writer = csv.writer(f, delimiter="\t")
        idx = 0
        for arrival in all_arrivals:

            t = round(arrival, 3)
            if 0.0 <= t and t < duration_sec:
                writer.writerow([idx, t, "GET", "index.html"])
                idx+=1

# duration in minutes
def generate_spike_simple(number_of_spikes, variance, samples):

    spike_arrivals = []

    for i in range(number_of_spikes):
        spike_arrivals.append(0)

    logging.info(f"spike arrival times (mean of normal distribution): {np.sort(spike_arrivals)}")

    arrival_list = []
    for arrival in spike_arrivals:
        arrival_list.append(np.random.normal(arrival, variance, size=samples))

    all_arrivals = np.concatenate(arrival_list).ravel()
    all_arrivals = np.sort(all_arrivals)

    first = abs(all_arrivals[0]) # offset of first request
    all_arrivals = all_arrivals+(first + 10.0)

    last = all_arrivals[-1]

    end = round(last+120.0) # add 2 minutes to the duration after last request
    duration = round(end/60.0) # round up to a multiple of 60 seconds
    
    filename = str(duration)+"-spike-simple-"+ str(variance)+"-"+str(samples)+ ".txt"
    filepath = os.path.join(os.getcwd(), 'traces', filename)
    logging.info(f"generate new spike trace: {filename}")
    common.generate_check_output(filename)
    logging.info(f"using variance: {variance} seconds")

    with open(filepath, "w") as f:
        writer = csv.writer(f, delimiter="\t")
        idx = 0
        for arrival in all_arrivals:
            t = round(arrival, 3)
            writer.writerow([idx, t, "GET", "index.html"])
            idx+=1

    return filename


