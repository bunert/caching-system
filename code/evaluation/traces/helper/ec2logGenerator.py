
import pandas as pd
import  os
import csv
import logging


## output format:
## id, delay, method, object

# duration in minutes
def generate(filepath, duration, rank):
    df = processInput(rank)

    duration_sec = duration*60


    # print(df.head(100))
    # print("\n")

    with open(filepath, "w") as f:
        writer = csv.writer(f, delimiter="\t")
        i = 0
        for index, row in df.iterrows():
            # take requests within the first 60 minutes
            if (row['TimeStamp'] > duration_sec):
                break
            writer.writerow([i, row['TimeStamp'], row['HttpMethod'], "index.html"])
            i=i+1

# get dataframe with only relevant infos
# - TimeStamp (seonds from first request)
# - HttpMethod (only GET requests)
# - Uri
# - Bytes
def processInput(rank):
    input = os.path.join(os.getcwd(), 'traces/EC2log', 'eclog_1day.csv')
    df = pd.read_csv(input)

    # drop unused fields
    df.drop('UserAgent', axis=1, inplace=True)
    df.drop('UserId', axis=1, inplace=True)
    # df.drop('IpId', axis=1, inplace=True)
    df.drop('Referrer', axis=1, inplace=True)
    df.drop('HttpVersion', axis=1, inplace=True)
    df.drop('ResponseCode', axis=1, inplace=True)

    # get start_timestamp of first row
    start_timestamp = df['TimeStamp'].iloc[0]

    # drop if not a GET request
    indexNames = df[df['HttpMethod'] != 'GET'].index
    df.drop(indexNames , inplace=True)

    # change timestamp to difference from start_timestamp
    # number of 100 ns intervals that have elapsed since 00:00:00 UTC on 1st January, 1 A.D.
    # take difference from start and convert to seconds: /10000000
    # /10000 for ms
    df['TimeStamp'] = df['TimeStamp'].apply(lambda x: (x-start_timestamp)/10000000)

    objCount = df.groupby('Uri')['IpId'].count()
    top_objects = objCount.reset_index(name='count').sort_values(['count'], ascending=False)

    # print(top_objects.iloc[rank])
    obj_uri = top_objects.iloc[rank-1]["Uri"]
    logging.info(f"{rank} most requested element: {obj_uri}")
    

    # drop all except a specific Uri
    indexNames = df[df['Uri'] != obj_uri].index # top element
    df.drop(indexNames , inplace=True)
    
    return df


