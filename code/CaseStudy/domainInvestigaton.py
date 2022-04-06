#!/usr/bin/python
"""
Takes as input the exclusive domains.

runtime: 
"""
 
import os
import sys
import logging
import argparse
from datetime import date, timedelta, datetime
import json

from helpers import valid_date, setup_logger, open_json, write_json

today = date.today()
yesterday = today - timedelta(days=1)
datestr = yesterday.strftime("%Y%m%d")

#parse args
parser = argparse.ArgumentParser(description='Merge openIntel avro files from yesterday into single avro file.')
parser.add_argument("-q", "--quiet", action="store_false", dest="verbose", help="don't print log messages to stdout", default=True)
parser.add_argument("-d", "--date", action="store", dest="arg_date", help="Define the date if you want to process data other than yesterday's data", type=valid_date, default=datestr)
args = parser.parse_args()

# set datestr to correct date (argument or default yesterday)
datestr = args.arg_date

####################################################################
# setup logger
####################################################################
setup_logger(args.verbose)

####################################################################
# input/output files handling
####################################################################
owd = os.getcwd()
output_dir = os.path.join(owd, 'openIntelData')
# filename example: 20210913-openintel-alexa1m.avro
name = datestr + '-openintel-alexa1m-non-exclusive'
input_file = os.path.join(output_dir, name + '.json')

exclusive_name = datestr + '-openintel-alexa1m-exclusive'
exclusive_input_file = os.path.join(output_dir, exclusive_name + '.json')

####################################################################
# Read data from json
####################################################################
data = open_json(input_file)
exclusive_data = open_json(exclusive_input_file)

####################################################################
# TODO: process data
####################################################################

print(json.dumps(exclusive_data, indent=4, sort_keys=True))
# print(json.dumps(data, indent=4, sort_keys=True))

# print all ns_addresses where one is from AWS
# for key,value in data.items(): #this gives you both
#     if any("aws" in address for address in value["ns_address"]):
#         print(json.dumps(value, indent=4, sort_keys=True))

# write_json(data, input_file)

####################################################################
# safe to json files
####################################################################

# stats_output_file = os.path.join(output_dir, 'stats-' + name + '.json')
# with open(stats_output_file, 'w') as fp:
#     json.dump(alexa1m_stats, fp) 



logging.info('Processing finished. All done.')
