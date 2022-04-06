#!/bin/bash
script="build_orchestrator"
#Declare the number of mandatory args
margs=0

# Common functions - BEGIN
function example {
    echo -e "example: $script -ip VAL"
}

function usage {
    echo -e "usage: $script MANDATORY [OPTION]\n"
}

function help {
  usage
    echo -e "OPTION:"
    echo -e "  -ip, --ipaddr  VAL      IP address of EC2 instance to transfer the executables"
    echo -e "  -h,  --help             Prints this help\n"
  example
}

# Ensures that the number of passed args are at least equals
# to the declared number of mandatory args.
# It also handles the special case of the -h or --help arg.
function margs_precheck {
	if [ $2 ] && [ $1 -lt $margs ]; then
		if [ $2 == "--help" ] || [ $2 == "-h" ]; then
			help
			exit
		else
	    	usage
			example
	    	exit 1 # error
		fi
	fi
}

# Ensures that all the mandatory args are not empty
function margs_check {
	if [ $# -lt $margs ]; then
	    usage
	  	example
	    exit 1 # error
	fi
}
# Custom functions - BEGIN
# Put here your custom functions
# Custom functions - END

# Main
margs_precheck $# $1

IP=

# Args while-loop
while [ "$1" != "" ];
do
   case $1 in
   -ip )  shift
                          IP=$1
                		  ;;
   -h   | --help )        help
                          exit
                          ;;
   *)                     
                          echo "$script: illegal option $1"
                          usage
						  example
						  exit 1 # error
                          ;;
    esac
    shift
done

# Pass here your mandatory args for check
# margs_check $IP

# Your stuff goes here
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o system_gateway ./cmd/gateway/main.go

if [ ! -z "$IP" ]
then
   scp -i ../aws-keypairs/ec2-gateway.pem system_gateway ec2-user@$IP:~
else
   echo "no IP specified, so don't transfer executables"
fi
