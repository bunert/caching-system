import requests
import logging
import queue
import time
import os
import csv
import multiprocessing as mp
from datetime import datetime, timezone
import pytz

gateway_addr = ""
sources = []

class Request: 
    def __init__(self, id, delay, method, object):
        self.id = id 
        self.delay = delay 
        self.method = method
        self.object = object

    def execute(self, error_event):
        # http://3.126.130.79:4000/api/v1/objects/
        # http://localhost:4000/api/v1/objects/
        start = time.time()
        try:
            r = requests.get(url= gateway_addr+self.object)
        except requests.exceptions.RequestException as e:
            logging.error("request exception occurred (probably gateway/proxy not running), trigger error event")
            logging.error(e)
            error_event.set()
            exit(1)

        end = time.time()

        latency = round((end-start)*1000, 3)
        # latency in ms
        # latency = round(r.elapsed.total_seconds()*1000, 3)
        logging.info(f"executed request (id: {self.id:>3}, delay: {self.delay:>6}): {latency:>6}ms")

        # TODO: write stats to output file which is used for plotting


        # assertions about response
        assert(r.status_code == 200)
        if not 'content-origin' in r.headers:
            print(r)
            logging.error("content-origin not present for request?")
        else:
            assert(r.headers['content-origin'] in sources)

        if not 'content-length' in r.headers:
            print(r)
            logging.error("content-length not present for request?")
        else:
            assert(int(r.headers['content-length']) == 144)
        return latency, r.headers['content-origin']

    def show(self):
        logging.info(f"{self.id}\t{self.delay}\t{self.method}\t{self.object}")

def worker(q, start, results, error_event):
    while True:
        try:
            request = q.get(block=False)
        except queue.Empty:
            break

        # compute if worker should sleep until request can be executed
        wait = start+request.delay - time.monotonic()
        if wait > 0:
            time.sleep(wait)

        # if error event occured, stop worker
        if error_event.is_set():
            break
        latency, origin = request.execute(error_event)
        results.append([request.id, request.delay, request.method, request.object, origin, latency])
        
    # logging.info("worker done")

def setup_jobs(filename, q):
    filepath = os.path.join(os.getcwd(), filename)
    with open(filepath, 'r') as f:
        reader = csv.reader(f, delimiter='\t')
        for row in reader:
            q.put(Request(int(row[0]), float(row[1]), row[2], row[3]))

def get_duration(path):
    # logging.info(f"filename: {filename}")
    filename = os.path.basename(path)
    return int(filename.split('-')[0])

def start(filename):
    # number of processes working on the request queue
    nprocs = mp.cpu_count()

    logging.info(f"Number of worker: {nprocs}")
    logging.info(f"Proxy/Gateway Address: {gateway_addr}")

    tz = pytz.timezone('Europe/Berlin')

    # request queue
    q = mp.Queue()

    # setup queue from input file
    setup_jobs(filename, q)

    error_event = mp.Event()

    # Manager list for results to be appended
    manager = mp.Manager()
    results = manager.list()
    duration = get_duration(filename)
    logging.info(f"Simulation duration (minutes): {duration}")
    logging.info(f"Simulation started: {time.monotonic()} \n")
    time_start = int(datetime.now(tz).timestamp())
    # -------------------------------
    # actual simulation
    # -------------------------------

    # nprocs worker processing queue elements one by one
    start = time.monotonic()
    pool = mp.Pool(nprocs, worker, (q, start, results, error_event, ))
    
    # close prevents any more tasks from being submitted to the pool
    pool.close()

    # wait for the worker processes to finish
    pool.join()

    # if error event occured, don't write to file
    if error_event.is_set():
        logging.warning("error event triggered, something went wrong")
        exit(0)
    
    time_end = time_start+(duration*60)
    if (int(datetime.now(tz).timestamp()) > time_end):
        logging.warning("timestamp after simulation > start+duration?")
    else:
        logging.warning("workers done, but wait until duration of simulation finished")
        wait_time = time_end - int(datetime.now(tz).timestamp())
        logging.warning(f"wait_time: {wait_time} seconds (required to copy the correct EC2-Redis runtime log file from the Orchestrator)")
        time.sleep(wait_time)

    logging.info("Simulation finished")
    return results, time_start, time_end