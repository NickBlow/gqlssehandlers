package eventstreams

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
)

var snsSvc *sns.SNS
var sqsSvc *sqs.SQS

func init() {
	sess := getSession()
	snsSvc = sns.New(sess)
	sqsSvc = sqs.New(sess)
}

// AWSEventStream is an event stream capable of processing events sent to SNS
type AWSEventStream struct{}

// returns a correctly configured AWS session with correct region etc.
func getSession() *session.Session {
	// Set some defaults if env vars not set
	region := os.Getenv("REGION")
	if region == "" {
		region = "eu-west-1"
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region)},
	))
	return sess
}

func getTaskID() string {
	metadataHost := os.Getenv("ECS_CONTAINER_METADATA_URI")
	if metadataHost == "" {
		fmt.Println("ECS_CONTAINER_METADATA_URI not set, returning a default task id.")
		return "GQL_SSE_HANDLERS"
	}
	resp, err := http.Get(fmt.Sprintf("%s/task", metadataHost))
	if err != nil {
		panic("Cannot access instance metadata")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic("Malformed instance metadata response")
	}
	data := struct {
		TaskARN string
	}{}
	err = json.Unmarshal(body, &data)
	splitTaskARN := strings.Split(data.TaskARN, "task/")
	if err != nil || len(splitTaskARN) < 2 {
		panic("Malformed instance metadata response")
	}
	taskID := splitTaskARN[1]
	return taskID
}

func setQueuePolicy(queueARN, queueURL, snsARN string) {
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Id": "RealtimeToSQSPOlicy",
		"Statement": [{
		   "Sid":"AllowSNSSend",
		   "Effect":"Allow",
		   "Principal":"*",
		   "Action":"sqs:SendMessage",
		   "Resource":"%s",
		   "Condition":{
			 "ArnEquals":{
			   "aws:SourceArn":"%s"
			 }
		   }
		}]
	  }
  `, queueARN, snsARN)
	_, err := sqsSvc.SetQueueAttributes(&sqs.SetQueueAttributesInput{
		Attributes: map[string]*string{"Policy": aws.String(policy)},
		QueueUrl:   aws.String(queueURL),
	})
	if err != nil {
		panic("Couldn't update queue policy")
	}
}

func createQueue(snsARN string) (string, string) {
	queuePrefix := "Subscription-Server-"
	queueName := fmt.Sprintf("%s%s", queuePrefix, getTaskID())
	output, err := sqsSvc.CreateQueue(&sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		fmt.Println(err)
		panic("Couldn't create Queue")
	}
	queueURL := *output.QueueUrl
	attributes, err := sqsSvc.GetQueueAttributes(&sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(queueURL),
		AttributeNames: []*string{aws.String("QueueArn")},
	})
	queueARN := attributes.Attributes["QueueArn"]
	if err != nil || queueARN == nil {
		fmt.Println(err)
		panic("Couldn't get Queue ARN")
	}
	setQueuePolicy(*queueARN, queueURL, snsARN)
	return *queueARN, queueURL
}

// StartListening sets up the infrastructure necessary to listen to events on an SQS queue in AWS
func (a *AWSEventStream) StartListening(eventChannel chan string, snsARN string) {
	queueARN, queueURL := createQueue(snsARN)
	fmt.Printf("Created SQS queue %s\n", queueARN)

	_, err := snsSvc.Subscribe(&sns.SubscribeInput{
		Endpoint:   aws.String(queueARN),
		Protocol:   aws.String("sqs"),
		TopicArn:   aws.String(snsARN),
		Attributes: map[string]*string{"RawMessageDelivery": aws.String("true")},
	})
	if err != nil {
		fmt.Println(err)
		panic("Couldn't set up subscription")
	}
	fmt.Println("Hooked up SNS and SQS")
	go beginProcessingQueue(eventChannel, queueURL)
}

func beginProcessingQueue(eventChannel chan string, queueURL string) {
	for {
		fmt.Println("Long polling for messages")
		output, err := sqsSvc.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(queueURL),
			MaxNumberOfMessages: aws.Int64(10),
			WaitTimeSeconds:     aws.Int64(15),
		})
		if err != nil {
			fmt.Println(err)
			fmt.Println("Error getting queue messages")
		}
		fmt.Printf("Got %d messages\n", len(output.Messages))
		processMessages(eventChannel, queueURL, output.Messages)
	}
}

func ackMessage(message *sqs.Message, queueURL string) {
	_, err := sqsSvc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: message.ReceiptHandle,
	})
	if err != nil {
		fmt.Println(err)
		fmt.Println("error deleting message, this isn't fatal as events are idempotent")
	}
}

func processMessages(eventChannel chan string, queueURL string, messages []*sqs.Message) {
	for _, msg := range messages {
		if msg.Body != nil {
			eventChannel <- *msg.Body
			ackMessage(msg, queueURL)
		}
	}
}
