package queue

import (
	"qp"
	"utils"

	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type Sqs struct {
	configuration sqsConfiguration
	queue         *sqs.SQS
	queueUrl      *string
	logger        *log.Entry
}

type sqsConfiguration struct {
	QueueName       string
	WaitTimeSeconds int
	AwsRegion       string
	AwsProfile      string
}

func (q *Sqs) Configure(configuration map[string]interface{}) error {
	q.logger = log.WithFields(log.Fields{
		"type":  "queue",
		"queue": "Sqs",
	})
	q.logger.Debug("Reading configuration")
	q.configuration.WaitTimeSeconds = 20 //Defaults
	utils.FillStruct(configuration, &q.configuration)

	if q.configuration.AwsProfile == "" {
		return errors.New("You need to provide AwsProfile for Sqs queue")
	}

	if q.configuration.AwsRegion == "" {
		return errors.New("You need to provide AwsRegion for Sqs queue")
	}

	q.queue = sqs.New(session.New(), &aws.Config{
		Region:      aws.String(q.configuration.AwsRegion),
		Credentials: credentials.NewSharedCredentials("", q.configuration.AwsProfile),
	})

	params := &sqs.GetQueueUrlInput{
		QueueName: aws.String(q.configuration.QueueName),
	}

	resp, err := q.queue.GetQueueUrl(params)

	if err != nil {
		q.logger.WithError(err).Error("Error on GetQueueUrl")
		return err
	}

	q.queueUrl = resp.QueueUrl

	q.logger.WithField("configuration", q.configuration).Info("Configuration loaded")

	return nil
}

func (q *Sqs) GetName() string {
	return "Sqs"
}

func (q *Sqs) Consume() (qp.IMessage, error) {
	q.logger.Debug("Message consume")
	for {
		params := &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(*q.queueUrl),
			MaxNumberOfMessages: aws.Int64(1),
			WaitTimeSeconds:     aws.Int64(int64(q.configuration.WaitTimeSeconds)),
		}

		resp, err := q.queue.ReceiveMessage(params)

		if err != nil {
			return nil, err
		}

		if len(resp.Messages) != 0 {
			return &qp.Message{
				Id:   *resp.Messages[0].ReceiptHandle,
				Body: *resp.Messages[0].Body,
			}, nil
		}
	}
}

func (q *Sqs) Ack(message qp.IMessage) error {
	q.logger.WithField("message", message).Debug("Message acknowledged")
	params := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(*q.queueUrl),
		ReceiptHandle: aws.String((message.GetId()).(string)),
	}

	_, err := q.queue.DeleteMessage(params)

	if err != nil {
		return err
	}

	return nil
}

func (q *Sqs) Reject(message qp.IMessage) error {
	q.logger.WithField("message", message).Debug("Message rejected")
	// Do nothing. Aws SQS will take care of not acknowledged messages
	// and will put them into dead letter queue for us
	// We can integrate here some RejectPolicy, but do not want to spent time on this now
	return nil
}

func (q *Sqs) GetNumberOfMessages() (int, error) {
	return 9999999, nil
}
