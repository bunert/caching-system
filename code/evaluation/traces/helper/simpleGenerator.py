import csv
import random

## output format:
## id, delay, method, object

# duration in minutes
def generate_simple(filepath, duration):
    duration_sec = duration*60

    with open(filepath, "w") as f:
        writer = csv.writer(f, delimiter="\t")
        cur = 1
        idx = 0
        while(True):
            # float delay between 0 and 3 seconds
            delay = random.uniform(0, 3.0)

            cur = round(cur+delay, 3)
            if cur > duration_sec:
                break
            writer.writerow([idx, cur, "GET", "index.html"])
            idx+=1



