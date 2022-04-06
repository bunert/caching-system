import csv
import random

## output format:
## id, delay, method, object

# duration in minutes
def generate_poisson(filepath, duration, lambd):
    duration_sec = duration*60

    with open(filepath, "w") as f:
        writer = csv.writer(f, delimiter="\t")
        cur = 1
        idx = 0
        while(True):
            # https://stackoverflow.com/questions/1155539/how-do-i-generate-a-poisson-process
            delay = random.expovariate(lambd/60.0)

            cur = round(cur+delay, 0)
            if cur > duration_sec:
                break
            writer.writerow([idx, cur, "GET", "index.html"])
            idx+=1



