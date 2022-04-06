import csv
import random

## output format:
## id, delay, method, object

# duration in minutes
def generate_latency(filepath, samples):

    with open(filepath, "w") as f:
        writer = csv.writer(f, delimiter="\t")
        delay = 5.0
        for i in range (samples):
            # float delay between 0 and 3 seconds

            writer.writerow([i, delay, "GET"])
            delay += 10.0

