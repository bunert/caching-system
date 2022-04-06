# Building a Reactive Caching System with Serverless Computing

Cloud-based services are becoming increasingly popular, with many new services emerging or existing services being adapted to current requirements. The same applies to cloud storage: the various services are optimized for a certain level regarding the trade-off between cost and latency. Latency is critical to the user experience, so it is essential to keep end-to-end latency as low as possible across the technology stack. Cloud storage is no exception, and in-memory caching can play a critical role in meeting desired latency requirements. Current fully managed in-memory caching systems suffer from the low flexibility of their pricing model in specific settings. In this work, we combine a well-known in-memory caching technology with serverless computing to build a custom managed system. Serverless computing enables our reactive system design to provide responsiveness through fast spin-up times, which is critical in terms of cost and latency. We demonstrate the applicability of serverless computing in a novel approach to providing in-memory caching. The results show that our reactive design with serverless computing can outperform managed services under certain conditions and achieve high cache hit rates despite its reactive approach.

Thesis: [Building a Reactive Caching System with Serverless Computing](https://gitlab.ethz.ch/nsg/students/projects/2021/ma-2021-46-tobias-buner-building-a-multicloud-based-serverless-web-ecosystem/-/blob/master/final-thesis.pdf)

## Overview
In the following, we give a short overview of this GitLab repository. It contains the work presented in the thesis and some additional work during the process of this thesis. The source files for the report are in the report_new directory, while the presentation is in the presentation directory. Everything else is located in the code directory shown below. We then present our system's general setup and execution, with a brief introduction to the evaluation environment we used in this work.
* [cloud-storage](https://gitlab.ethz.ch/nsg/students/projects/2021/ma-2021-46-tobias-buner-building-a-multicloud-based-serverless-web-ecosystem/-/tree/master/code/cloud-storage): It contains all the code for our system. It includes building scripts to generate the executable binaries, providing the functionality to transfer the binary to the EC2 instance for execution. It also contains the script to build the zip file for the AWS Lambda function, which also updates the function.
* [evaluation](https://gitlab.ethz.ch/nsg/students/projects/2021/ma-2021-46-tobias-buner-building-a-multicloud-based-serverless-web-ecosystem/-/tree/master/code/evaluation): It contains all the Python scripts for the evaluation part of our work. The scripts simulate the client sending queries to our system and processing the results for the numbers presented in the report.
* [aws-keypairs](https://gitlab.ethz.ch/nsg/students/projects/2021/ma-2021-46-tobias-buner-building-a-multicloud-based-serverless-web-ecosystem/-/tree/master/code/aws-keypairs): Contains the SSH key pairs for the EC2 instances on the AWS platform.
* [datasets](https://gitlab.ethz.ch/nsg/students/projects/2021/ma-2021-46-tobias-buner-building-a-multicloud-based-serverless-web-ecosystem/-/tree/master/code/datasets/EClog): Contains scripts for the initial record evaluation of the web server log used in our evaluation (not relevant for the evaluation in this paper).
* [CaseStudy](https://gitlab.ethz.ch/nsg/students/projects/2021/ma-2021-46-tobias-buner-building-a-multicloud-based-serverless-web-ecosystem/-/tree/master/code/CaseStudy): Contains scripts for the DNS case study we did at the beginning regarding several DNS service providers.

**Important**: The cloud-storage directory contains the ".env" file, which contains the secret access key of the AWS account used in this work. So, keep this repository private or remove this file. The file is required to create the binaries.

## Setup
We briefly discuss the setup of our system with respect to AWS. It is important to note that this description focuses primarily on running the system on the AWS account provided as part of this work. In order to run it on a separate AWS account, we provide a brief overview here and note that some hard coded parts in the source code must be modified. It contains a `.env` file that contains the AWS account information and the name of the S3 bucket. In addition, the instance IDs, AWS Lambda function name, and Redis authentication password are hard-coded in the source code and need to be changed when running on a separate AWS account.

- **EC2**

    All instances we use are `t2.micro` EC2 instances with 8 GiB EBS of type `gp2`. We use the instance IDs of the orchestrator and reverse proxy hard coded in the source code in `ip.go` as well as the instance ID used for the Redis instance in `ec2.go`. Below is a table of the inbound rules of the security groups of the respective instances:
    
    * **Orchestrator**:

    | IP version | Type | Protocol | Port range | Source | Description |
    | ------ | ------ |  ------ | ------ |  ------ | ------- |
    | IPv6 | Custom TCP | TCP | 6000 | ::/0 | Reverse Proxy Port Listener |
    | IPv4 | Custom TCP | TCP | 6000 | 0.0.0.0/0 | Reverse Proxy Port Listener |
    | -    | Custom TCP | TCP | 7000 | default VPC | Lambda Port Listener |
    | IPv4 | SSH | TCP | 22 | 0.0.0.0/0 | SSH |
    | IPv4 | Custom TCP | TCP | 3000 | 0.0.0.0/0 | Swagger Port Listener |


    * **Reverse Proxy**:
    
    | IP version | Type | Protocol | Port range | Source | Description |
    | ------ | ------ |  ------ | ------ |  ------ | ------- |
    | IPv6 | Custom TCP | TCP | 5000 | ::/0 | Orchestrator Port Listener |
    | IPv4 | Custom TCP | TCP | 5000 | 0.0.0.0/0 | Orchestrator Port Listener |
    | -    | Custom TCP | TCP | 8000 | default VPC | Lambda Port Listener |
    | IPv4 | SSH | TCP | 22 | 0.0.0.0/0 | SSH |
    | IPv4 | Custom TCP | TCP | 4000 | 0.0.0.0/0 | Swagger Port Listener |

    * **Redis instance**:
    
    | IP version | Type | Protocol | Port range | Source | Description |
    | ------ | ------ |  ------ | ------ |  ------ | ------- |
    | IPv4 | All TCP | TCP | 0-65535 | 0.0.0.0/0 | Redis Listener |

    The Redis instance must be created before the system is started, because we need to prepare the instance so that Redis runs as a daemon once the instance is started later. For this we refer to this [guide](https://medium.com/@calvin.hsieh/spin-up-redis-with-aws-ec2-e71911c55d61), we also enabled authentication on Redis with the password "cloud-testing" hardcoded in `client.go`.

- **S3**
    
    Our AWS S3 bucket is named `bunert-testbucket` and contains the object we used within the scope of our work. The bucket blocks all public access and the bucket policy is as follow:
    <details><summary>Bucket policy</summary>

        {
            "Version": "2012-10-17",
            "Id": "ExamplePolicy",
            "Statement": [
                {
                    "Sid": "ExampleStmt",
                    "Effect": "Allow",
                    "Principal": {
                        "AWS": "arn:aws:iam::339332244306:role/lambda-role"
                    },
                    "Action": "s3:GetObject",
                    "Resource": "arn:aws:s3:::bunert-testbucket/*"
                }
            ]
        }
    </details>

- **Lambda**

    To create the zip file for the AWS Lambda function, the `build_lambda.sh` script is used. It creates the binary and packs it into a zip file, which is then uploaded to the AWS Lambda function. The function name is `get-s3-object` and is hard coded in several places. The function is configured with 1024 MB of memory and a timeout of 4 minutes, and uses the Go 1.x runtime and x86_64 architecture. We have added the function to the default VPC according to our AWS account. The execution role (`lambda-role`) contains the two default policies (`AWSXRayDaemonWriteAccess` and `AWSLambdaBasicExecutionRole`) and two custom:
    <details><summary>Policies:</summary>
        
        {
            "Version": "2012-10-17",
            "Statement": [
                {
                    "Sid": "ExampleStmt",
                    "Action": [
                        "s3:GetObject"
                    ],
                    "Effect": "Allow",
                    "Resource": [
                        "arn:aws:s3:::bunert-testbucket/*"
                    ]
                }
            ]
        }

        {
            "Version": "2012-10-17",
            "Statement": [
                {
                    "Effect": "Allow",
                    "Action": [
                        "ec2:DescribeNetworkInterfaces",
                        "ec2:CreateNetworkInterface",
                        "ec2:DeleteNetworkInterface",
                        "ec2:DescribeInstances",
                        "ec2:AttachNetworkInterface"
                    ],
                    "Resource": "*"
                }
            ]
        }
    </details>


## Execution
The Orchestrator and the Reverse Proxy EC2 instances must be started to start our system. Once the instances are up and running, proceed:
- **Orchestrator**

    SSH onto the EC2 instance which runs the orchestrator. Generate the system parameter file (conf.yaml) and run the orchestrator binary.
    ```bash
    ssh -v -i [SSH key pair] ec2-user@[public IPv4 address]
    ./system_configurator -rate 3.0 -sensitivity 4
    ./system_orchestrator
    ```
- **Reverse Proxy**

    SSH onto the EC2 instance which runs the reverse proxy and execute the reverse proxy binary.
    ```bash
    ssh -v -i [SSH key pair] ec2-user@[public IPv4 address]
    ./system_gateway
    ```

## Evaluation
This part is executed from the local computer used for testing. The simulation results are moved to subdirectories to keep them available but to empty the workspace. Navigate to the \code\evaluation directory:
- **Generate/Derive Trace**

    Depending on the -t/--trace parameter, several additional parameters are required. It generates a tracefile in the trace directory used for the next step.
    ```bash
    python generate.py -t poisson --duration 30 --lambda 4
    ```
- **Simulate Trace**

    The same simulation script can be used to simulate the trace on our system (system argument) and the two comparison systems ElastiCache and S3 only. The name helps to provide an overview about the simulation results as it is only used for file naming. Creates an output file in the simulation directory and if running on our system an additional log file from the orchestrator.
    ```bash
    python simulate.py -i traces/[trace file] --system [system, EC, S3] --name [additional name for output file naming]
    ```

- **Evaluate Simulation**

    Reads the simulation files according to the argument and generates the cost overview textfile and the figures as presented in the report.
    ```bash
    python evaluate.py -i simulation/[simulation file]
    ```

### Additional scripts
We provide a rough overview about the remaining evaluation scripts:
- **Latency Measurements**

    Each simulation (`latency.py`) requires to copy the log output from the reverse proxy into the simulation folder which was then used by the evaluation script (`evaluate_latency.py`).
    ```bash
    python latency.py -i traces/5-latency-30.txt --system [system, EC, S3] --object [100B.txt, 1KB.txt, 1MB.txt]
    python evaluate_latency.py
    ```
- **Startup Latency**

    Each simulation (`startup.py`) uses a specific request pattern to trigger the desired state in our system. The log of the   
    orchestrator is then used to evaluate the measurements. For these measurements, some additional timestamps were added to the 
    orchestrator. Based on the timestamps and the request pattern, the evaluation script (`evaluation_startup.py`) can derive the 
    startup times.
    ```bash
    python startup.py --system [cold-start, warm-start, redis] --object [100B.txt, 1KB.txt, 1MB.txt]
    python evaluation_startup.py
    ```
