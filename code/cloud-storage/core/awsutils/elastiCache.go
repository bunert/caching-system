package awsutils

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elasticache"
)

var (
	ElastiCache *elasticache.ElastiCache
	clusterName = "ElastiCacheRedis"

	ErrClusterAddr = errors.New("no address for ElastiCache available")
)

func CreateCluster() {
	input := &elasticache.CreateCacheClusterInput{
		AutoMinorVersionUpgrade: aws.Bool(true),
		CacheClusterId:          aws.String(clusterName),
		CacheNodeType:           aws.String("cache.t2.micro"),
		CacheSubnetGroupName:    aws.String("default"),
		Engine:                  aws.String("redis"),
		EngineVersion:           aws.String("6.2"),
		NumCacheNodes:           aws.Int64(1),
		SecurityGroupIds:        aws.StringSlice([]string{"sg-06dd017d7137a5c1c"}),
		Port:                    aws.Int64(6379),
		SnapshotRetentionLimit:  aws.Int64(0), // automatic backups are disabled for this cache cluster
	}

	result, err := ElastiCache.CreateCacheCluster(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elasticache.ErrCodeReplicationGroupNotFoundFault:
				log.Warnln(elasticache.ErrCodeReplicationGroupNotFoundFault, aerr.Error())
			case elasticache.ErrCodeInvalidReplicationGroupStateFault:
				log.Warnln(elasticache.ErrCodeInvalidReplicationGroupStateFault, aerr.Error())
			case elasticache.ErrCodeCacheClusterAlreadyExistsFault:
				log.Warnln(elasticache.ErrCodeCacheClusterAlreadyExistsFault, aerr.Error())
			case elasticache.ErrCodeInsufficientCacheClusterCapacityFault:
				log.Warnln(elasticache.ErrCodeInsufficientCacheClusterCapacityFault, aerr.Error())
			case elasticache.ErrCodeCacheSecurityGroupNotFoundFault:
				log.Warnln(elasticache.ErrCodeCacheSecurityGroupNotFoundFault, aerr.Error())
			case elasticache.ErrCodeCacheSubnetGroupNotFoundFault:
				log.Warnln(elasticache.ErrCodeCacheSubnetGroupNotFoundFault, aerr.Error())
			case elasticache.ErrCodeClusterQuotaForCustomerExceededFault:
				log.Warnln(elasticache.ErrCodeClusterQuotaForCustomerExceededFault, aerr.Error())
			case elasticache.ErrCodeNodeQuotaForClusterExceededFault:
				log.Warnln(elasticache.ErrCodeNodeQuotaForClusterExceededFault, aerr.Error())
			case elasticache.ErrCodeNodeQuotaForCustomerExceededFault:
				log.Warnln(elasticache.ErrCodeNodeQuotaForCustomerExceededFault, aerr.Error())
			case elasticache.ErrCodeCacheParameterGroupNotFoundFault:
				log.Warnln(elasticache.ErrCodeCacheParameterGroupNotFoundFault, aerr.Error())
			case elasticache.ErrCodeInvalidVPCNetworkStateFault:
				log.Warnln(elasticache.ErrCodeInvalidVPCNetworkStateFault, aerr.Error())
			case elasticache.ErrCodeTagQuotaPerResourceExceeded:
				log.Warnln(elasticache.ErrCodeTagQuotaPerResourceExceeded, aerr.Error())
			case elasticache.ErrCodeInvalidParameterValueException:
				log.Warnln(elasticache.ErrCodeInvalidParameterValueException, aerr.Error())
			case elasticache.ErrCodeInvalidParameterCombinationException:
				log.Warnln(elasticache.ErrCodeInvalidParameterCombinationException, aerr.Error())
			default:
				log.Warnln(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Warnln(err.Error())
		}
		return
	}
	log.Debugln(result)
}

func WaitForCluster() {

	input := &elasticache.DescribeCacheClustersInput{
		CacheClusterId:    aws.String(clusterName),
		ShowCacheNodeInfo: aws.Bool(false),
	}

	err := ElastiCache.WaitUntilCacheClusterAvailable(input)
	if err != nil {
		log.WithError(err).Error("WaitUntilCacheClusterAvailable failed")
		return
	}

}

func DeleteCluster() {
	input := &elasticache.DeleteCacheClusterInput{
		CacheClusterId: aws.String(clusterName),
	}

	result, err := ElastiCache.DeleteCacheCluster(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elasticache.ErrCodeCacheClusterNotFoundFault:
				log.Warn("ElastiCache Cluster not found, nothing to delete")
			case elasticache.ErrCodeInvalidCacheClusterStateFault:
				log.Warnln(elasticache.ErrCodeInvalidCacheClusterStateFault, aerr.Error())
			case elasticache.ErrCodeSnapshotAlreadyExistsFault:
				log.Warnln(elasticache.ErrCodeSnapshotAlreadyExistsFault, aerr.Error())
			case elasticache.ErrCodeSnapshotFeatureNotSupportedFault:
				log.Warnln(elasticache.ErrCodeSnapshotFeatureNotSupportedFault, aerr.Error())
			case elasticache.ErrCodeSnapshotQuotaExceededFault:
				log.Warnln(elasticache.ErrCodeSnapshotQuotaExceededFault, aerr.Error())
			case elasticache.ErrCodeInvalidParameterValueException:
				log.Warnln(elasticache.ErrCodeInvalidParameterValueException, aerr.Error())
			case elasticache.ErrCodeInvalidParameterCombinationException:
				log.Warnln(elasticache.ErrCodeInvalidParameterCombinationException, aerr.Error())
			default:
				log.Warnln(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Warnln(err.Error())
		}
		return
	}

	log.Debugln(result)
}

func GetClusterAddr() (addr string, err error) {
	input := &elasticache.DescribeCacheClustersInput{
		CacheClusterId:    aws.String(clusterName),
		ShowCacheNodeInfo: aws.Bool(true),
	}

	result, err := ElastiCache.DescribeCacheClusters(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case elasticache.ErrCodeCacheClusterNotFoundFault:
				log.Warnln(elasticache.ErrCodeCacheClusterNotFoundFault, aerr.Error())
			case elasticache.ErrCodeInvalidParameterValueException:
				log.Warnln(elasticache.ErrCodeInvalidParameterValueException, aerr.Error())
			case elasticache.ErrCodeInvalidParameterCombinationException:
				log.Warnln(elasticache.ErrCodeInvalidParameterCombinationException, aerr.Error())
			default:
				log.Warnln(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Warnln(err.Error())
		}
		return "", ErrClusterAddr
	}

	log.Debugln(result)
	return *result.CacheClusters[0].CacheNodes[0].Endpoint.Address, nil
}
