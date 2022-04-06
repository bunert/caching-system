#!/usr/bin/python
"""
Downloads the alexa1m dataset (https://data.openintel.nl/data/) from yesterday and extracts all NS query types for further processing. 
The extracted NS queries are safed to a new avro file in openIntelData with the corresponding date in its name. 

runtime: about 50 minutes single core

avro merging code from: https://blog.han.life/posts/2014/2014-12-02-merging-avro-files-using-python/
"""
 
import avro.schema
from avro.datafile import DataFileReader, DataFileWriter
from avro.io import DatumReader, DatumWriter
import os
import logging
import argparse
import urllib.request
from datetime import date, timedelta
import tarfile

from helpers import valid_date, setup_logger

# prepare default date value (yesterday)
today = date.today()
yesterday = today - timedelta(days=1)
datestr = yesterday.strftime("%Y%m%d")

# parse args
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
tmp_dir = os.path.join(owd, 'tmp')
# filename example: openintel-alexa1m-20210913.tar
name = datestr + '-openintel-alexa1m'

tar_file = 'openintel-alexa1m-' + datestr + '.tar'
output_file = os.path.join(output_dir, name + '.avro')
url = 'https://data.openintel.nl/data/alexa1m/2021/'+tar_file

####################################################################
# check if file already downloaded
####################################################################
if (os.path.isfile(output_file)):
    logging.error('Outputfile %s already exists', os.path.basename(output_file))
    exit()
else:
    # download data of the given date (datestr)
    if (not os.path.isdir(tmp_dir)):
        os.mkdir(tmp_dir)
    os.chdir(tmp_dir)
    # TODO: download if tmp directory is empty, if not use the avro files there (just for testing atm) 
    if (not os.listdir(os.getcwd())):
        logging.info('Start downloading tar file and extract to directory: %s', os.path.relpath(tmp_dir, owd))
        urllib.request.urlretrieve(url, tar_file)
        tar = tarfile.open(tar_file, "r:")
        tar.extractall()
        tar.close()
        if os.path.isfile(tar_file):
            os.remove(tar_file)
            logging.info('Removed tar file')
    else:
        logging.error('tmp directory not empty clean up manually (not sure if data needed)')
        exit()
 
####################################################################
# merge avro files
####################################################################
avrometa = ""
avrorecords = []
avrostats = ""
 
logging.info('Start writing merged data into: %s',os.path.relpath(output_file, owd))

for target_file in os.listdir(os.getcwd()):
    if target_file.endswith(".avro"):
        avrorecords = []
        logging.info('merging: %s',target_file)
        try:
            # target_rows = DataFileReader(open(target_file, "r"), DatumReader())
            target_rows = DataFileReader(open(target_file ,"rb"), DatumReader())
        except Exception as e:
            raise avro.schema.AvroException(e)
        #need to capture very first file's first line for avrometa, otherwise skip first line to remove avrometa
        if avrometa != "" and avrostats != "":
            next(target_rows)
        for row in target_rows:
            # filter for NS records (see notion for more details why we only keep this response type)
            if (row["query_type"] == "NS" and row["response_type"] == "NS"):
                avrorecords.append(row)

        #capture avrometa(header) and bogus avrostats(footer)
        if avrometa == "":
            avrometa = avrorecords[0]
            schema_json = target_rows.get_meta('avro.schema')
            schema = avro.schema.parse(schema_json)
            writer = DataFileWriter(open(output_file, "wb"), DatumWriter(), schema)
            writer.append(avrometa)
            # logging.info('avrometa: %s',avrometa)
            del avrorecords[0]
        if avrostats == "":
            avrostats = avrorecords[-1]
            # logging.info('avrostats: %s',avrostats)
        #remove avrostats then append records
        del avrorecords[-1]
        for avrorecord in avrorecords:
            writer.append(avrorecord)
        target_rows.close()
 
writer.append(avrostats)
writer.close()

####################################################################
# cleanup tmp directory
####################################################################
logging.info('Remove directory %s', os.path.relpath(tmp_dir, owd))
for file in os.listdir(os.getcwd()):
    if file.endswith(".avro"):
        # logging.info('Remove file %s', file)
        os.remove(file)
os.chdir(owd)
os.rmdir(tmp_dir)

logging.info('Writing finished. All done.')
