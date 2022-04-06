import logging
import boto3
from collections import defaultdict

# check if orchestrator and gateway running on EC2 instance, abort if not
# if running, return gateway address used for the requests
def getSystemAddr():
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
        return ec2info["Gateway"]["Public IP"], ec2info["Orchestrator"]["Public IP"]
    else:
        logging.warning("Orchestrator or Gateway not running on EC2, please check before running simulation")
        exit() 

# check if S3 proxy running on EC2 instance, abort if not
# if running, return address used for the requests
def getS3Addr():
    # Connect to EC2
    ec2 = boto3.resource('ec2')

    # Get information for all running instances
    running_instances = ec2.instances.filter(Filters=[{
        'Name': 'instance-state-name',
        'Values': ['running']}])

    ec2info = defaultdict()
    for instance in running_instances:
        for tag in instance.tags:
            if 'Name'in tag['Key'] and tag['Value'] == "S3 Proxy":
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
    if "S3 Proxy" in ec2info:
        logging.info("EC2 instance overview:")
        logging.info("------------------")
        for name, instance in ec2info.items():
            for key in attributes:
                logging.info(f"{key}: {instance[key]}")
            logging.info("------------------")
        return ec2info["S3 Proxy"]["Public IP"]
    else:
        logging.warning("S3 Proxy not running on EC2, please check before running simulation")
        exit() 

# check if ElastiCache proxy running on EC2 instance, abort if not
# if running, return address used for the requests
def getECAddr():
    # Connect to EC2
    ec2 = boto3.resource('ec2')

    # Get information for all running instances
    running_instances = ec2.instances.filter(Filters=[{
        'Name': 'instance-state-name',
        'Values': ['running']}])

    ec2info = defaultdict()
    for instance in running_instances:
        for tag in instance.tags:
            if 'Name'in tag['Key'] and tag['Value'] == "ElastiCache Proxy":
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
    if "ElastiCache Proxy" in ec2info:
        logging.info("EC2 instance overview:")
        logging.info("------------------")
        for name, instance in ec2info.items():
            for key in attributes:
                logging.info(f"{key}: {instance[key]}")
            logging.info("------------------")
        return ec2info["ElastiCache Proxy"]["Public IP"]
    else:
        logging.warning("ElastiCache Proxy not running on EC2, please check before running simulation")
        exit() 