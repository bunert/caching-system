#!/usr/bin/python
"""
Takes as input the openintel-alexa1m-date.avro file from yesterday and process them further.
Generates a json file with key equal to the response_name of the NS query and a list of all observed ns_addresses for the given domain. 

runtime: about 7 minutes single core
"""
 
import avro.schema
from avro.datafile import DataFileReader, DataFileWriter
from avro.io import DatumReader, DatumWriter
import os
import logging
import argparse
from datetime import date, timedelta
import json

from helpers import valid_date, setup_logger, open_json, write_json

# prepare default date value (yesterday)
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
#Add Logging
####################################################################
setup_logger(args.verbose)

####################################################################
# input/output files handling
####################################################################
owd = os.getcwd()
output_dir = os.path.join(owd, 'openIntelData')
# filename example: openintel-alexa1m-20210913.avro
name = datestr + '-openintel-alexa1m'
input_file = os.path.join(output_dir, name + '.avro')

####################################################################
# Group the entries by response_name
####################################################################

# helper function to print avro entries (skip null entries)
def print_record(record):
    # keys to ignore when printing
    ignore_keys = ["rtt", "timestamp", "worker_id", "response_ttl"]
    print(json.dumps({k: v for k, v in record.items() if v and k not in ignore_keys},indent=4))

logging.info('Processing and grouping NS data for %s', os.path.relpath(input_file, owd))
try:
    data_rows = DataFileReader(open(input_file ,"rb"), DatumReader())
except Exception as e:
    raise avro.schema.AvroException(e)


# dictionary with key for each observed domain name and a list of the ns_addresses associated to the given domain as value
ns_address_dictionary = dict()

for record in data_rows:
    # print(json.dumps(record, indent=4))

    if record["response_name"] in ns_address_dictionary:
        ns_address_dictionary[record["response_name"]]["ns_address"].append(record["ns_address"])

    else:
        ns_address_dictionary[record["response_name"]] = {"ns_address": [record["ns_address"]]}

logging.info('Grouping finished.')

"""
example entry:
"educationalservicesinc.com.": {
    "ns_addresses": [
        "ns39.domaincontrol.com.",
        "ns40.domaincontrol.com."
    ]
},
"""

####################################################################
# Categorize the grouped domains to exclusive or non-exclusive
####################################################################

dns_provider_list = ["akam", "awsdns", "azure-dns", "microsoftonline", "cdnet", "cloudflare", 
                    "dnsimple", "dynect", "easydns", "googledomains", "no-ip", "telindustelecom", 
                    "ultradns", "nstld", "ali", "oraclevcn", "nsone", "domaincontrol"]
      

exclusive_dict = dict()
non_exclusive_dict = dict()


# checks a domain and the NS address list if the list is exclusive or non-exclusive
def check_domain(domain, address_list):
    
    if (len(ns_address_dictionary[domain]['ns_address']) == 1):
        # if only one name server address classify as exclusive directly
        exclusive_dict[domain] = {"ns_address": address_list}
    else:
        # check if any entry contains an address from dns_provider_list, if yes and every other entry contains it as well add to exclusive
        for mdns_address in dns_provider_list:
            if any(mdns_address in address for address in ns_address_dictionary[domain]['ns_address']):
                if all(mdns_address in addr for addr in ns_address_dictionary[domain]['ns_address']):
                    exclusive_dict[domain] = {"ns_address": address_list}
                    return
                else: 
                    # if one is from the dns_provider_list but not all add to non_exclusive_
                    non_exclusive_dict[domain] = {"ns_address": address_list}
                    return

        # not part of dns_provider_list, add to exclusive
        exclusive_dict[domain] = {"ns_address": address_list}


logging.info('Split domains to exclusive and non-exclusive')
for domain in ns_address_dictionary:
    check_domain(domain, ns_address_dictionary[domain]['ns_address'])


####################################################################
# stats json file (meta data used later probably?)
####################################################################
alexa1m_stats = dict()
alexa1m_stats["entries"] = len(ns_address_dictionary.keys())
alexa1m_stats["non-exclusive"] = len(non_exclusive_dict.keys())
alexa1m_stats["exclusive"] = len(exclusive_dict.keys())

####################################################################
# safe to json files
####################################################################

stats_output_file = os.path.join(output_dir, name + '-stats' '.json')
write_json(alexa1m_stats, stats_output_file)

exclusive_output_file = os.path.join(output_dir,  name + '-exclusive' + '.json')
write_json(exclusive_dict, exclusive_output_file)

non_exclusive_output_file = os.path.join(output_dir, name + '-non-exclusive' + '.json')
write_json(non_exclusive_dict, non_exclusive_output_file)


# print(json.dumps(non_exclusive_dict, indent=4))

logging.info('Processing finished. All done.')
logging.info('alexa1m_stats %s: ', datestr)
print(json.dumps(alexa1m_stats, indent=4))
