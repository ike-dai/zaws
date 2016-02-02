package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/AlekSi/zabbix-sender"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"net"
	"os"
	"strconv"
	"time"
)

// Declare Struct
type LldJson struct {
	Data []Data `json:"data"`
}

type Data struct {
	MetricName          string `json:"{#METRIC.NAME},omitempty"`
	MetricUnit          string `json:"{#METRIC.UNIT},omitempty"`
	MetricNamespace     string `json:"{#METRIC.NAMESPACE},omitempty"`
	InstanceName        string `json:"{#INSTANCE.NAME},omitempty"`
	InstanceType        string `json:"{#INSTANCE.TYPE},omitempty"`
	InstanceId          string `json:"{#INSTANCE.ID},omitempty"`
	InstancePrivateAddr string `json:"{#INSTANCE.PRIVATE.ADDR},omitempty"`
}

// Common util
func parse_option() {
	flag.StringVar(&region, "region", "ap-northeast-1", "Set AWS region")
	flag.StringVar(&region, "r", "ap-northeast-1", "Set AWS region")
	flag.StringVar(&access_key_id, "key", os.Getenv("AWS_ACCESS_KEY_ID"), "Set AWS API Access key id")
	flag.StringVar(&access_key_id, "k", os.Getenv("AWS_ACCESS_KEY_ID"), "Set AWS API Access key id")
	flag.StringVar(&secret_key_id, "secret", os.Getenv("AWS_SECRET_ACCESS_KEY"), "Set AWS API Secret key id")
	flag.StringVar(&secret_key_id, "s", os.Getenv("AWS_SECRET_ACCESS_KEY"), "Set AWS API Secret key id")
	flag.StringVar(&target_id, "id", "", "Set target object id")
	flag.StringVar(&target_id, "i", "", "Set target object id")
	flag.StringVar(&metric_name, "metric", "", "Set metric name")
	flag.StringVar(&metric_name, "m", "", "Set metric name")
	flag.StringVar(&zabbix_host, "host", "localhost", "Set zabbix host name")
	flag.StringVar(&zabbix_host, "h", "localhost", "Set zabbix host name")
	flag.StringVar(&zabbix_port, "port", "10051", "Set zabbix host port")
	flag.StringVar(&zabbix_port, "p", "10051", "Set zabbix host port")
	flag.Parse()
	if access_key_id == "" || secret_key_id == "" {
		fmt.Println("[ERROR]: Please set key information")
		usage()
	}

}

func new_session(region, access_key_id, secret_key_id string) *session.Session {
	sess := session.New(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(access_key_id, secret_key_id, ""),
	})
	return sess
}

func usage() {
	fmt.Println("Usage: zaws service method [target] [-region|-r] [-key|-k] [-secret|-s] [-id|-i] [-metric|-m] [-host|h] [-port|p]")
	os.Exit(1)
}

func convert_to_lldjson_string(data []Data) string {
	lld_json := LldJson{data}
	convert_json, _ := json.Marshal(lld_json)
	return string(convert_json)
}

// Access AWS API
func get_metric_list(sess *session.Session) []*cloudwatch.Metric {
	var metrics []*cloudwatch.Metric
	svc := cloudwatch.New(sess)
	params := &cloudwatch.ListMetricsInput{
		Dimensions: []*cloudwatch.DimensionFilter{
			{
				Name:  aws.String("InstanceId"),
				Value: aws.String(target_id),
			},
		},
	}
	resp, err := svc.ListMetrics(params)
	if err != nil {
		fmt.Println(err.Error())
		return metrics
	}
	metrics = resp.Metrics
	return metrics
}

func get_metric_stats(sess *session.Session, metric_name string, metric_namespace string) []*cloudwatch.Datapoint {

	var datapoints []*cloudwatch.Datapoint
	svc := cloudwatch.New(sess)
	t := time.Now()
	input := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String(metric_namespace),
		Statistics: []*string{aws.String("Average")},
		EndTime:    aws.Time(t),
		Period:     aws.Int64(300),
		StartTime:  aws.Time(t.Add(time.Duration(-10) * time.Minute)),
		MetricName: aws.String(metric_name),
		Dimensions: []*cloudwatch.Dimension{
			{
				Name:  aws.String("InstanceId"),
				Value: aws.String(target_id),
			},
		},
	}
	value, err := svc.GetMetricStatistics(input)
	if err != nil {
		fmt.Println(err.Error())
		return datapoints
	}
	datapoints = value.Datapoints
	return datapoints
}

func get_ec2_list(sess *session.Session) []*ec2.Instance {
	var instances []*ec2.Instance
	svc := ec2.New(sess)
	resp, err := svc.DescribeInstances(nil)

	if err != nil {
		fmt.Println("Error")
		os.Exit(1)
	}
	for _, reservation := range resp.Reservations {
		instances = reservation.Instances
		return instances
	}
	return instances
}

// zaws method
func show_ec2_list(sess *session.Session) {
	list := make([]Data, 0)
	instances := get_ec2_list(sess)
	for _, instance := range instances {
		for _, tag := range instance.Tags {
			if *tag.Key == "Name" {
				var private_addr string
				if instance.PrivateIpAddress != nil {
					private_addr = *instance.PrivateIpAddress
				}
				data := Data{InstanceName: *tag.Value, InstanceType: *instance.InstanceType, InstanceId: *instance.InstanceId, InstancePrivateAddr: private_addr}
				list = append(list, data)
			}
		}
	}
	fmt.Printf(convert_to_lldjson_string(list))
}

func show_cloudwatch_metrics_list(sess *session.Session) {
	list := make([]Data, 0)
	metrics := get_metric_list(sess)
	for _, metric := range metrics {
		datapoints := get_metric_stats(sess, *metric.MetricName, *metric.Namespace)
		data := Data{MetricName: *metric.MetricName, MetricNamespace: *metric.Namespace}
		if len(datapoints) > 0 {
			data.MetricUnit = *datapoints[0].Unit
		}
		list = append(list, data)
	}

	fmt.Printf(convert_to_lldjson_string(list))
}

func send_metric_stats(sess *session.Session) {
	var send_data []zabbix_sender.DataItem

	metrics := get_metric_list(sess)
	for _, metric := range metrics {
		datapoints := get_metric_stats(sess, *metric.MetricName, *metric.Namespace)

		if len(datapoints) > 0 {
			data_time := *datapoints[0].Timestamp
			send_data = append(send_data, zabbix_sender.DataItem{Hostname: target_id, Key: "cloudwatch.metric[" + *metric.MetricName + "]", Value: strconv.FormatFloat(*datapoints[0].Average, 'f', 4, 64), Timestamp: data_time.Unix()})
		}
	}
	addr, _ := net.ResolveTCPAddr("tcp", zabbix_host+":"+zabbix_port)
	res, err := zabbix_sender.Send(addr, send_data)
	if err != nil {
		fmt.Printf("[ERROR]: zabbix sender error!: %s", err)
		os.Exit(1)
	}
	fmt.Printf("[INFO]: Successful sending data to Zabbix: resp", res)
}

var region, access_key_id, secret_key_id, target_id, metric_name, zabbix_host, zabbix_port string

func main() {
	if len(os.Args) < 3 {
		usage()
	}
	switch os.Args[1] {
	case "ec2":
		switch os.Args[2] {
		case "list":
			os.Args = os.Args[2:]
			parse_option()
			show_ec2_list(new_session(region, access_key_id, secret_key_id))
		default:
			usage()
		}
	case "cloudwatch":
		switch os.Args[2] {
		case "list":
			if len(os.Args) < 4 {
				usage()
			}
			switch os.Args[3] {
			case "ec2":
				os.Args = os.Args[3:]
				parse_option()
				show_cloudwatch_metrics_list(new_session(region, access_key_id, secret_key_id))
			case "rds":
			case "elb":
			default:
				usage()
			}
		case "stats":
			os.Args = os.Args[2:]
			parse_option()
			send_metric_stats(new_session(region, access_key_id, secret_key_id))
		default:
			usage()
		}

	default:
		usage()
	}
	os.Exit(0)
}
