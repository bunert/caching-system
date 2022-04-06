import argparse
import os
import logging
import boto3
from collections import defaultdict
import sys


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
    parser.add_argument("-i", "--input",
        dest="filename", required=True, type=common.simulate_check_input,
        help="input file with access trace", metavar="FILE")
    parser.add_argument("-s", "--system",
        dest="system", required=True, choices={"S3", "EC", "system"},
        help="specify system used for simulation")
    parser.add_argument("-n", "--name",
        dest="name", required=True, type=str,
        help="specify test name")

    parser.add_argument("-o", "--object", 
        dest="object", required=True, choices={"100B.txt", "1KB.txt", "1MB.txt", "100MB.txt"},
        help="specify object to query (different sizes)") # flag with default False

    return parser.parse_args()

if __name__ == '__main__':
    common.setupLogging()
    
    # argparser
    args = setupArgs()

    # S3 only:
    if args.system == "S3":
        logging.info(f"Simulate on S3")
        system = "S3"

        addr = addr.getS3Addr()
        request.gateway_addr = "http://"+addr+":4000/api/v1/objects/"
        request.sources = ["S3"]
        logging.info(f"using the trace: {args.filename}")
    
    # ElastiCache only: 
    elif args.system == "EC":
        logging.info(f"Simulate on ElastiCache")
        system = "EC"

        addr = addr.getECAddr()
        request.gateway_addr = "http://"+addr+":4000/api/v1/objects/"
        request.sources = ["S3", "EC"]
        logging.info(f"using the trace: {args.filename}")
    
    # our system:
    elif args.system == "system":
        logging.info(f"Simulate on our System")
        system = "system"

        gateway_addr, orchestrator_addr = addr.getSystemAddr()
        request.gateway_addr = "http://"+gateway_addr+":4000/api/v1/objects/"
        
        logging.info(f"using the trace: {args.filename}")
        
        request.sources = ["S3", "redis", "lambda"]

    else:
        logging.warning("System used for simulation not specified, abort")
        exit(0)

    output_filename, output_filepath = common.latency_check_output(args.name, args.filename, system, args.object)

    results, start, end = request.start(args.filename, args.object)

    common.write_results(results, output_filepath, system, start, end)

    logging.info(f"write results to file (simulate/): {output_filename}")