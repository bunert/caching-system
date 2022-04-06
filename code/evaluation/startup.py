import argparse
import os
import logging
import boto3
from collections import defaultdict
import sys
import requests
import time
import subprocess

import simulation.helper.latency_request as request
import simulation.helper.addr as addr
import common

# check if orchestrator and gateway running on EC2 instance, abort if not
# if running, return gateway address used for the requests
def getIpAddr():
    # Connect to EC2
    ec2 = boto3.resource('ec2')

    # Get information for all running instances
    running_instances = ec2.instances.filter(Filters=[{
        'Name': 'instance-state-name',
        'Values': ['running']}])

    ec2info = defaultdict()
    for instance in running_instances:
        for tag in instance.tags:
            if 'Name'in tag['Key'] and tag['Value'] in ["Gateway", "Orchestrator"]:
                name = tag['Value']
                # Add instance info to a dictionary         
                ec2info[name] = {
                    'Name': name,
                    'InstanceId': instance.id,
                    'Type': instance.instance_type,
                    'State': instance.state['Name'],
                    'Private IP': instance.private_ip_address,
                    'Public IP': instance.public_ip_address,
                    'Launch Time': instance.launch_time
                    }

    attributes = ['Name', 'InstanceId', 'Type', 'State', 'Private IP', 'Public IP', 'Launch Time']
    if "Gateway" in ec2info and "Orchestrator" in ec2info:
        logging.info("EC2 instance overview:")
        logging.info("------------------")
        for name, instance in ec2info.items():
            for key in attributes:
                logging.info(f"{key}: {instance[key]}")
            logging.info("------------------")
        return ec2info["Gateway"]["Public IP"]
    else:
        logging.warning("Orchestrator or Gateway not running on EC2, please check before running simulation")
        exit() 

def setupArgs():
    parser = argparse.ArgumentParser(description="simulator")
    parser.add_argument("-s", "--system",
        dest="system", required=True, choices={"cold-start", "warm-start", "redis"},
        help="specify which startup latency to measure")

    parser.add_argument("-o", "--object", 
        dest="object", required=True, choices={"100B.txt", "1KB.txt", "1MB.txt", "100MB.txt"},
        help="specify object to query (different sizes)") # flag with default False

    return parser.parse_args()

def executeWarmStart(endpoint):
    for i in range(31):
        logging.info(f"request: {i}")
        r = requests.get(url= endpoint)
        time.sleep(1)
        r = requests.get(url= endpoint)

        time.sleep(10)

    return

def executeColdStart(endpoint):
    for i in range(30):
        logging.info(f"request: {i}")
        
        subprocess.Popen(['aws', 'lambda', 'update-function-code', '--function-name', 'get-s3-object', '--zip-file', 'fileb://../cloud-storage/build/main.zip'], stdout=subprocess.DEVNULL)
        time.sleep(3)

        r = requests.get(url= endpoint)
        time.sleep(1)
        r = requests.get(url= endpoint)

        time.sleep(10)

    return

def executeRedis(endpoint):
    for i in range(30):
        logging.info(f"request: {i}")
        
        r = requests.get(url= endpoint)
        time.sleep(1)
        r = requests.get(url= endpoint)
        time.sleep(1)
        # Lambda function running
        r = requests.get(url= endpoint)
        time.sleep(1)
        r = requests.get(url= endpoint)
        time.sleep(1)
        r = requests.get(url= endpoint)
        time.sleep(1)
        r = requests.get(url= endpoint)
        time.sleep(1)
        r = requests.get(url= endpoint)
        # Redis started

        # wait until running
        time.sleep(50)

        # wait one minute until shutdown
        time.sleep(60)
        # wait until stopped
        time.sleep(180)


    return

if __name__ == '__main__':
    common.setupLogging()
    
    # argparser
    args = setupArgs()

    gateway_addr, orchestrator_addr = addr.getSystemAddr()
    endpoint = "http://"+gateway_addr+":4000/api/v1/objects/"+args.object
    

    if args.system == "cold-start":
        logging.info("simulation for cold-start measurements")
        executeColdStart(endpoint)
    elif args.system == "warm-start":
        logging.info("simulation for warm-start measurements")
        executeWarmStart(endpoint)
    elif args.system == "redis":
        logging.info("simulation for Redis measurements")
        executeRedis(endpoint)
    else:
        logging.warning("System used for simulation not specified, abort")
        exit(0)



    logging.info("simulation done")