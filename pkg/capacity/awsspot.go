package capacity

import (
	"context"
	"time"

	"waverless/pkg/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// SpotInfo Spot 实例信息
type SpotInfo struct {
	SpecName     string
	InstanceType string
	Score        int     // 1-10, 越高越容易获取
	Price        float64 // USD/hour
}

// NodePoolInstanceTypeFetcher 从 NodePool 获取 instance types 的接口
type NodePoolInstanceTypeFetcher interface {
	GetInstanceTypesFromNodePool(ctx context.Context, nodePoolName string) ([]string, error)
}

// AWSSpotChecker AWS Spot 容量和价格检查器
type AWSSpotChecker struct {
	ec2Client         *ec2.Client
	region            string
	specToInstance    map[string]string   // spec name -> instance type (from config)
	specToNodePool    map[string]string   // spec name -> nodepool name
	nodePoolFetcher   NodePoolInstanceTypeFetcher
	instanceTypeCache map[string]string   // spec name -> instance type (cached from nodepool)
}

func NewAWSSpotChecker(ec2Client *ec2.Client, region string, specToInstance map[string]string) *AWSSpotChecker {
	return &AWSSpotChecker{
		ec2Client:         ec2Client,
		region:            region,
		specToInstance:    specToInstance,
		specToNodePool:    make(map[string]string),
		instanceTypeCache: make(map[string]string),
	}
}

// SetNodePoolFetcher 设置 NodePool 获取器
func (c *AWSSpotChecker) SetNodePoolFetcher(fetcher NodePoolInstanceTypeFetcher, specToNodePool map[string]string) {
	c.nodePoolFetcher = fetcher
	c.specToNodePool = specToNodePool
}

// getInstanceType 获取 spec 对应的 instance type
func (c *AWSSpotChecker) getInstanceType(ctx context.Context, specName string) string {
	// 1. 先从配置获取
	if instanceType, ok := c.specToInstance[specName]; ok {
		return instanceType
	}

	// 2. 从缓存获取
	if instanceType, ok := c.instanceTypeCache[specName]; ok {
		return instanceType
	}

	// 3. 从 NodePool 获取
	if c.nodePoolFetcher != nil {
		if nodePool, ok := c.specToNodePool[specName]; ok {
			instanceTypes, err := c.nodePoolFetcher.GetInstanceTypesFromNodePool(ctx, nodePool)
			if err != nil {
				logger.WarnCtx(ctx, "Failed to get instance types from nodepool %s: %v", nodePool, err)
			} else if len(instanceTypes) > 0 {
				// 取第一个作为主要类型
				c.instanceTypeCache[specName] = instanceTypes[0]
				logger.InfoCtx(ctx, "Got instance type %s from nodepool %s for spec %s", instanceTypes[0], nodePool, specName)
				return instanceTypes[0]
			}
		}
	}

	return ""
}

// CheckSpotScore 检查 Spot 容量评分
func (c *AWSSpotChecker) CheckSpotScore(ctx context.Context, specName string) (*SpotInfo, error) {
	instanceType := c.getInstanceType(ctx, specName)
	if instanceType == "" {
		return nil, nil
	}

	// 获取 Spot Placement Score
	scoreResp, err := c.ec2Client.GetSpotPlacementScores(ctx, &ec2.GetSpotPlacementScoresInput{
		InstanceTypes:           []string{instanceType},
		TargetCapacity:          aws.Int32(1),
		SingleAvailabilityZone:  aws.Bool(false),
		RegionNames:             []string{c.region},
		TargetCapacityUnitType:  types.TargetCapacityUnitTypeUnits,
	})
	if err != nil {
		logger.WarnCtx(ctx, "Failed to get spot placement score for %s: %v", instanceType, err)
		return nil, err
	}

	score := 5 // 默认中等
	if len(scoreResp.SpotPlacementScores) > 0 {
		if scoreResp.SpotPlacementScores[0].Score != nil {
			score = int(*scoreResp.SpotPlacementScores[0].Score)
		}
	}

	// 获取 Spot 价格
	price, err := c.getSpotPrice(ctx, instanceType)
	if err != nil {
		logger.WarnCtx(ctx, "Failed to get spot price for %s: %v", instanceType, err)
		// 继续，价格不是必须的
	}

	return &SpotInfo{
		SpecName:     specName,
		InstanceType: instanceType,
		Score:        score,
		Price:        price,
	}, nil
}

// getSpotPrice 获取当前 Spot 价格
func (c *AWSSpotChecker) getSpotPrice(ctx context.Context, instanceType string) (float64, error) {
	resp, err := c.ec2Client.DescribeSpotPriceHistory(ctx, &ec2.DescribeSpotPriceHistoryInput{
		InstanceTypes: []types.InstanceType{types.InstanceType(instanceType)},
		ProductDescriptions: []string{"Linux/UNIX"},
		StartTime:     aws.Time(time.Now().Add(-1 * time.Hour)),
		MaxResults:    aws.Int32(1),
	})
	if err != nil {
		return 0, err
	}

	if len(resp.SpotPriceHistory) > 0 && resp.SpotPriceHistory[0].SpotPrice != nil {
		var price float64
		parseFloatSimple(*resp.SpotPriceHistory[0].SpotPrice, &price)
		return price, nil
	}

	return 0, nil
}

// CheckAllSpots 检查所有 spec 的 Spot 信息
func (c *AWSSpotChecker) CheckAllSpots(ctx context.Context) ([]SpotInfo, error) {
	var results []SpotInfo

	// 收集所有需要检查的 spec
	specsToCheck := make(map[string]bool)
	for specName := range c.specToInstance {
		specsToCheck[specName] = true
	}
	for specName := range c.specToNodePool {
		specsToCheck[specName] = true
	}

	for specName := range specsToCheck {
		info, err := c.CheckSpotScore(ctx, specName)
		if err != nil {
			logger.WarnCtx(ctx, "Failed to check spot for %s: %v", specName, err)
			continue
		}
		if info != nil {
			results = append(results, *info)
		}
	}

	return results, nil
}

func parseFloatSimple(s string, result *float64) {
	var val float64
	var decimal float64 = 0
	var inDecimal bool
	var decimalPlace float64 = 0.1

	for _, c := range s {
		if c >= '0' && c <= '9' {
			if inDecimal {
				decimal += float64(c-'0') * decimalPlace
				decimalPlace *= 0.1
			} else {
				val = val*10 + float64(c-'0')
			}
		} else if c == '.' {
			inDecimal = true
		}
	}
	*result = val + decimal
}
