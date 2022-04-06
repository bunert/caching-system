import paramiko
import os
import shutil
import logging


def getLogFileRemote(name, filename, addr):
    filename, ext = os.path.splitext(filename)
    filename = filename+'-log.txt'
    filepath = os.path.join(os.getcwd(), 'simulation', filename)

    keypath = os.path.join(os.getcwd(), '..', 'aws-keypairs', 'ec2-orchestrator.pem' )
    key = paramiko.RSAKey.from_private_key_file(keypath)
    
    transport = paramiko.Transport((addr, 22))
    transport.connect(username="ec2-user", pkey=key)

    sftp = paramiko.SFTPClient.from_transport(transport)
    sftp.get('/home/ec2-user/ec2-redis.txt' , filepath)
    logging.info(f"saved EC2-redis log (simulate/): {filename}")

def getLogFileLocal(name, filename):
    filename, ext = os.path.splitext(filename)
    filename = filename+'-log.txt'
    filepath = os.path.join(os.getcwd(), 'simulation', filename)

    logfilepath = os.path.join(os.getcwd(), '..', 'cloud-storage', 'ec2-redis.txt' )
    shutil.move(logfilepath, filepath)
    logging.info(f"saved EC2-redis log (simulate/): {filename}")
    return